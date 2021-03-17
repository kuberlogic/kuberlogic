package tests

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

type tLogs struct {
	service  tService
	instance string
}

func (tl *tLogs) Get(tail int) func(t *testing.T) {
	return func(t *testing.T) {
		api := newApi(t)
		api.setBearerToken()
		api.query = &url.Values{
			"service_instance": []string{tl.instance},
			"tail":             []string{strconv.Itoa(tail)},
		}
		api.sendRequestTo(http.MethodGet, fmt.Sprintf("/services/%s:%s/logs",
			tl.service.ns, tl.service.name))
		api.responseCodeShouldBe(200)
		api.encodeResponseToJson()
		api.responseTypeOf(reflect.Map)

		resp := api.jsonResponse.(map[string]interface{})

		body, ok := resp["body"]
		if !ok {
			t.Errorf("body field does not exist")
		}
		lines, ok := resp["lines"]
		if !ok {
			t.Errorf("lines field does not exist")
		}
		linesInt, ok := lines.(float64)
		if !ok {
			t.Errorf("lines is not float64 type, %v (%T)", lines, lines)
		}

		if tail != int(linesInt) {
			t.Errorf("expected vs actual, %d != %f", tail, linesInt)
		}

		bodyStr, ok := body.(string)
		if !ok {
			t.Errorf("body is not string type")
		}

		bodyStr = strings.TrimSuffix(bodyStr, "\n") // remove last empty line if exists
		result := strings.Split(bodyStr, "\n")
		// result could be less than required
		if len(result) != tail {
			t.Logf("expected vs actual, %d != %d", len(result), tail)
		}
		if len(result) > tail {
			t.Errorf("expected vs actual, %d > %d", len(result), tail)
		}
		emptyLines := 0
		for _, line := range result {
			if line == "" {
				emptyLines++
			}
		}
		if emptyLines == lines {
			t.Errorf("Log is empty")
		}
	}
}

func makeTestLogs(tlogs tLogs) func(t *testing.T) {
	return func(t *testing.T) {
		steps := []func(t *testing.T){
			tlogs.service.Create,
			tlogs.service.WaitForStatus("Ready", 5, 2*60),

			tlogs.Get(10),
			tlogs.Get(50),
			tlogs.Get(100),
			tlogs.service.Delete,
		}

		for _, item := range steps {
			t.Run(GetFunctionName(item), item)
		}
	}
}

func TestLogs(t *testing.T) {
	for _, svc := range []tLogs{
		{
			service:  pgTestService,
			instance: fmt.Sprintf("kuberlogic-%s-0", pgService.name),
		}, {
			service:  mysqlTestService,
			instance: fmt.Sprintf("%s-mysql-0", mysqlService.name),
		}} {
		t.Run(svc.service.type_, makeTestLogs(svc))
	}
}
