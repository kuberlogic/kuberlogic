/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package cli

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/olekukonko/tablewriter"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
)

func TestServiceSecretsListFormatJson(t *testing.T) {
	// make own http client
	expected := []map[string]interface{}{
		{
			"id":    "a",
			"value": "b",
		},
		{
			"id":    "c",
			"value": "d",
		},
	}
	client := makeTestClient(200, expected)
	cmd, err := MakeRootCmd(client, nil)
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"service", "secrets",
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
		t.Fatalf("expected vs actual: %+v vs %+v", expected, actual)
	}
}

func TestServiceSecretsListFormatYaml(t *testing.T) {
	// make own http client
	expected := []map[string]interface{}{
		{
			"id":    "a",
			"value": "b",
		},
		{
			"id":    "c",
			"value": "d",
		},
	}
	client := makeTestClient(200, expected)
	cmd, err := MakeRootCmd(client, nil)
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"service", "secrets",
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
	actual := new(models.ServiceSecrets)
	err = yaml.Unmarshal(out, &actual)
	if err != nil {
		t.Fatal(err)
	}
	r, _ := yaml.Marshal(expected)
	if strings.TrimSpace(string(r)) != strings.TrimSpace(string(out)) {
		t.Fatalf("expected vs actual:\n%s \nvs\n\n%s", r, out)
	}
}

func TestServiceSecretsListFormatStr(t *testing.T) {
	// make own http client
	expected := []map[string]interface{}{
		{
			"id":    "a",
			"value": "b",
		},
		{
			"id":    "c",
			"value": "d",
		},
	}
	client := makeTestClient(200, expected)
	cmd, err := MakeRootCmd(client, nil)
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"service", "secrets"})
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
	table.SetHeader([]string{"ID", "Value"})
	table.SetBorder(false)
	for _, item := range expected {
		table.Append([]string{
			item["id"].(string),
			item["value"].(string),
		})
	}
	table.Render()

	if strings.TrimSpace(string(out)) != strings.TrimSpace(buff.String()) {
		t.Fatalf("expected vs actual: %s vs %s", buff.String(), out)
	}
}
