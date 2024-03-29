/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package cli

import (
	"bytes"
	"encoding/json"
	"github.com/ghodss/yaml"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
)

func TestBackupRestoreInvalidValidation(t *testing.T) {
	// make own http client
	client := makeTestClient(422, map[string]string{
		"message": "id in body is required",
	})

	cmd, err := MakeRootCmd(client, nil)
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"backup", "restore"})
	err = cmd.Execute()
	expected := "id in body is required"
	if err != nil && err.Error() != expected {
		t.Fatalf("expected vs actual: %v vs %v", expected, err.Error())
	}
}

func TestBackupRestoreSuccessFormatJson(t *testing.T) {
	// make own http client
	expected := map[string]interface{}{
		"created_at": "2022-05-10T16:00:53.000Z",
		"id":         "test",
		"status":     "Unknown",
		"backup_id":  "test",
	}
	client := makeTestClient(201, expected)
	cmd, err := MakeRootCmd(client, nil)
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	//cmd.SetErr(b)
	cmd.SetArgs([]string{"backup", "restore",
		"--backup_id", "test",
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
	actual := make(map[string]interface{})
	err = json.Unmarshal(out, &actual)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("expected vs actual: %s vs %s", expected, actual)
	}
}

func TestBackupRestoreSuccessFormatYaml(t *testing.T) {
	// make own http client
	expected := map[string]interface{}{
		"created_at": "2022-05-10T16:00:53.000Z",
		"id":         "test",
		"status":     "Unknown",
		"backup_id":  "test",
	}
	client := makeTestClient(201, expected)
	cmd, err := MakeRootCmd(client, nil)
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	//cmd.SetErr(b)
	cmd.SetArgs([]string{"backup", "restore",
		"--backup_id", "test",
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
	actual := new(models.Service)
	err = yaml.Unmarshal(out, &actual)
	if err != nil {
		t.Fatal(err)
	}
	r, _ := yaml.Marshal(expected)
	if strings.TrimSpace(string(r)) != strings.TrimSpace(string(out)) {
		t.Fatalf("expected vs actual:\n%s \nvs\n\n%s", r, out)
	}
}

func TestBackupRestoreSuccessFormatStr(t *testing.T) {
	// make own http client
	expected := map[string]interface{}{
		"created_at": "2022-05-10T16:00:53.000Z",
		"id":         "test",
		"status":     "Unknown",
		"backup_id":  "test",
	}
	client := makeTestClient(201, expected)
	cmd, err := MakeRootCmd(client, nil)
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	//cmd.SetErr(b)
	cmd.SetArgs([]string{"backup", "restore",
		"--backup_id", "test",
	})
	err = cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
	out, err := ioutil.ReadAll(b)
	if err != nil {
		t.Fatal(err)
	}
	expectedResult := "A request 'test' to restore backup 'test' successfully created"
	if strings.TrimSpace(string(out)) != expectedResult {
		t.Fatalf("expected vs actual: %s vs %s", expectedResult, out)
	}
}
