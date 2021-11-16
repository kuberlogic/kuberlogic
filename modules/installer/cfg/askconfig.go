package cfg

import (
	"bufio"
	"fmt"
	logger "github.com/kuberlogic/kuberlogic/modules/installer/log"
	"github.com/pkg/errors"
	"golang.org/x/term"
	"os"
	"strings"
	"syscall"
)

func readString(reader *bufio.Reader, defaultValue string) (*string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, errors.New("could not read for debug logs")
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return &defaultValue, nil
	}

	return &line, nil
}

func readPassword() (*string, error) {
	bytePassword, err := term.ReadPassword(syscall.Stdin)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read the password")
	}

	password := strings.TrimSpace(string(bytePassword))
	return &password, nil
}

func AskConfig(log logger.Logger, defaultCfgLocation string) *Config {
	config := new(Config)

	reader := bufio.NewReader(os.Stdin)
	log.Infof("Config is not found, please answer several questions")

	fmt.Println(fmt.Sprintf(`kubeconfig path: (default=%s)`, DefaultKubeconfigPath))
	if kubeconfigPath, err := readString(reader, DefaultKubeconfigPath); err != nil {
		log.Fatalf("cannot parse kubeconfigPath: %+v", err)
	} else {
		config.KubeconfigPath = kubeconfigPath
		log.Infof(`Using "%s" for kubeconfig path`, *kubeconfigPath)
	}

	defaultNamespace := "kuberlogic"
	log.Infof(fmt.Sprintf(`Namespace: (default=%s)`, defaultNamespace))
	if ns, err := readString(reader, defaultNamespace); err != nil {
		log.Fatalf("cannot parse Namespace: %+v", err)
	} else {
		config.Namespace = ns
		log.Infof(`Using "%s" for namespace`, *ns)
	}

	defaultKuberlogicEndpoint := "example.com"
	log.Infof(fmt.Sprintf(`Kuberlogic endpoint: (default=%s)`, defaultKuberlogicEndpoint))
	if endpoint, err := readString(reader, defaultKuberlogicEndpoint); err != nil {
		log.Fatalf("cannot parse Kuberlogic endpoint: %+v", err)
	} else {
		config.Endpoints.Kuberlogic = *endpoint
		log.Infof(`Using "%s" for kuberlogic endpoint`, *endpoint)
	}

	defaultMonitoringEndpoint := "mc.example.com"
	log.Infof(fmt.Sprintf(`Monitoring endpoint: (default=%s)`, defaultMonitoringEndpoint))
	if endpoint, err := readString(reader, defaultMonitoringEndpoint); err != nil {
		log.Fatalf("cannot parse Monitoring endpoint: %+v", err)
	} else {
		config.Endpoints.MonitoringConsole = *endpoint
		log.Infof(`Using "%s" for monitoring endpoint`, *endpoint)
	}

	defaultAdminPassword := ""
	log.Infof(fmt.Sprintf(`Admin password: (default=%s)`, defaultAdminPassword))
	if adminPassword, err := readPassword(); err != nil {
		log.Fatalf("cannot parse Admin password: %+v", err)
	} else {
		config.Auth.AdminPassword = *adminPassword
		log.Infof(`Using "*****" for admin password`)
	}

	defaultDemoUserPassword := ""
	log.Infof(fmt.Sprintf(`Demo user password: (default=%s)`, defaultDemoUserPassword))
	if demoUserPassword, err := readPassword(); err != nil {
		log.Fatalf("cannot parse Demo user password: %+v", err)
	} else {
		config.Auth.DemoUserPassword = demoUserPassword
		log.Infof(`Using "*****" for demo user password`)
	}
	if err := config.SetDefaults(log); err != nil {
		log.Fatalf("cannot set default values for config %+v", err)
	}

	if err := newFileFromConfig(config, defaultCfgLocation); err != nil {
		log.Fatalf("cannot create config file %+v", err)
	}

	return config
}
