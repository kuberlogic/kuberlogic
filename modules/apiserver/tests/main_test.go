/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tests

import (
	"flag"
	"fmt"
	cmd2 "github.com/kuberlogic/kuberlogic/modules/apiserver/cmd"
	"github.com/kuberlogic/kuberlogic/modules/operator/cmd"
	"github.com/prometheus/common/log"
	"net/url"
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
	name:  "pgsql",
	ns:    testNs,
	type_: "postgresql",
}
var mysqlService = Service{
	name:  "mys",
	ns:    testNs,
	type_: "mysql",
}

var services = []Service{
	pgService,
	mysqlService,
}

// default api contact values
var (
	apiUrl = &url.URL{
		Scheme: "http",
		Host:   "127.0.0.1:8001",
	}
)

// these values are related to keycloak values in config/keycloak manifests
const (
	testUser     = "user@kuberlogic.com"
	testNs       = "ad7fa60af4d66b8aaa4c693a92a7ff87" // md5 hashed testUser email
	testPassword = "secret"
)

func setup() {
	if apiAddr := os.Getenv("REMOTE_HOST"); apiAddr == "" {
		log.Info("Starting the apiserver in the goroutine...")
		args := []string{"--scheme=http"}
		go cmd2.Main(args) // start the api server
		log.Info("Starting the operator in the goroutine...")
		go cmd.Main([]string{}) // start the operator
		log.Info("Waiting 15 seconds for the starting goroutines...")
		time.Sleep(15 * time.Second)
	} else {
		url, err := url.Parse(apiAddr)
		if err != nil {
			panic(fmt.Sprintf("REMOTE_HOST is not a valid url: %v", err))
		}
		apiUrl = url
	}

	if isWait := os.Getenv("WAIT_BEFORE_CREATE"); isWait == "true" {
		wait(60 * 60 * 60)(&testing.T{}) // wait the 1 hour
	}
	flag.Parse()
	if testing.Short() {
		parallelFunc(createService)
	}
	if isWait := os.Getenv("WAIT_BEFORE_TESTS"); isWait == "true" {
		wait(60 * 60 * 60)(&testing.T{}) // wait the 1 hour
	}
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
		limits:   map[string]string{"cpu": "0.25", "memory": "0.5", "volumeSize": "1"},
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
	} else if isWait := os.Getenv("WAIT_IF_TESTS_FAILS"); isWait == "true" {
		wait(60 * 60 * 60)(&testing.T{}) // wait the 1 hour
	}

	os.Exit(code)
}
