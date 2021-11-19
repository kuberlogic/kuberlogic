package cfg

import (
	"bufio"
	"fmt"
	logger "github.com/kuberlogic/kuberlogic/modules/installer/log"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-password/password"
	"io"
	"strings"
)

func readString(reader *bufio.Reader, defaultValue string) (*string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, errors.New("could not read from reader")
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return &defaultValue, nil
	}

	return &line, nil
}

func generatePassword(log logger.Logger) string {
	value, err := password.Generate(8, 3, 0, true, false)
	if err != nil {
		log.Fatalf("unable to generate default password", err)
	}
	return value
}

func AskConfig(screen io.Reader, log logger.Logger, defaultCfgLocation string) *Config {
	config := new(Config)

	reader := bufio.NewReader(screen)
	log.Infof("Config is not found, please answer several questions")
	fmt.Printf(fmt.Sprintf(`kubeconfig path: (default=%s):> `, DefaultKubeconfigPath))
	if kubeconfigPath, err := readString(reader, DefaultKubeconfigPath); err != nil {
		log.Fatalf("cannot parse kubeconfigPath: %+v", err)
	} else {
		config.KubeconfigPath = kubeconfigPath
	}

	defaultNamespace := "kuberlogic"
	fmt.Printf(fmt.Sprintf(`Namespace: (default=%s):> `, defaultNamespace))
	if ns, err := readString(reader, defaultNamespace); err != nil {
		log.Fatalf("cannot parse Namespace: %+v", err)
	} else {
		config.Namespace = ns
	}

	defaultKuberlogicEndpoint := "example.com"
	fmt.Printf(fmt.Sprintf(`Kuberlogic endpoint: (default=%s):> `, defaultKuberlogicEndpoint))
	if endpoint, err := readString(reader, defaultKuberlogicEndpoint); err != nil {
		log.Fatalf("cannot parse Kuberlogic endpoint: %+v", err)
	} else {
		config.Endpoints.Kuberlogic = *endpoint
	}

	defaultMonitoringEndpoint := "mc.example.com"
	fmt.Printf(fmt.Sprintf(`Monitoring endpoint: (default=%s):> `, defaultMonitoringEndpoint))
	if endpoint, err := readString(reader, defaultMonitoringEndpoint); err != nil {
		log.Fatalf("cannot parse Monitoring endpoint: %+v", err)
	} else {
		config.Endpoints.MonitoringConsole = *endpoint
	}

	defaultAdminPassword := generatePassword(log)
	fmt.Printf(fmt.Sprintf(`Admin password: (default=%s):> `, defaultAdminPassword))
	if adminPassword, err := readString(reader, defaultAdminPassword); err != nil {
		log.Fatalf("cannot parse Admin password: %+v", err)
	} else {
		config.Auth.AdminPassword = *adminPassword
	}

	defaultDemoUserPassword := generatePassword(log)
	fmt.Printf(fmt.Sprintf(`Demo user password: (default=%s):> `, defaultDemoUserPassword))
	if demoUserPassword, err := readString(reader, defaultDemoUserPassword); err != nil {
		log.Fatalf("cannot parse Demo user password: %+v", err)
	} else {
		config.Auth.DemoUserPassword = demoUserPassword
	}
	if err := config.SetDefaults(log); err != nil {
		log.Fatalf("cannot set default values for config %+v", err)
	}

	if err := newFileFromConfig(config, defaultCfgLocation); err != nil {
		log.Fatalf("cannot create config file %+v", err)
	}

	return config
}
