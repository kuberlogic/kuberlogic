package cfg

import (
	logger "github.com/kuberlogic/kuberlogic/modules/installer/log"
	"os"
	"strings"
	"testing"
)

func TestAskConfig(t *testing.T) {
	log := logger.NewLogger(true)
	cfgLocation := "/tmp/test-kuberlogic-config.yaml"
	defer os.Remove(cfgLocation)

	ns := "custom_ns"
	kuberlogicEndpoint := "my.endpoint.com"
	consoleEndpoint := "mc.my.endpoint.com"
	adminPassword := "secret"
	demoPassword := "secret-demo"

	content := strings.NewReader(strings.Join([]string{
		"",
		ns,
		kuberlogicEndpoint,
		consoleEndpoint,
		adminPassword,
		demoPassword + "\n",
	}, "\n"))

	config := AskConfig(content, log, cfgLocation)

	if *config.DebugLogs != false {
		t.Errorf("value is incorrect, actual vs expected: %v vs %v", *config.DebugLogs, false)
	}
	if config.Endpoints.Kuberlogic != kuberlogicEndpoint {
		t.Errorf("value is incorrect, actual vs expected: %v vs %v", config.Endpoints.Kuberlogic,
			kuberlogicEndpoint)
	}
	if config.Endpoints.KuberlogicTLS != nil {
		t.Errorf("value is incorrect, actual vs expected: %v vs %v", config.Endpoints.KuberlogicTLS, nil)
	}
	if config.Endpoints.MonitoringConsole != consoleEndpoint {
		t.Errorf("value is incorrect, actual vs expected: %v vs %v", config.Endpoints.MonitoringConsoleTLS,
			consoleEndpoint)
	}
	if *config.Namespace != ns {
		t.Errorf("value is incorrect, actual vs expected: %v vs %v", config.Endpoints.KuberlogicTLS, ns)
	}
	if *config.KubeconfigPath != DefaultKubeconfigPath {
		t.Errorf("value is incorrect, actual vs expected: %v vs %v", config.KubeconfigPath,
			DefaultKubeconfigPath)
	}
	if config.Auth.AdminPassword != adminPassword {
		t.Errorf("value is incorrect, actual vs expected: %v vs %v", config.Auth.AdminPassword,
			adminPassword)
	}

	if *config.Auth.DemoUserPassword != demoPassword {
		t.Errorf("value is incorrect, actual vs expected: %v vs %v", config.Auth.DemoUserPassword,
			demoPassword)
	}

	if config.Platform != "generic" {
		t.Errorf("value is incorrect, actual vs expected: %v vs %v", config.Platform, "generic")
	}

	if _, err := os.Stat(cfgLocation); err != nil {
		t.Errorf("file does not exists %s", cfgLocation)
	}
}
