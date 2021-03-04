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

type TestLogs struct{}

func (td *TestLogs) Get(ns, name, instance string, tail int) func(t *testing.T) {
	return func(t *testing.T) {
		api := newApi(t)
		api.setBearerToken()
		api.query = &url.Values{
			"service_instance": []string{instance},
			"tail":             []string{strconv.Itoa(tail)},
		}
		api.sendRequestTo(http.MethodGet, fmt.Sprintf("/services/%s:%s/logs", ns, name))
		api.responseCodeShouldBe(200)
		api.encodeResponseToJson()
		api.responseTypeOf(reflect.Map)

		resp := api.jsonResponse.(map[string]interface{})
		//t.Log(resp)

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

func Logs(t *testing.T, ns, name, type_, instance string) {
	logs := TestLogs{}
	ts := tService{ns: ns, name: name, type_: type_, force: false, replicas: 1}
	steps := []func(t *testing.T){
		ts.Create,
		ts.WaitForStatus("Ready", 5, 2*60),
		//wait(10 * 60),
		logs.Get(ns, name, instance, 10),
		logs.Get(ns, name, instance, 50),
		logs.Get(ns, name, instance, 100),
		ts.Delete,
	}

	for _, item := range steps {
		t.Run(GetFunctionName(item), item)
	}

}

func TestLogsPg(t *testing.T) {
	// TODO: rewrite it with receive service_instance
	Logs(t, pgService.ns, pgService.name, pgService.type_, fmt.Sprintf("kuberlogic-%s-0", pgService.name))
}

func TestLogsMysql(t *testing.T) {
	Logs(t, mysqlService.ns, mysqlService.name, mysqlService.type_, fmt.Sprintf("%s-mysql-0", mysqlService.name))
}
