package cli

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
)

func TestExplainJson(t *testing.T) {
	input := models.Explain{
		Ingress: &models.ExplainIngress{
			Hosts:        []string{"example.com", "example.org"},
			IngressClass: "kong",
		},
		Pod: &models.ExplainPod{
			Containers: []*models.ExplainPodContainersItems0{
				{
					Name:         "foo",
					RestartCount: util.Int64AsPointer(5),
					Status:       "running",
				},
				{
					Name:         "bar",
					RestartCount: util.Int64AsPointer(30),
					Status:       "failing",
				},
			},
		},
		Pvc: &models.ExplainPvc{
			Phase:        "Bound",
			Size:         "10Gi",
			StorageClass: "standard",
		},
	}

	expectedJson, _ := json.Marshal(input)
	client := makeTestClient(200, input)
	cmd, err := MakeRootCmd(client, nil)
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"service", "explain", "--service_id", "test", "--format", "json"})
	err = cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
	actual, err := ioutil.ReadAll(b)
	if err != nil {
		t.Fatal(err)
	}

	result := &models.Explain{}
	err = json.Unmarshal(actual, result)
	if err != nil {
		t.Fatal(err)
	}

	if reflect.DeepEqual(result, input) {
		t.Errorf("Expected %+v, got %+v", expectedJson, result)
	}
}
