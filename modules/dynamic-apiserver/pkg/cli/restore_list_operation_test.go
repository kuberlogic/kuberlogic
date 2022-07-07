/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package cli

import (
	"bytes"
	"encoding/json"
	"github.com/ghodss/yaml"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	"github.com/olekukonko/tablewriter"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestRestoreListFormatJson(t *testing.T) {
	// make own http client
	expected := []map[string]interface{}{
		{
			"created_at": "2022-05-10T16:00:53.000Z",
			"id":         "test-1",
			"status":     "Unknown",
			"backup_id":  "test-1",
		},
		{
			"created_at": "2022-05-10T16:00:53.000Z",
			"id":         "test-2",
			"status":     "Unknown",
			"backup_id":  "test-2",
		},
	}
	client := makeTestClient(200, expected)
	cmd, err := MakeRootCmd(client)
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"restore", "list",
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

func TestRestoreListEmptyFormatJson(t *testing.T) {
	// make own http client
	var expected []map[string]interface{}
	client := makeTestClient(200, expected)
	cmd, err := MakeRootCmd(client)
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"restore", "list",
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

func TestRestoreListFormatYaml(t *testing.T) {
	// make own http client
	expected := []map[string]interface{}{
		{
			"created_at": "2022-05-10T16:00:53.000Z",
			"id":         "test-1",
			"status":     "Unknown",
			"backup_id":  "test-1",
		},
		{
			"created_at": "2022-05-10T16:00:53.000Z",
			"id":         "test-2",
			"status":     "Unknown",
			"backup_id":  "test-2",
		},
	}
	client := makeTestClient(200, expected)
	cmd, err := MakeRootCmd(client)
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"restore", "list",
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

func TestRestoreListFormatStr(t *testing.T) {
	// make own http client
	expected := []map[string]interface{}{
		{
			"created_at": "2022-05-10T16:00:53.000Z",
			"id":         "test-1",
			"status":     "Unknown",
			"backup_id":  "test-1",
		},
		{
			"created_at": "2022-05-10T16:00:53.000Z",
			"id":         "test-2",
			"status":     "Unknown",
			"backup_id":  "test-2",
		},
	}
	client := makeTestClient(200, expected)
	cmd, err := MakeRootCmd(client)
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"restore", "list"})
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
	table.SetHeader([]string{"№", "ID", "Backup ID", "Created", "Status"})
	table.SetBorder(false)
	for i, item := range expected {
		table.Append([]string{
			strconv.Itoa(i),
			item["id"].(string),
			item["backup_id"].(string),
			item["created_at"].(string),
			item["status"].(string),
		})
	}
	table.Render()

	if strings.TrimSpace(string(out)) != strings.TrimSpace(buff.String()) {
		t.Fatalf("expected vs actual: %s vs %s", buff.String(), out)
	}
}
