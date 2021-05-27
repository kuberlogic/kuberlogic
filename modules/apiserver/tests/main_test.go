package tests

import (
	"flag"
	cmd2 "github.com/kuberlogic/operator/modules/apiserver/cmd"
	"github.com/kuberlogic/operator/modules/operator/cmd"
	"github.com/prometheus/common/log"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

type Service struct {
	ns      string
	name    string
	type_   string
	version string
}

var pgService = Service{
	ns:    "default",
	name:  "pgsql",
	type_: "postgresql",
}
var mysqlService = Service{
	ns:    "default",
	name:  "my",
	type_: "mysql",
}

var services = []Service{
	pgService,
	mysqlService,
}

const (
	testUser     = "kuberlogic@cloudlinux.com"
	testPassword = "password"
)

func setup() {
	args := []string{"--scheme=http"}
	log.Info("Starting the apiserver in the goroutine...")
	go cmd2.Main(args) // start the api server
	log.Info("Starting the operator in the goroutine...")
	go cmd.Main([]string{}) // start the operator
	log.Info("Waiting 15 seconds for the starting goroutines...")
	time.Sleep(15 * time.Second)

	flag.Parse()
	if testing.Short() {
		parallelFunc(createService)
	}

	//wait(60 * 60)(&testing.T{}) // for the manual tests
}

func tearDown() {
	if testing.Short() {
		parallelFunc(deleteService)
	}
}

func createService(service Service) {
	ts := tService{
		ns:       service.ns,
		name:     service.name,
		type_:    service.type_,
		version:  service.version,
		force:    true,
		replicas: 0,
		limits:   map[string]string{"cpu": "250m", "memory": "512Mi", "volumeSize": "1Gi"},
	}
	log.Infof("Creating a single %s:%s (%s) service", service.ns, service.name, service.type_)
	ts.Create(&testing.T{})
	ts.WaitForStatus("Ready", 5, 5*60)(&testing.T{})
}

func deleteService(service Service) {
	ts := tService{ns: service.ns, name: service.name, type_: service.type_, force: true}
	log.Infof("Deleting a %s:%s (%s) service", service.ns, service.name, service.type_)
	ts.Delete(&testing.T{})
}

func parallelFunc(f func(service Service)) {
	var wg sync.WaitGroup

	type_ := serviceType()
	for _, s := range services {
		usingVersion(&s)
		if type_ == "" || type_ == s.type_ {
			wg.Add(1)
			go func(svc Service) {
				defer wg.Done()
				f(svc)
			}(s)
		}
	}
	wg.Wait()
}

func usingVersion(service *Service) {
	var env string
	if service.type_ == "postgresql" {
		env = "PG_VERSION"

	} else if service.type_ == "mysql" {
		env = "MY_VERSION"
	}

	if version := os.Getenv(env); version != "" {
		log.Infof("Using version %s for %s", version, service.type_)
		service.version = version
	}
}

func serviceType() string {
	// RUN - TestService/postgresql, TestService, /postgresql
	s := strings.Split(os.Getenv("RUN"), "/")
	if len(s) == 2 {
		return s[1]
	}
	return ""
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	if code == 0 {
		// no need destroy if the tests are failed
		tearDown()
	}
	os.Exit(code)
}
