package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kuberlogic/kuberlogic/modules/apiserver/internal/generated/models"
	"github.com/kuberlogic/kuberlogic/modules/operator/service-operator/interfaces"
	"github.com/kuberlogic/kuberlogic/modules/operator/service-operator/util/kuberlogic"
	"github.com/prometheus/common/log"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

type baseUrl struct {
	scheme string
	host   string
	port   int
	base   string
}

func (bu *baseUrl) buildUrl(endpoint string) string {
	endpoint = strings.TrimSuffix(endpoint, "/")
	endpoint = strings.TrimPrefix(endpoint, "/")
	return fmt.Sprintf("%s://%s:%d%s%s/", bu.scheme, bu.host, bu.port, bu.base, endpoint)
}

type API struct {
	baseUrl      baseUrl
	request      *http.Request
	response     *http.Response
	jsonResponse interface{}
	jsonRequest  string
	token        string
	query        *url.Values
	t            *testing.T
}

func (a *API) getAuthToken(user, password string) (string, error) {
	loginData := &models.UserCredentials{
		Password: &password,
		Username: &user,
	}
	data, err := json.Marshal(loginData)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, a.baseUrl.buildUrl("/login"), bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := newHttpClient()
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	authData := new(models.AccessTokenResponse)
	if err := json.Unmarshal([]byte(resBody), authData); err != nil {
		return "", err
	}

	if authData.AccessToken == "" {
		return "", fmt.Errorf("access token should not be empty after login")
	}
	return "Bearer " + authData.AccessToken, nil
}

func (a *API) setBearerToken() {
	token, err := a.getAuthToken(testUser, testPassword)
	if err != nil || token == "" {
		a.t.Errorf("error getting acess token")
	}
	a.token = token
}

func (a *API) sendRequestTo(method, endpoint string) {
	req, err := http.NewRequest(method, a.baseUrl.buildUrl(endpoint), bytes.NewBuffer([]byte(a.jsonRequest)))
	if err != nil {
		a.t.Error(err)
	}
	if a.query != nil {
		req.URL.RawQuery = a.query.Encode()
	}

	// handle panic
	defer func() {
		switch t := recover().(type) {
		case string:
			a.t.Errorf(t)
		case error:
			a.t.Error(err)
		}
	}()

	req.Header.Set("Content-Type", "application/json")
	if a.token != "" {
		req.Header.Add("Authorization", a.token)
	}

	client := newHttpClient()
	resp, err := client.Do(req)
	if err != nil {
		a.t.Error(err)
	}
	a.response = resp
	return
}

func (a *API) responseCodeShouldBe(code int) {
	if code != a.response.StatusCode {
		a.t.Errorf("expected response code to be: %d, but actual is: %d", code, a.response.StatusCode)
	}
}

func (a *API) responseShouldMatchJson(body string) {
	var expected interface{}

	// re-encode expected response
	if err := json.Unmarshal([]byte(body), &expected); err != nil {
		a.t.Error(err)
	}

	// the matching may be adapted per different requirements.
	if !reflect.DeepEqual(expected, a.jsonResponse) {
		a.t.Errorf("expected JSON does not match actual, %v vs. %v", expected, a.jsonResponse)
	}
}

func (a *API) responseShouldHas(field, value interface{}) {
	if !a.responseHas(field, value) {
		a.t.Errorf("expected field (value) does not match actual, %v (%v) vs. %v", field, value, a.jsonResponse)
	}
}

func (a *API) responseHas(field, value interface{}) bool {
	iter := reflect.ValueOf(a.jsonResponse).MapRange()
	for iter.Next() {
		k, v := iter.Key().Interface(), iter.Value().Interface()
		if reflect.DeepEqual(k, field) && reflect.DeepEqual(v, value) {
			return true
		}
	}
	return false
}

func (a *API) responseHasInSlice(field, value interface{}) bool {
	array := a.jsonResponse.([]interface{})
	for _, item := range array {
		iter := reflect.ValueOf(item).MapRange()
		for iter.Next() {
			k, v := iter.Key().Interface(), iter.Value().Interface()
			if reflect.DeepEqual(k, field) && reflect.DeepEqual(v, value) {
				return true
			}
		}
	}
	return false
}

func (a *API) lengthOfResponseIs(expected int) {
	val := reflect.ValueOf(a.jsonResponse)
	if val.Len() != expected {
		a.t.Errorf("expected length is %d, but got: %d", expected, val.Len())
	}
}

func (a *API) encodeResponseToJson() {
	var actual interface{}
	a.encodeResponseTo(&actual)
	a.jsonResponse = actual
}

func (a *API) encodeResponseTo(actual interface{}) {
	result, err := ioutil.ReadAll(a.response.Body)
	if err != nil {
		a.t.Error(err)
	}
	if err = json.Unmarshal(result, &actual); err != nil {
		a.t.Error(err)
	}
	a.jsonResponse = actual
}

func (a *API) responseTypeOf(kind reflect.Kind) {
	val := reflect.ValueOf(a.jsonResponse)
	if val.Kind() != kind {
		a.t.Errorf("expected %s, but got: %s", kind, val.Kind())
	}
}

func (a *API) fieldContains(field, value string) {
	resp := a.jsonResponse.(map[string]interface{})
	res, ok := resp[field]
	if !ok {
		a.t.Errorf("%s field does not exist", field)
	}
	message, ok := res.(string)
	if !ok {
		a.t.Errorf("%s is not string type", field)
	}
	if !strings.Contains(message, value) {
		a.t.Errorf(`expected "%s", but got: "%s"`, value, res)
	}
}

func (a *API) fieldIs(field string, value interface{}) {
	resp := a.jsonResponse.(map[string]interface{})
	res, ok := resp[field]
	if !ok {
		a.t.Errorf("%s field does not exist", field)
	}

	if reflect.DeepEqual(res, value) {
		a.t.Errorf(`expected "%f", but got: "%f"`, value, res)
	}
}

func (a *API) setRequestBody(body string) {
	a.jsonRequest = body
}

func (a *API) setJsonRequestBody(body interface{}) {
	a.jsonRequest = toJson(body)
}

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func wait(seconds int) func(t *testing.T) {
	return func(t *testing.T) {
		log.Infof("Waiting for %d seconds", seconds)
		time.Sleep(time.Duration(seconds) * time.Second)
	}
}

func toJson(v interface{}) string {
	res, _ := json.Marshal(v)
	return string(res)
}

func GetSession(ns, serviceName, db string) (session interfaces.Session, err error) {
	client, resource, err := Connect(ns, serviceName)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to the k8s resource: %s", err)
	}

	session, err = kuberlogic.GetSession(resource, client, db)
	if err != nil {
		return nil, fmt.Errorf("cannot get session:%s", err)
	}
	return session, nil
}

func newApi(t *testing.T) *API {
	return &API{
		t: t,
		baseUrl: baseUrl{
			scheme: "http",
			host:   apiHost,
			port:   apiPort,
			base:   "/api/v1/",
		},
	}
}
