/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/internal/generated/models"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
)

func TestDeleteInvalidValidation(t *testing.T) {
	// make own http client
	client := makeTestClient(400, map[string]string{
		"message": "id can't be empty",
	})

	cmd, err := MakeRootCmd(client)
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"service", "delete"})
	err = cmd.Execute()
	expected := "id can't be empty"
	if err != nil && err.Error() != expected {
		t.Fatalf("expected vs actual: %v vs %v", expected, err.Error())
	}
}

func TestDeleteNotFound(t *testing.T) {
	// make own http client
	expected := "Record not found with id 'test'"
	client := makeTestClient(404, map[string]string{
		"message": expected,
	})

	cmd, err := MakeRootCmd(client)
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"service", "delete",
		"--id", "test",
	})
	err = cmd.Execute()
	if err != nil && err.Error() != expected {
		t.Fatalf("expected vs actual: %v vs %v", expected, err.Error())
	}
}

func TestDeleteSuccessFormatJson(t *testing.T) {
	// make own http client
	expected := map[string]interface{}{}
	client := makeTestClient(200, expected)
	cmd, err := MakeRootCmd(client)
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	//cmd.SetErr(b)
	cmd.SetArgs([]string{"service", "delete",
		"--id", "test",
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

func TestDeleteSuccessFormatYaml(t *testing.T) {
	// make own http client
	expected := map[string]interface{}{}
	client := makeTestClient(200, expected)
	cmd, err := MakeRootCmd(client)
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	//cmd.SetErr(b)
	cmd.SetArgs([]string{"service", "delete",
		"--id", "test",
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

func TestDeleteSuccessFormatStr(t *testing.T) {
	// make own http client
	expected := map[string]interface{}{}
	client := makeTestClient(200, expected)
	cmd, err := MakeRootCmd(client)
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	//cmd.SetErr(b)
	id := "test"
	cmd.SetArgs([]string{"service", "delete",
		"--id", id,
	})
	err = cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
	out, err := ioutil.ReadAll(b)
	if err != nil {
		t.Fatal(err)
	}
	expectedResult := fmt.Sprintf("Service '%s' successfully removed", id)
	if strings.TrimSpace(string(out)) != expectedResult {
		t.Fatalf("expected vs actual: %s vs %s", expectedResult, out)
	}
}
