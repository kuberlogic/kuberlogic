package cli

import (
	"archive/zip"
	"context"
	_ "embed"
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"io/fs"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

// embed kustomize files into cli binary
//go:embed kustomize-configs.zip
var klConfigZipData []byte

var (
	kubectlBin   = "kubectl"
	kustomizeBin = "kustomize"
)

func makeInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Installs KuberLogic to Kubernetes cluster",
		RunE:  runInstall(),
	}

	_ = cmd.PersistentFlags().String("docker-compose", "", "Path to application docker-compose.yml")
	_ = cmd.PersistentFlags().Bool("backups-enabled", false, "Enable backup/restore support")
	_ = cmd.PersistentFlags().Bool("backups-snapshots-enabled", false, "Enable volume snapshot backups (Must be supported by Velero provider plugin)")
	_ = cmd.PersistentFlags().String("tls-key", "", "Path to TLS key to use for provisioned applications")
	_ = cmd.PersistentFlags().String("tls-crt", "", "Path to TLS certificate to use for provisioned applications")
	return cmd
}

// runInstall function prepares configs and installs KuberLogic by calling kubectl and kustomize binaries
// it then uses client-go to get some config values and viper to write config file to disk
func runInstall() func(command *cobra.Command, args []string) error {
	return func(command *cobra.Command, args []string) error {
		command.Println("Preparing KuberLogic configs")
		tmpdir, err := os.MkdirTemp("", "kuberlogic-install")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmpdir)

		kustomizeRootDir, err := unzipConfigs(klConfigZipData, tmpdir)
		if err != nil {
			return errors.Wrap(err, "failed to unpack KuberLogic configs")
		}

		// cache config files passed by flags
		configBaseDir := filepath.Dir(configFile)
		cacheDir := filepath.Join(configBaseDir, "cache", "config")
		if err := os.MkdirAll(cacheDir, os.ModePerm); err != nil {
			return errors.Wrap(err, "error creating config directory")
		}

		// handle cmd flags
		if value, err := getString(command, "docker-compose"); err != nil {
			return err
		} else if value != nil {
			if err := cacheConfigFile(*value, filepath.Join(cacheDir, "manager/docker-compose.yaml")); err != nil {
				return err
			}
		}

		if backupsEnabled, err := getBool(command, "backups-enabled"); err != nil {
			return err
		} else if backupsEnabled != nil {
			backupProperties := fmt.Sprintf("backups_enabled=%t\n", *backupsEnabled)
			if snapshotsEnabled, err := getBool(command, "backups-snapshots-enabled"); err != nil {
				return err
			} else if snapshotsEnabled != nil {
				backupProperties = backupProperties + fmt.Sprintf("backups_snapshots_enabled=%t", *snapshotsEnabled)
			}

			if err := cacheConfigFileWithContent(backupProperties, filepath.Join(cacheDir, "/manager/backups.properties")); err != nil {
				return err
			}
		}

		if tlsCertPath, err := getString(command, "tls-crt"); err != nil {
			return err
		} else if tlsCertPath != nil {
			tlsKeyPath, err := getString(command, "tls-key")
			if err != nil {
				return err
			}
			if tlsKeyPath == nil {
				return errors.New("tls certificate is set but tls key is missing")
			}

			if err := cacheConfigFile(*tlsCertPath, filepath.Join(cacheDir, "/certificates/tls.crt")); err != nil {
				return errors.Wrap(err, "failed to cache tls.crt")
			}
			if err := cacheConfigFile(*tlsKeyPath, filepath.Join(cacheDir, "/certificates/tls.key")); err != nil {
				return errors.Wrap(err, "failed to cache tls.key")
			}
		}

		if token := viper.GetString(tokenFlag); token != "" {
			apiserverProperties := fmt.Sprintf("token=%s", token)
			if err := cacheConfigFileWithContent(apiserverProperties, filepath.Join(cacheDir, "manager/apiserver.properties")); err != nil {
				return err
			}
		}

		// copy cached configuration files to kustomize configs
		if err := useCachedConfigFiles(cacheDir, kustomizeRootDir, command.Printf); err != nil {
			return errors.Wrap(err, "failed to restore cached configs")
		}

		// check if kubernetes is available
		if out, err := exec.Command("kubectl", "cluster-info").Output(); err != nil {
			command.Println("Kubernetes is not available via kubectl")
			command.Println(string(out))
			os.Exit(1)
		}

		// run kustomize via exec and apply manifests via kubectl
		command.Println("Installing cert-manager")
		cmd := fmt.Sprintf("%s build %s/cert-manager | %s apply -f -", kustomizeBin, kustomizeRootDir, kubectlBin)
		out, err := exec.Command("sh", "-c", cmd).Output()
		fmt.Println(string(out))
		if err != nil {
			return err
		}
		fmt.Println("Waiting for cert-manager to be ready")
		cmd = fmt.Sprintf("%s -n cert-manager wait --for=condition=ready pods --all --timeout=300s", kubectlBin)
		if out, err := exec.Command("sh", "-c", cmd).Output(); err != nil {
			fmt.Println(string(out))
			return errors.New("Failed installing cert-manager")
		}

		command.Println("Installing KuberLogic")
		cmd = fmt.Sprintf("%s build %s/default | %s apply -f -", kustomizeBin, kustomizeRootDir, kubectlBin)
		fmt.Println(string(out))
		out, err = exec.Command("sh", "-c", cmd).Output()
		if err != nil {
			return err
		}
		cmd = fmt.Sprintf("%s -n kuberlogic wait --for=condition=ready pods --all --timeout=300s", kubectlBin)
		if out, err := exec.Command("sh", "-c", cmd).Output(); err != nil {
			fmt.Println(string(out))
			return errors.New("Failed installing kuberlogic")
		}

		command.Println("Updating KuberLogic config file at " + configFile)

		k8scfg, err := clientcmd.BuildConfigFromFlags("", homedir.HomeDir()+"/.kube/config")
		if err != nil {
			return errors.Wrap(err, "failed to build Kubernetes client config")
		}
		k8sclient, err := kubernetes.NewForConfig(k8scfg)
		if err != nil {
			return errors.Wrap(err, "failed to build Kubernetes client")
		}
		endpoint, err := getKuberlogicEndpoint(k8sclient)
		if err != nil {
			return errors.Wrap(err, "failed to get KuberLogic api server host")
		}
		viper.Set(apiHostFlag, endpoint)

		err = viper.WriteConfig()
		if errors.Is(err, os.ErrNotExist) {
			err = viper.WriteConfigAs(configFile)
		}
		return errors.Wrap(err, "failed to write KuberLogic config")
	}
}

func unzipConfigs(zipData []byte, dir string) (string, error) {
	zipFname := "kustomization.zip"
	// save embedded config to temp dir
	if err := os.WriteFile(dir+"/"+zipFname, zipData, 0644); err != nil {
		return "", errors.Wrap(err, "failed to save archived config files")
	}

	// unzip kustomize.zip
	reader, err := zip.OpenReader(dir + "/" + zipFname)
	if err != nil {
		return "", errors.Wrap(err, "failed to unzip config files")
	}
	defer reader.Close()

	var filenames []string
	for _, f := range reader.File {
		fp := filepath.Join(dir, f.Name)
		filenames = append(filenames, fp)
		if f.FileInfo().IsDir() {
			err := os.MkdirAll(fp, os.ModePerm)
			if err != nil {
				return "", errors.Wrap(err, "failed to create config dir "+fp)
			}
			continue
		}
		err := os.MkdirAll(filepath.Dir(fp), os.ModePerm)
		if err != nil {
			return "", errors.Wrap(err, "failed to create config dir "+fp)
		}
		file, err := os.OpenFile(fp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return "", errors.Wrap(err, "failed to create config file "+fp)
		}

		rc, err := f.Open()
		if err != nil {
			return "", errors.Wrap(err, "failed to open config file for writing "+fp)
		}
		if _, err := io.Copy(file, rc); err != nil {
			return "", errors.Wrap(err, "failed to write config data to file "+fp)
		}

		_ = file.Close()
		_ = rc.Close()
	}
	return dir + "/config", nil
}

func useCachedConfigFiles(configCacheDir, configDir string, printf func(f string, i ...interface{})) error {
	return filepath.Walk(configCacheDir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		relative, err := filepath.Rel(configCacheDir, path)
		if err != nil {
			return errors.Wrap(err, "failed to get relative path for "+path)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return errors.Wrap(err, "failed to get content of "+path)
		}

		target := filepath.Join(configDir, relative)
		if err := os.WriteFile(target, data, os.ModePerm); err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to restore %s at %s", path, target))
		}

		printf("Restored %s from cache\n", relative)
		return nil
	})
}

func cacheConfigFile(src, name string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return errors.Wrap(err, "failed to read "+src)
	}
	return os.WriteFile(name, data, os.ModePerm)
}

func cacheConfigFileWithContent(data string, name string) error {
	if dir := filepath.Dir(name); dir != "." {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}
	return os.WriteFile(name, []byte(data), os.ModePerm)
}

func getKuberlogicEndpoint(c *kubernetes.Clientset) (string, error) {
	var endpoint string
	svc, err := c.CoreV1().Services("kuberlogic").Get(context.TODO(), "kls-api-server", metav1.GetOptions{})
	if err != nil {
		return "", errors.Wrap(err, "failed to get KuberLogic Service")
	}
	port := strconv.Itoa(int(svc.Spec.Ports[0].Port))

	if extIPs := svc.Spec.ExternalIPs; len(extIPs) != 0 {
		endpoint = extIPs[0]
	}

	if lbIPs := svc.Status.LoadBalancer.Ingress; len(lbIPs) != 0 {
		endpoint = lbIPs[0].IP
	}
	if svc.Spec.ExternalName != "" {
		endpoint = svc.Spec.ExternalName
	}

	if endpoint != "" {
		return endpoint + ":" + port, nil
	}

	nodePort := strconv.Itoa(int(svc.Spec.Ports[0].NodePort))
	nodes, err := c.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return "", errors.Wrap(err, "failed to list Kubernetes nodes")
	}
	if len(nodes.Items) < 1 {
		return "", errors.New("no nodes found")
	}

	priorities := map[corev1.NodeAddressType]int{
		corev1.NodeExternalIP:  10,
		corev1.NodeExternalDNS: 5,
		corev1.NodeInternalIP:  1,
	}
	var endpointPri int

	for _, addr := range nodes.Items[0].Status.Addresses {
		if endpointPri < priorities[addr.Type] {
			endpoint = addr.Address + ":" + nodePort
			endpointPri = priorities[addr.Type]
		}
	}
	return endpoint, nil
}
