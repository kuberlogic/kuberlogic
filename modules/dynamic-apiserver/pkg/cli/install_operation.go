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
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

// embed kustomize files into cli binary
//go:embed kustomize-configs.zip
var klConfigZipData []byte

var (
	kubectlBin   = "kubectl"
	kustomizeBin = "kustomize"

	errTokenEmpty         = errors.New("token can't be empty")
	errChargebeeKeyNotSet = errors.New("chargebee key can't be empty")
)

const (
	klSentryDSN = "https://b16abaff497941468fdf21aff686ff52@kl.sentry.cloudlinux.com/9"
)

const (
	installIngressClassName             = "ingress_class"
	installStorageClassName             = "storage_class"
	installDockerComposeParam           = "docker_compose"
	installBackupsEnabledParam          = "backups_enabled"
	installBackupsSnapshotsEnabledParam = "backups_snapshots_enabled"
	installTLSKeyParam                  = "tls_key"
	installTLSCrtParam                  = "tls_crt"
	installChargebeeSiteParam           = "chargebee_site"
	installChargebeeKeyParam            = "chargebee_key"
	installKuberlogicDomainParam        = "kuberlogic_domain"
	installReportErrors                 = "report_errors"
	installSentryDSNParam               = "sentry_dsn"
)

func makeInstallCmd(k8sclient kubernetes.Interface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Installs KuberLogic to Kubernetes cluster",
		RunE:  runInstall(k8sclient),
	}

	_ = cmd.PersistentFlags().Bool("non-interactive", false, "Do not enter interactive mode")
	_ = cmd.PersistentFlags().String(installIngressClassName, "", "Choose Kubernetes Ingress controller that will be used for application")
	_ = cmd.PersistentFlags().String(installStorageClassName, "", "Choose Kubernetes storageClass that will be used for application instances")
	_ = cmd.PersistentFlags().String(installDockerComposeParam, "", "Specify path to your docker-compose file with the application you want to provide as SaaS. You can skip this step by pressing 'Enter', then the sample application will be used")
	_ = cmd.PersistentFlags().Bool(installBackupsEnabledParam, false, "Enable backup/restore support? For more information, read https://kuberlogic.com/docs/configuring/backups. Type 'N' if you have not set up integration with Velero to support backup/restore capabilities, otherwise type 'y'")
	_ = cmd.PersistentFlags().Bool(installBackupsSnapshotsEnabledParam, false, "Enable volume snapshot backups (Must be supported by the Velero provider plugin)")
	_ = cmd.PersistentFlags().String(installTLSKeyParam, "", "Specify path to the TLS certificate. It is assumed that the TLS certificate will be a wildcard certificate. All KuberLogic managed applications share the same certificate by sharing the same ingress controller. You can skip this step by pressing 'Enter', In this case, a self-signed (demo) certificate will be used")
	_ = cmd.PersistentFlags().String(installTLSCrtParam, "", "Specify path to TLS key to use for provisioned applications")
	_ = cmd.PersistentFlags().String(installChargebeeSiteParam, "", "Specify ChargeBee site name. For more information, read https://kuberlogic.com/docs/configuring/billing. You can skip this step by pressing 'Enter', and set up the integration later")
	_ = cmd.PersistentFlags().String(installChargebeeKeyParam, "", "Specify ChargeBee API-key. For more information, read https://apidocs.chargebee.com/docs/api/?prod_cat_ver=2. API Keys are used to authenticate KuberLogic and control its access to the Chargebee API")
	_ = cmd.PersistentFlags().String(installKuberlogicDomainParam, "example.com", "Specify \"KuberLogic default domain\". This configuration parameter is used by KuberLogic to generate subdomains for the application instances when they are provisioned. (e.g. instance1.defaultdomain.com)")
	_ = cmd.PersistentFlags().Bool(installReportErrors, false, "Report errors to KuberLogic? Please type 'Y' if you want to help us improve KuberLogic, otherwise, type 'n'. Error reports will be generated and sent automatically, these reports contain only information about the errors and do not contain any user data. Let us receive errors at least from your test environments")
	_ = cmd.PersistentFlags().String(installSentryDSNParam, "", "Specify Sentry Data Source Name (DSN). For more information, read https://docs.sentry.io/product/sentry-basics/dsn-explainer/. (KuberLogic team will not be notified in case of errors)")
	return cmd
}

// runInstall function prepares configs and installs KuberLogic by calling kubectl and kustomize binaries
// it then uses client-go to get some config values and viper to write config file to disk
func runInstall(k8sclient kubernetes.Interface) func(command *cobra.Command, args []string) error {
	return func(command *cobra.Command, args []string) error {
		command.Println("Preparing KuberLogic configs...")
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

		// handle kuberlogic parameters
		klParams := viper.New()

		klConfigFile := filepath.Join(cacheDir, "manager", "kuberlogic-config.env")
		if err := os.MkdirAll(filepath.Dir(klConfigFile), os.ModePerm); err != nil {
			return err
		}
		klParams.SetConfigFile(klConfigFile)
		err = klParams.ReadInConfig()
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}

		if value, err := getStringPrompt(command, tokenFlag, viper.GetString(tokenFlag), nil); err != nil {
			return errors.Wrapf(err, "error processing %s flag", tokenFlag)
		} else if value != "" {

			klParams.Set(tokenFlag, value)
			// also set global viper key (for user config)
			viper.Set(tokenFlag, value)
		} else {
			return errTokenEmpty
		}

		cachedDockerCompose := filepath.Join(cacheDir, "manager/docker-compose.yaml")
		if value, err := getStringPrompt(command, installDockerComposeParam, cachedConfigOrEmpty(cachedDockerCompose), nil); err != nil {
			return errors.Wrapf(err, "error processing %s flag", installDockerComposeParam)
		} else if value != "" {
			if err := cacheConfigFile(value, cachedDockerCompose); err != nil {
				return errors.Wrapf(err, "error caching %s config file", value)
			}
		}

		var backupsEnabled, snapshotsEnabled bool
		if backupsEnabled, err = getBoolPrompt(command, klParams.GetBool(installBackupsEnabledParam), installBackupsEnabledParam); err != nil {
			return errors.Wrapf(err, "error processing %s flag", installBackupsEnabledParam)
		} else if backupsEnabled {
			snapshotsEnabled, err = getBoolPrompt(command, klParams.GetBool(installBackupsSnapshotsEnabledParam), installBackupsSnapshotsEnabledParam)
			if err != nil {
				return errors.Wrapf(err, "error processing %s flag", installBackupsSnapshotsEnabledParam)
			}
		}
		klParams.Set(installBackupsEnabledParam, backupsEnabled)
		klParams.Set(installBackupsSnapshotsEnabledParam, snapshotsEnabled)

		var cSite, cKey, kuberlogicDomain string
		if cSite, err = getStringPrompt(command, installChargebeeSiteParam, klParams.GetString(installChargebeeSiteParam), nil); err != nil {
			return errors.Wrapf(err, "error processing %s flag", installChargebeeSiteParam)
		} else if cSite != "" {
			cKey, err = getStringPrompt(command, installChargebeeKeyParam, klParams.GetString(installChargebeeKeyParam), func(s string) error {
				if s == "" {
					return errChargebeeKeyNotSet
				}
				return nil
			})
			if err != nil {
				return errors.Wrapf(err, "error processing %s flag", installChargebeeKeyParam)
			}

			kuberlogicDomain, err = getStringPrompt(command, installKuberlogicDomainParam, klParams.GetString(installKuberlogicDomainParam), nil)
			if err != nil {
				return errors.Wrapf(err, "error processing %s flag", installKuberlogicDomainParam)
			}
		}
		klParams.Set(installChargebeeSiteParam, cSite)
		klParams.Set(installChargebeeKeyParam, cKey)
		klParams.Set(installKuberlogicDomainParam, kuberlogicDomain)

		cachedTlsCrt := filepath.Join(cacheDir, "certificates/tls.crt")
		if tlsCertPath, err := getStringPrompt(command, installTLSCrtParam, cachedConfigOrEmpty(cachedTlsCrt), nil); err != nil {
			return errors.Wrapf(err, "error processing %s flag", installTLSCrtParam)
		} else if tlsCertPath != "" {
			cachedTlsKey := filepath.Join(cacheDir, "/certificates/tls.key")
			tlsKeyPath, err := getStringPrompt(command, installTLSKeyParam, cachedConfigOrEmpty(cachedTlsKey), func(s string) error {
				if s == "" {
					return errors.New("tls certificate is set but tls key is missing")
				}
				return nil
			})
			if err != nil {
				return errors.Wrapf(err, "error processing %s flag", installTLSKeyParam)
			}

			if err := cacheConfigFile(tlsCertPath, cachedTlsCrt); err != nil {
				return errors.Wrap(err, "failed to cache tls.crt")
			}
			if err := cacheConfigFile(tlsKeyPath, cachedTlsKey); err != nil {
				return errors.Wrap(err, "failed to cache tls.key")
			}
		}

		var useKLSentry bool
		var sentrDSN string
		if useKLSentry, err = getBoolPrompt(command, klParams.GetBool(installReportErrors), installReportErrors); err != nil {
			return errors.Wrapf(err, "error processing %s flag", installReportErrors)
		} else if useKLSentry {
			sentrDSN = klSentryDSN
		} else {
			sentrDSN, err = getStringPrompt(command, installSentryDSNParam, klParams.GetString(installSentryDSNParam), nil)
			if err != nil {
				return errors.Wrapf(err, "error processing %s flag", installSentryDSNParam)
			}
		}
		klParams.Set(installReportErrors, useKLSentry)
		klParams.Set(installSentryDSNParam, sentrDSN)

		ingressClasses, err := k8sclient.NetworkingV1().IngressClasses().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list available ingress classes")
		}
		var availableIngressClasses []string
		for _, ic := range ingressClasses.Items {
			availableIngressClasses = append(availableIngressClasses, ic.GetName())
		}
		if ingressClass, err := getSelectPrompt(command, installIngressClassName, klParams.GetString(installIngressClassName), availableIngressClasses); err != nil {
			return errors.Wrapf(err, "error processing %s flag", installIngressClassName)
		} else {
			klParams.Set(installIngressClassName, ingressClass)
		}

		storageClasses, err := k8sclient.StorageV1().StorageClasses().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return errors.Wrapf(err, "failed to list available storage classes")
		}
		var availableStorageClasses []string
		for _, sc := range storageClasses.Items {
			availableStorageClasses = append(availableStorageClasses, sc.GetName())
		}
		if storageClass, err := getSelectPrompt(command, installStorageClassName, klParams.GetString(installStorageClassName), availableStorageClasses); err != nil {
			return errors.Wrapf(err, "error processing %s flag", installStorageClassName)
		} else {
			klParams.Set(installStorageClassName, storageClass)
		}

		// write config file
		if err := klParams.WriteConfigAs(klConfigFile); err != nil {
			return errors.Wrap(err, "failed to write Kuberlogic installation config file")
		}

		// copy cached configuration files to kustomize configs
		if err := useCachedConfigFiles(cacheDir, kustomizeRootDir, command.Printf); err != nil {
			return errors.Wrap(err, "failed to restore cached configs")
		}

		// check if kubernetes is available
		if _, err := k8sclient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{}); err != nil {
			return err
		}
		if out, err := exec.Command(kubectlBin, "cluster-info").CombinedOutput(); err != nil {
			command.Println(string(out))
			return errors.New("Kubernetes is not available via kubectl")
		}

		// run kustomize via exec and apply manifests via kubectl
		command.Println("Installing cert-manager...")
		cmd := fmt.Sprintf("%s build %s/cert-manager | %s apply -f -", kustomizeBin, kustomizeRootDir, kubectlBin)
		out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
		command.Println(string(out))
		if err != nil {
			return err
		}
		command.Println("Waiting for cert-manager to be ready...")
		cmd = fmt.Sprintf("%s -n cert-manager wait --for=condition=ready pods --all --timeout=300s", kubectlBin)
		if out, err := exec.Command("sh", "-c", cmd).CombinedOutput(); err != nil {
			command.Println(string(out))
			return errors.New("Failed installing cert-manager")
		}

		command.Println("Installing KuberLogic...")
		cmd = fmt.Sprintf("%s build %s/default | %s apply -f -", kustomizeBin, kustomizeRootDir, kubectlBin)
		out, err = exec.Command("sh", "-c", cmd).CombinedOutput()
		command.Println(string(out))
		if err != nil {
			return err
		}
		cmd = fmt.Sprintf("%s -n kuberlogic wait --for=condition=ready pods --all --timeout=300s", kubectlBin)
		if out, err := exec.Command("sh", "-c", cmd).CombinedOutput(); err != nil {
			command.Println(string(out))
			return errors.New("Failed installing kuberlogic")
		}

		command.Println("Fetching KuberLogic endpoint...")
		endpoint, err := getKuberlogicEndpoint(k8sclient)
		if err != nil {
			return errors.Wrap(err, "failed to get KuberLogic api server host")
		}
		viper.Set(apiHostFlag, endpoint)

		command.Println("Updating KuberLogic config file at " + configFile + "...")
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

	for _, f := range reader.File {
		fp := filepath.Join(dir, f.Name)
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
		if err != nil {
			return errors.Wrap(err, "error accessing "+path)
		}

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

		printf("Using data `%s` as %s\n", string(data), relative)
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

func cachedConfigOrEmpty(path string) string {
	if _, err := os.Stat(path); err != nil {
		return ""
	}
	return path
}

func getKuberlogicEndpoint(c kubernetes.Interface) (string, error) {
	var endpoint string
	var maxLBWaitSec = 60
	var svc *corev1.Service
	var err error

	for ; maxLBWaitSec > 0; maxLBWaitSec -= 1 {
		time.Sleep(time.Second)

		svc, err = c.CoreV1().Services("kuberlogic").Get(context.TODO(), "kls-api-server", metav1.GetOptions{})
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
