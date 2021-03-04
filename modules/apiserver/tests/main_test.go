package tests

import (
	cmd2 "github.com/kuberlogic/operator/modules/apiserver/cmd"
	"github.com/kuberlogic/operator/modules/operator/cmd"
	"github.com/prometheus/common/log"
	"os"
	"sync"
	"testing"
	"time"
)

type Service struct {
	ns    string
	name  string
	type_ string
}

var pgService = Service{"default", "pgsql", "postgresql"}
var mysqlService = Service{"default", "my", "mysql"}

var services = []Service{
	pgService,
	mysqlService,
}

func setup() {
	args := []string{"--scheme=http"}
	log.Info("Starting the apiserver in the goroutine...")
	go cmd2.Main(args) // start the api server
	log.Info("Starting the operator in the goroutine...")
	go cmd.Main([]string{}) // start the operator
	log.Info("Waiting 15 seconds for the starting goroutines...")
	time.Sleep(15 * time.Second)

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
	ts := tService{ns: service.ns, name: service.name, type_: service.type_, force: true, replicas: 0}
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
	wg.Add(len(services))

	for _, s := range services {
		go func(svc Service) {
			defer wg.Done()
			f(svc)
		}(s)
	}
	wg.Wait()
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	tearDown()
	os.Exit(code)
}
