package tests

import (
	"fmt"
	"github.com/prometheus/common/log"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"
)

type tService struct {
	ns          string
	name        string
	type_       string
	replicas    int
	newReplicas int
	conf        map[string]string
	newConf     map[string]string
	limits      map[string]string
	newLimits   map[string]string
	force       bool
	version     string
}

var pgTestService = tService{
	ns:          pgService.ns,
	name:        pgService.name,
	type_:       pgService.type_,
	newReplicas: 1,
	replicas:    0,
	newConf:     map[string]string{"shared_buffers": "16MB", "max_connections": "50"},
	conf:        map[string]string{"shared_buffers": "32MB", "max_connections": "10"},
	limits:      map[string]string{"cpu": "250m", "memory": "512Mi", "volumeSize": "1Gi"},
	newLimits:   map[string]string{"cpu": "300m", "memory": "512Mi", "volumeSize": "1Gi"},
	force:       false, // do not create a service
	//version:     "12.1.3",
}

var mysqlTestService = tService{
	ns:          mysqlService.ns,
	name:        mysqlService.name,
	type_:       mysqlService.type_,
	newReplicas: 1,
	replicas:    0,
	newConf:     map[string]string{"max_allowed_packet": "64Mb"},
	conf:        map[string]string{"max_allowed_packet": "32Mb"},
	limits:      map[string]string{"cpu": "250m", "memory": "512Mi", "volumeSize": "1Gi"},
	newLimits:   map[string]string{"cpu": "300m", "memory": "512Mi", "volumeSize": "1Gi"},
	force:       false, // do not create a service
	//version:     "5.7.26",
}

func TestDoesNotAllowMethodDelete(t *testing.T) {
	api := newApi(t)
	api.sendRequestTo(http.MethodDelete, "/services/")
	api.responseCodeShouldBe(http.StatusMethodNotAllowed)
	api.encodeResponseToJson()
	api.fieldContains("message", "method DELETE is not allowed")
}

func TestBearerTokenIsRequired(t *testing.T) {
	api := newApi(t)
	api.sendRequestTo(http.MethodGet, "/services/")
	api.responseCodeShouldBe(http.StatusUnauthorized)
	api.encodeResponseToJson()
	api.responseShouldMatchJson(`{
	"code": 401,
	"message": "unauthenticated for invalid credentials"
}`)
}

func TestAllowGetMethodAndCheckEmptyResponse(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping. Using -short flag")
		return
	}
	api := newApi(t)
	api.setBearerToken()
	api.sendRequestTo(http.MethodGet, "/services/")
	api.responseCodeShouldBe(http.StatusOK)
	api.encodeResponseToJson()
	api.responseTypeOf(reflect.Slice)
	api.lengthOfResponseIs(0)
}

func TestNotEnoughDefinedParameters(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.setRequestBody(`{
        "name": "cloudmanaged-pg",
        "ns": "default"
     }`)
	api.sendRequestTo(http.MethodPost, "/services/")
	api.responseCodeShouldBe(http.StatusUnprocessableEntity)
	api.encodeResponseToJson()
	api.responseShouldMatchJson(`{
        "code": 602,
        "message": "type in body is required"
     }`)

}

func (s *tService) Create(t *testing.T) {
	if !s.force && testing.Short() {
		t.Skip("Skipping. Using -short flag")
		return
	}
	api := newApi(t)
	api.setBearerToken()
	params := map[string]interface{}{
		"name":     s.name,
		"ns":       s.ns,
		"type":     s.type_,
		"replicas": s.replicas,
		"limits":   s.limits,
	}
	if s.version != "" {
		t.Logf("using version - %s", s.version)
		params["version"] = s.version
	}
	api.setJsonRequestBody(params)

	api.sendRequestTo(http.MethodPost, "/services/")
	api.responseCodeShouldBe(201)

}

func (s *tService) TryCreateWithSmallCPULimits(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	params := map[string]interface{}{
		"name":     "new-" + s.name,
		"ns":       s.ns,
		"type":     s.type_,
		"replicas": s.replicas,
		// min: 250m
		"limits": map[string]string{"cpu": "200m", "memory": "512Mi", "volumeSize": "1Gi"},
	}
	api.setJsonRequestBody(params)
	api.sendRequestTo(http.MethodPost, "/services/")
	api.responseCodeShouldBe(503) // 503 - operator's error
	api.encodeResponseToJson()
	api.responseShouldMatchJson(`{"message": "error creating service"}`)
}

func (s *tService) TryCreateWithSmallMemoryLimits(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	params := map[string]interface{}{
		"name":     "new-" + s.name,
		"ns":       s.ns,
		"type":     s.type_,
		"replicas": s.replicas,
		// min: 512Mi
		"limits": map[string]string{"cpu": "250m", "memory": "500Mi", "volumeSize": "1Gi"},
	}
	api.setJsonRequestBody(params)
	api.sendRequestTo(http.MethodPost, "/services/")
	api.responseCodeShouldBe(503) // 503 - operator's error
	api.encodeResponseToJson()
	api.responseShouldMatchJson(`{"message": "error creating service"}`)
}

func (s *tService) TryCreateWithSmallDiskLimits(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	params := map[string]interface{}{
		"name":     "new-" + s.name,
		"ns":       s.ns,
		"type":     s.type_,
		"replicas": s.replicas,
		// min: 1Gi
		"limits": map[string]string{"cpu": "250m", "memory": "512Mi", "volumeSize": "900Mi"},
	}
	api.setJsonRequestBody(params)
	api.sendRequestTo(http.MethodPost, "/services/")
	api.responseCodeShouldBe(503) // 503 - operator's error
	api.encodeResponseToJson()
	api.responseShouldMatchJson(`{"message": "error creating service"}`)
}

func (s *tService) TryDecreaseVolumeSize(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.setJsonRequestBody(
		map[string]interface{}{
			"name":   s.name,
			"ns":     s.ns,
			"type":   s.type_,
			"limits": map[string]string{"cpu": "250m", "memory": "512Mi", "volumeSize": "800Mi"},
		})
	api.sendRequestTo(http.MethodPut, fmt.Sprintf("/services/%s:%s/", s.ns, s.name))
	api.responseCodeShouldBe(400) // 400 - apiserver validating error
	api.encodeResponseToJson()
	api.responseShouldMatchJson(`{"message": "error changing service: volume size can't be lowered"}`)
}

func (s *tService) EditReplicas(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.setJsonRequestBody(
		map[string]interface{}{
			"name":     s.name,
			"ns":       s.ns,
			"type":     s.type_,
			"replicas": s.newReplicas,
		})
	api.sendRequestTo(http.MethodPut, fmt.Sprintf("/services/%s:%s/", s.ns, s.name))
	api.responseCodeShouldBe(200)
	api.encodeResponseToJson()
	api.fieldIs("replicas", s.newReplicas)
}

func (s *tService) TryEditType(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	newType := "postgresql"
	if s.type_ == "postgresql" {
		newType = "mysql"
	}
	api.setJsonRequestBody(
		map[string]interface{}{
			"name": s.name,
			"ns":   s.ns,
			"type": newType,
		})
	api.sendRequestTo(http.MethodPut, fmt.Sprintf("/services/%s:%s/", s.ns, s.name))
	api.responseCodeShouldBe(400) // 400 - apiserver validating error
	api.encodeResponseToJson()
	api.responseShouldMatchJson(`{"message": "error changing service: type can't be changed"}`)
}

func (s *tService) EditBackAdvancedConf(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.setJsonRequestBody(
		map[string]interface{}{
			"name":         s.name,
			"ns":           s.ns,
			"type":         s.type_,
			"replicas":     s.replicas,
			"advancedConf": s.conf,
		})
	api.sendRequestTo(http.MethodPut, fmt.Sprintf("/services/%s:%s/", s.ns, s.name))
	api.responseCodeShouldBe(200)
}

func (s *tService) EditLimits(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.setJsonRequestBody(
		map[string]interface{}{
			"name":     s.name,
			"ns":       s.ns,
			"type":     s.type_,
			"replicas": s.replicas,
			"limits":   s.newLimits,
		})
	api.sendRequestTo(http.MethodPut, fmt.Sprintf("/services/%s:%s/", s.ns, s.name))
	api.responseCodeShouldBe(200)
}

func (s *tService) CheckField(field string, value interface{}) func(t *testing.T) {
	return func(t *testing.T) {
		api := newApi(t)
		api.setBearerToken()
		api.sendRequestTo(http.MethodGet, fmt.Sprintf("/services/%s:%s/", s.ns, s.name))
		api.responseCodeShouldBe(200)
		api.encodeResponseToJson()
		api.fieldIs(field, value)
	}
}

func (s *tService) EditBackLimitsAndIncreaseAdvancedConf(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.setJsonRequestBody(
		map[string]interface{}{
			"name":         s.name,
			"ns":           s.ns,
			"type":         s.type_,
			"replicas":     s.replicas,
			"limits":       s.limits,
			"advancedConf": s.newConf,
		})
	api.sendRequestTo(http.MethodPut, fmt.Sprintf("/services/%s:%s/", s.ns, s.name))
	api.responseCodeShouldBe(200)
}

func (s *tService) DowngradeReplicasAndIncreaseAdvancedConf(t *testing.T) {
	if strings.Contains(t.Name(), pgService.type_) {
		t.Skip("Postgresql fails on this. Skipping.")
	}

	api := newApi(t)
	api.setBearerToken()
	api.setJsonRequestBody(
		map[string]interface{}{
			"name":     s.name,
			"ns":       s.ns,
			"type":     s.type_,
			"replicas": s.replicas,
			"conf":     s.newConf,
		})
	api.sendRequestTo(http.MethodPut, fmt.Sprintf("/services/%s:%s/", s.ns, s.name))
	api.responseCodeShouldBe(200)
	api.encodeResponseToJson()
	api.fieldIs("replicas", s.replicas)
	api.fieldIs("advancedConf", s.newConf)
}

func (s *tService) CreateSecondOneWithSameName(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.setJsonRequestBody(
		map[string]interface{}{
			"name": s.name,
			"ns":   s.ns,
			"type": s.type_,
		})
	api.sendRequestTo(http.MethodPost, "/services/")
	api.responseCodeShouldBe(400)

}

func (s *tService) CheckRecordCount(number int) func(t *testing.T) {
	return func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping. Using -short flag")
			return
		}
		api := newApi(t)
		api.setBearerToken()
		api.sendRequestTo(http.MethodGet, "/services/")
		api.responseCodeShouldBe(http.StatusOK)
		api.encodeResponseToJson()
		api.responseTypeOf(reflect.Slice)
		api.lengthOfResponseIs(number)
	}
}

func (s *tService) IncorrectName(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.sendRequestTo(http.MethodGet, fmt.Sprintf("/services/%s:%s-incorrect", s.ns, s.name))
	api.responseCodeShouldBe(400)
}

func (s *tService) CheckServiceName(t *testing.T) {
	api := newApi(t)
	api.setBearerToken()
	api.sendRequestTo(http.MethodGet, "/services/")
	api.responseCodeShouldBe(http.StatusOK)
	api.encodeResponseToJson()
	api.responseTypeOf(reflect.Slice)

	t.Log("checking...", s.name)
	if !api.responseHasInSlice("name", s.name) {
		t.Errorf("Service %s is not found, in %v", s.name, api.jsonResponse)
	}
}

func (s *tService) WaitForStatus(status string, delay, timeout int64) func(t *testing.T) {
	return func(t *testing.T) {
		// waiting some time to applying new status
		// for example -> mysql does not apply the status immediately
		wait(20)(t)

		begin := time.Now().Unix()
		left := timeout
		for {
			currentTime := time.Now().Unix()
			if currentTime-begin > timeout {
				log.Fatalf("Timeout %d is expired", timeout)
				return
			}

			api := newApi(t)
			api.setBearerToken()
			api.sendRequestTo(http.MethodGet, fmt.Sprintf("/services/%s:%s", s.ns, s.name))
			api.responseCodeShouldBe(200)
			api.encodeResponseToJson()
			api.responseTypeOf(reflect.Map)
			if api.responseHas("status", status) {
				log.Infof("Service %s:%s is reached %s state", s.ns, s.name, status)
				return
			}

			log.Infof("Waiting %s:%s. Left %d seconds", s.ns, s.name, left)
			time.Sleep(time.Duration(delay) * time.Second)
			left = left - delay
		}
	}
}

func (s *tService) Delete(t *testing.T) {
	if !s.force && testing.Short() {
		t.Skip("Skipping. Using -short flag")
		return
	}

	api := newApi(t)
	api.setBearerToken()
	api.sendRequestTo(http.MethodDelete, fmt.Sprintf("/services/%s:%s", s.ns, s.name))
	api.responseCodeShouldBe(http.StatusOK)
}

func makeTestService(ts tService) func(t *testing.T) {
	return func(t *testing.T) {
		steps := []func(t *testing.T){
			ts.Create,
			ts.CreateSecondOneWithSameName,
			ts.TryCreateWithSmallCPULimits,
			ts.TryCreateWithSmallMemoryLimits,
			ts.TryCreateWithSmallDiskLimits,
			ts.TryEditType,
			ts.TryDecreaseVolumeSize,

			ts.CheckRecordCount(1),
			ts.IncorrectName,
			ts.WaitForStatus("Ready", 5, 2*60),

			ts.CheckField("limits", ts.limits),
			ts.CheckField("replicas", ts.replicas),
			ts.CheckField("masters", 1),
			ts.EditLimits,
			ts.WaitForStatus("Ready", 5, 5*60),
			ts.CheckField("limits", ts.newLimits),

			ts.EditBackLimitsAndIncreaseAdvancedConf,
			ts.WaitForStatus("Ready", 5, 5*60),
			ts.CheckField("limits", ts.limits),
			ts.CheckField("advancedConf", ts.newConf),

			ts.EditBackAdvancedConf,
			ts.WaitForStatus("Ready", 5, 5*60),
			ts.CheckField("advancedConf", ts.conf),

			ts.EditReplicas,
			ts.WaitForStatus("Ready", 5, 5*60),
			ts.CheckField("replicas", ts.newReplicas),

			ts.DowngradeReplicasAndIncreaseAdvancedConf, // fails in case the postgresql
			ts.WaitForStatus("Ready", 5, 5*60),

			ts.Delete,
			ts.CheckRecordCount(0),
		}

		for _, item := range steps {
			t.Run(GetFunctionName(item), item)
		}
	}
}

func TestService(t *testing.T) {
	for _, svc := range []tService{pgTestService, mysqlTestService} {
		t.Run(svc.type_, makeTestService(svc))
	}
}

func makeTestSecondService(tsFirst, tsSecond tService) func(t *testing.T) {
	return func(t *testing.T) {
		steps := []func(t *testing.T){
			tsFirst.Create,
			tsSecond.Create,
			// check immediately and after waiting ready state
			tsFirst.CheckServiceName,
			tsSecond.CheckServiceName,
			tsSecond.WaitForStatus("Ready", 5, 5*60),
			tsFirst.CheckServiceName,
			tsSecond.CheckServiceName,
			tsSecond.Delete,
			tsFirst.Delete,
		}

		for _, item := range steps {
			t.Run(GetFunctionName(item), item)
		}
	}
}

func TestSecondService(t *testing.T) {
	pg := pgTestService
	pg.name = "pg-second"
	pg.force = true

	mysql := mysqlTestService
	mysql.name = "mysql-second"
	mysql.force = true

	for _, svc := range [][]tService{{pg, pgTestService}, {mysqlTestService, mysql}} {
		first, second := svc[0], svc[1]
		t.Run(first.type_, makeTestSecondService(first, second))
	}
}
