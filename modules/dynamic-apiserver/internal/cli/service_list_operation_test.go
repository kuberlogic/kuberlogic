/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package cli

import (
	"bytes"
	"encoding/json"
	"github.com/ghodss/yaml"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/internal/generated/models"
	"github.com/olekukonko/tablewriter"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestListFormatJson(t *testing.T) {
	// make own http client
	expected := []map[string]interface{}{
		{
			"created_at": "2022-05-10T16:00:53.000Z",
			"limits": map[string]interface{}{
				"cpu":    "250m",
				"memory": "256Mi",
			},
			"id":         "test-1",
			"replicas":   float64(0),
			"status":     "Unknown",
			"type":       "postgresql",
			"version":    "13",
			"volumeSize": "1Gi",
		},
		{
			"created_at": "2022-05-10T16:00:53.000Z",
			"limits": map[string]interface{}{
				"cpu":    "250m",
				"memory": "256Mi",
			},
			"id":         "test-2",
			"replicas":   float64(0),
			"status":     "Unknown",
			"type":       "postgresql",
			"version":    "13",
			"volumeSize": "1Gi",
		},
	}
	client := makeTestClient(200, expected)
	cmd, err := MakeRootCmd(client)
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"service", "list",
		"--format", "json",
	})
	err = cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
	out, err := ioutil.ReadAll(b)
	if err != nil {
		t.Fatal(err)
	}
	var actual []map[string]interface{}
	err = json.Unmarshal(out, &actual)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("expected vs actual: %s vs %s", expected, actual)
	}
}

func TestListEmptyFormatJson(t *testing.T) {
	// make own http client
	var expected []map[string]interface{}
	client := makeTestClient(200, expected)
	cmd, err := MakeRootCmd(client)
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"service", "list",
		"--format", "json",
	})
	err = cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
	out, err := ioutil.ReadAll(b)
	if err != nil {
		t.Fatal(err)
	}
	var actual []map[string]interface{}
	err = json.Unmarshal(out, &actual)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("expected vs actual: %s vs %s", expected, actual)
	}
}

func TestListFormatYaml(t *testing.T) {
	// make own http client
	expected := []map[string]interface{}{
		{
			"created_at": "2022-05-10T16:00:53.000Z",
			"limits": map[string]interface{}{
				"cpu":    "250m",
				"memory": "256Mi",
			},
			"id":         "test-1",
			"replicas":   float64(0),
			"status":     "Unknown",
			"type":       "postgresql",
			"version":    "13",
			"volumeSize": "1Gi",
		},
		{
			"created_at": "2022-05-10T16:00:53.000Z",
			"limits": map[string]interface{}{
				"cpu":    "250m",
				"memory": "256Mi",
			},
			"id":         "test-2",
			"replicas":   float64(0),
			"status":     "Unknown",
			"type":       "postgresql",
			"version":    "13",
			"volumeSize": "1Gi",
		},
	}
	client := makeTestClient(200, expected)
	cmd, err := MakeRootCmd(client)
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"service", "list",
		"--format", "yaml",
	})
	err = cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
	out, err := ioutil.ReadAll(b)
	if err != nil {
		t.Fatal(err)
	}
	actual := new(models.Services)
	err = yaml.Unmarshal(out, &actual)
	if err != nil {
		t.Fatal(err)
	}
	r, _ := yaml.Marshal(expected)
	if strings.TrimSpace(string(r)) != strings.TrimSpace(string(out)) {
		t.Fatalf("expected vs actual:\n%s \nvs\n\n%s", r, out)
	}
}

func TestListFormatStr(t *testing.T) {
	// make own http client
	expected := []map[string]interface{}{
		{
			"created_at": "2022-05-10T16:00:53.000Z",
			"limits": map[string]interface{}{
				"cpu":    "250m",
				"memory": "256Mi",
			},
			"id":         "test-1",
			"replicas":   float64(0),
			"status":     "Unknown",
			"type":       "postgresql",
			"version":    "13",
			"volumeSize": "1Gi",
		},
		{
			"created_at": "2022-05-10T16:00:53.000Z",
			"limits": map[string]interface{}{
				"cpu":    "250m",
				"memory": "256Mi",
			},
			"id":         "test-2",
			"replicas":   float64(0),
			"status":     "Unknown",
			"type":       "postgresql",
			"version":    "13",
			"volumeSize": "1Gi",
		},
	}
	client := makeTestClient(200, expected)
	cmd, err := MakeRootCmd(client)
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"service", "list"})
	err = cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
	out, err := ioutil.ReadAll(b)
	if err != nil {
		t.Fatal(err)
	}
	buff := bytes.NewBufferString("")
	table := tablewriter.NewWriter(buff)
	table.SetHeader([]string{"№", "ID", "Type", "Replica", "Version", "Status"})
	table.SetBorder(false)
	for i, item := range expected {
		table.Append([]string{
			strconv.Itoa(i),
			item["id"].(string),
			item["type"].(string),
			strconv.Itoa(int(item["replicas"].(float64))),
			item["version"].(string),
			item["status"].(string),
		})
	}
	table.Render()

	if strings.TrimSpace(string(out)) != strings.TrimSpace(buff.String()) {
		t.Fatalf("expected vs actual: %s vs %s", buff.String(), out)
	}
}
