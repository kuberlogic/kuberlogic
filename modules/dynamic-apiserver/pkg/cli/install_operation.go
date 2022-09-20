package cli

import (
	"archive/zip"
	"context"
	_ "embed"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/ghodss/yaml"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xeipuuv/gojsonschema"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// embed kustomize files into cli binary
//go:embed kustomize-configs.zip
var klConfigZipData []byte

var (
	kubectlBin = "kubectl"

	errTokenEmpty         = errors.New("token can't be empty")
	errSentryInvalidURI   = errors.New("invalid uri")
	errDirFound           = errors.New("text file required, directory found")
	errVeleroNotAvailable = errors.New("velero resources are not available")
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
	installChargebeeMappingParam        = "chargebee_mapping"
	installKuberlogicDomainParam        = "kuberlogic_domain"
	installReportErrors                 = "report_errors"
	installSentryDSNParam               = "sentry_dsn"
	installDeploymentId                 = "deployment_id"
)

var (
	chargebeeMappingSchema = map[string]interface{}{
		"type": "array",
		"items": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"src": map[string]interface{}{
					"type": "string",
				},
				"dst": map[string]interface{}{
					"type": "string",
				},
			},
			"required":             []string{"src", "dst"},
			"additionalProperties": false,
		},
	}
)

func makeInstallCmd(k8sclient kubernetes.Interface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Installs KuberLogic to Kubernetes cluster",
		RunE:  runInstall(k8sclient),
	}

	_ = cmd.PersistentFlags().Bool("non-interactive", false, "Do not enter interactive mode")
	_ = cmd.PersistentFlags().String(installIngressClassName, "", "Choose Kubernetes ingress class that will be used to configure external access for application instances.")
	_ = cmd.PersistentFlags().String(installStorageClassName, "", "Choose Kubernetes storage class that will be used to configure storage volumes for application instances.")
	_ = cmd.PersistentFlags().String(installDockerComposeParam, "", "Specify the path to your docker-compose file with the application you want to provide as SaaS.\nSee https://kuberlogic.com/docs/configuring/docker-compose for additional information. You can skip this step by pressing 'Enter', then the sample application will be used.")
	_ = cmd.PersistentFlags().Bool(installBackupsEnabledParam, false, "Enable backup/restore support\nFor more information, read https://kuberlogic.com/docs/configuring/backups for more information. Choose 'no' if you have not set up integration with Velero to support backup/restore capabilities, otherwise choose 'yes'")
	_ = cmd.PersistentFlags().Bool(installBackupsSnapshotsEnabledParam, false, "Enable volume snapshot backups (Must be supported by the Velero provider plugin).")
	_ = cmd.PersistentFlags().String(installTLSCrtParam, "", "Specify path to the TLS certificate.\nIt is assumed that the TLS certificate will be a wildcard certificate. All applications managed by Kuberlogic share the same certificate by sharing the same ingress controller. You can skip this step by pressing 'Enter', In this case, a self-signed (demo) certificate will be used.")
	_ = cmd.PersistentFlags().String(installTLSKeyParam, "", "Specify path to TLS key to use for provisioned applications.")
	_ = cmd.PersistentFlags().String(installChargebeeSiteParam, "", "Specify ChargeBee site name.\nFor more information, read https://kuberlogic.com/docs/configuring/billing page. You can skip this step by pressing 'Enter', and set up the integration later.")
	_ = cmd.PersistentFlags().String(installChargebeeKeyParam, "", "Specify ChargeBee API key.\nFor more information, read https://apidocs.chargebee.com/docs/api/?prod_cat_ver=2 . API Key are used to configure Kuberlogic ChargeBee integration.")
	_ = cmd.PersistentFlags().String(installChargebeeMappingParam, "", "Specify ChargeBee mapping file.\nFor more information, read https://kuberlogic.com/docs/configuring/billing/#mapping-custom-fields.")
	_ = cmd.PersistentFlags().String(installKuberlogicDomainParam, "", "Specify “Domain name”.\nThis configuration setting is used by KuberLogic to create endpoints for application instances. (e.g. instance1.domainname.com).")
	_ = cmd.PersistentFlags().Bool(installReportErrors, false, "Report errors to KuberLogic?\nChoose 'yes' if you want to help us improve KuberLogic, otherwise, select 'no'. Error reports will be generated and sent automatically, these reports contain only information about the errors and do not contain any user data. Let us receive errors at least from your test environments.")
	_ = cmd.PersistentFlags().String(installSentryDSNParam, "", "Specify Sentry Data Source Name (DSN).\nFor more information, read https://docs.sentry.io/product/sentry-basics/dsn-explainer/ . (KuberLogic team will not be notified in case of errors).")
	return cmd
}

// runInstall function prepares configs and installs KuberLogic by calling kubectl and kustomize binaries
// it then uses client-go to get some config values and viper to write config file to disk
func runInstall(k8sclient kubernetes.Interface) func(command *cobra.Command, args []string) error {
	return func(command *cobra.Command, args []string) error {
		command.Println("Checking available environment...")

		// check if kubernetes is available
		if _, err := k8sclient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{}); err != nil {
			return err
		}
		if out, err := exec.Command(kubectlBin, "cluster-info").CombinedOutput(); err != nil {
			command.Println(string(out))
			return errors.New("kubernetes is not available via kubectl")
		}

		// collect storageClasses
		storageClasses, err := k8sclient.StorageV1().StorageClasses().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return errors.Wrapf(err, "failed to list available storage classes")
		}
		if len(storageClasses.Items) == 0 {
			return errors.New("storage classes not found")
		}

		// collect ingressClasses
		ingressClasses, err := k8sclient.NetworkingV1().IngressClasses().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list available ingress classes")
		}
		if len(ingressClasses.Items) == 0 {
			return errors.New("ingress classes not found")
		}

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
		configBaseDir := filepath.Dir(viper.ConfigFileUsed())
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
		deploymentId := klParams.GetString(installDeploymentId)
		if deploymentId == "" {
			klParams.Set(installDeploymentId, uuid.New().String())
		}

		if value, err := getStringPrompt(command, tokenFlag, viper.GetString(tokenFlag), true, nil); err != nil {
			return errors.Wrapf(err, "error processing %s flag", tokenFlag)
		} else if value != "" {

			klParams.Set(tokenFlag, value)
			// also set global viper key (for user config)
			viper.Set(tokenFlag, value)
		} else {
			return errTokenEmpty
		}

		cachedDockerCompose := filepath.Join(cacheDir, "manager/docker-compose.yaml")
		if value, err := getStringPrompt(command, installDockerComposeParam, cachedConfigOrEmpty(cachedDockerCompose), false, validateFileAvailable); err != nil {
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
			// check velero
			if out, err := exec.Command("sh", "-c", kubectlBin+" get crd backups.velero.io").CombinedOutput(); err != nil {
				fmt.Println(string(out))
				return errVeleroNotAvailable
			}

			snapshotsEnabled, err = getBoolPrompt(command, klParams.GetBool(installBackupsSnapshotsEnabledParam), installBackupsSnapshotsEnabledParam)
			if err != nil {
				return errors.Wrapf(err, "error processing %s flag", installBackupsSnapshotsEnabledParam)
			}
		}
		klParams.Set(installBackupsEnabledParam, backupsEnabled)
		klParams.Set(installBackupsSnapshotsEnabledParam, snapshotsEnabled)

		var cSite, cKey, cMappingFile, kuberlogicDomain string
		cachedChargebeeMappingFile := filepath.Join(cacheDir, "manager/mapping-fields.yaml")
		if cSite, err = getStringPrompt(command, installChargebeeSiteParam, klParams.GetString(installChargebeeSiteParam), false, nil); err != nil {
			return errors.Wrapf(err, "error processing %s flag", installChargebeeSiteParam)
		} else if cSite != "" {
			cKey, err = getStringPrompt(command, installChargebeeKeyParam, klParams.GetString(installChargebeeKeyParam), true, nil)
			if err != nil {
				return errors.Wrapf(err, "error processing %s flag", installChargebeeKeyParam)
			}
			cMappingFile, err = getStringPrompt(command, installChargebeeMappingParam, klParams.GetString(installChargebeeMappingParam), true, validateChargebeeMappingfile)
			if err != nil {
				return errors.Wrapf(err, "error processing %s flag", installChargebeeMappingParam)
			}
			if cMappingFile != "" {
				if err = cacheConfigFile(cMappingFile, cachedChargebeeMappingFile); err != nil {
					return errors.Wrapf(err, "error caching %s config file", cachedChargebeeMappingFile)
				}
			}
		}
		kuberlogicDomain, err = getStringPrompt(command, installKuberlogicDomainParam, klParams.GetString(installKuberlogicDomainParam), true, nil)
		if err != nil {
			return errors.Wrapf(err, "error processing %s flag", installKuberlogicDomainParam)
		} else if kuberlogicDomain == "" {
			return errors.New("kuberlogic domain cannot be empty")
		}

		klParams.Set(installChargebeeSiteParam, cSite)
		klParams.Set(installChargebeeKeyParam, cKey)
		klParams.Set(installKuberlogicDomainParam, kuberlogicDomain)

		cachedTlsCrt := filepath.Join(cacheDir, "certificate/tls.crt")
		if tlsCertPath, err := getStringPrompt(command, installTLSCrtParam, cachedConfigOrEmpty(cachedTlsCrt), false, validateFileAvailable); err != nil {
			return errors.Wrapf(err, "error processing %s flag", installTLSCrtParam)
		} else if tlsCertPath != "" {
			cachedTlsKey := filepath.Join(cacheDir, "/certificate/tls.key")
			tlsKeyPath, err := getStringPrompt(command, installTLSKeyParam, cachedConfigOrEmpty(cachedTlsKey), false, validateFileAvailable)
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
			sentrDSN, err = getStringPrompt(command, installSentryDSNParam, klParams.GetString(installSentryDSNParam), false, validateEmptyStrOrUri)
			if err != nil {
				return errors.Wrapf(err, "error processing %s flag", installSentryDSNParam)
			}
		}
		klParams.Set(installReportErrors, useKLSentry)
		klParams.Set(installSentryDSNParam, sentrDSN)

		var availableIngressClasses []string
		for _, ic := range ingressClasses.Items {
			availableIngressClasses = append(availableIngressClasses, ic.GetName())
		}
		if ingressClass, err := getSelectPrompt(command, installIngressClassName, klParams.GetString(installIngressClassName), availableIngressClasses); err != nil {
			return errors.Wrapf(err, "error processing %s flag", installIngressClassName)
		} else {
			klParams.Set(installIngressClassName, ingressClass)
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

		// run kustomize via exec and apply manifests via kubectl
		command.Println("Installing cert-manager...")

		cmd := fmt.Sprintf("%s apply --kustomize %s/cert-manager", kubectlBin, kustomizeRootDir)
		out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
		command.Println(string(out))
		if err != nil {
			return err
		}
		command.Println("Waiting for cert-manager to be ready...")
		cmd = fmt.Sprintf("%s -n cert-manager wait --for=condition=ready pods --all --timeout=300s", kubectlBin)
		if out, err := exec.Command("sh", "-c", cmd).CombinedOutput(); err != nil {
			command.Println(string(out))
			return errors.New("failed to install cert-manager")
		}

		command.Println("Installing KuberLogic...")
		cmd = fmt.Sprintf("%s apply --kustomize %s/default", kubectlBin, kustomizeRootDir)
		out, err = exec.Command("sh", "-c", cmd).CombinedOutput()
		command.Println(string(out))
		if err != nil {
			return err
		}
		command.Println("Waiting for Kuberlogic to be ready...")
		cmd = fmt.Sprintf("%s -n kuberlogic wait --for=condition=ready pods --all --timeout=300s", kubectlBin)
		if out, err := exec.Command("sh", "-c", cmd).CombinedOutput(); err != nil {
			command.Println(string(out))
			return errors.New("failed to install Kuberlogic")
		}

		command.Println("Fetching KuberLogic endpoint...")
		endpoint, err := getKuberlogicServiceEndpoint("kls-api-server", 60, k8sclient)
		if err != nil {
			return errors.Wrap(err, "failed to get KuberLogic api server host")
		}
		viper.Set(apiHostFlag, endpoint)

		command.Println("Updating KuberLogic config file at " + viper.ConfigFileUsed())
		if err = viper.WriteConfig(); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				if err := viper.WriteConfigAs(viper.ConfigFileUsed()); err != nil {
					return errors.Wrap(err, "failed to write KuberLogic config")
				}
			} else {
				return errors.Wrap(err, "failed to write KuberLogic config")
			}
		}

		command.Printf("Installation completed successfully.\nRun `%s info` to see information about your Kuberlogic installation.", exeName)
		return nil
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
	if err := os.MkdirAll(filepath.Dir(name), os.ModePerm); err != nil {
		return err
	}
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

// getKuberlogicServiceEndpoint gets an access endpoint for a kuberlogic service named `name`
func getKuberlogicServiceEndpoint(name string, maxLbWaitSec int, c kubernetes.Interface) (string, error) {
	var endpoint string
	var svc *corev1.Service
	var err error

	for ; maxLbWaitSec > 0; maxLbWaitSec -= 1 {
		time.Sleep(time.Second)

		svc, err = c.CoreV1().Services("kuberlogic").Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return "", errors.Wrap(err, "failed to get KuberLogic Service")
		}
		port := strconv.Itoa(int(svc.Spec.Ports[0].Port))

		if extIPs := svc.Spec.ExternalIPs; len(extIPs) != 0 {
			endpoint = extIPs[0]
		}

		if lbIPs := svc.Status.LoadBalancer.Ingress; len(lbIPs) != 0 {
			if ing := lbIPs[0]; ing.Hostname != "" {
				endpoint = ing.Hostname
			} else {
				endpoint = ing.IP
			}
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

func validateEmptyStrOrUri(uri string) error {
	if uri == "" {
		return nil
	}
	u, err := url.ParseRequestURI(uri)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return errors.Wrap(errSentryInvalidURI, uri)
	}
	return nil
}

func validateFileAvailable(f string) error {
	fInfo, err := os.Stat(f)
	if err != nil {
		return err
	}
	if fInfo.IsDir() {
		return errDirFound
	}
	return nil
}

func validateChargebeeMappingfile(f string) error {
	if f == "" {
		return nil
	}
	if err := validateFileAvailable(f); err != nil {
		return err
	}
	yamlFile, err := ioutil.ReadFile(f)
	if err != nil {
		return errors.Errorf("cannot read the file %s: #%v", f, err)
	}
	mapping := make([]map[string]string, 0)
	err = yaml.Unmarshal(yamlFile, &mapping)
	if err != nil {
		return errors.Errorf("cannot parse the file %s: #%v", f, err)
	}
	schemaLoader := gojsonschema.NewGoLoader(chargebeeMappingSchema)
	dataLoader := gojsonschema.NewGoLoader(mapping)
	result, err := gojsonschema.Validate(schemaLoader, dataLoader)
	if err != nil {
		return err
	}
	if !result.Valid() {
		return errors.Errorf("Invalid yaml schema: %v", result.Errors())
	}
	return nil
}
