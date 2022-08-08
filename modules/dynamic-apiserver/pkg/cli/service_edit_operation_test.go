/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package cli

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"
)

func TestEditEmptyID(t *testing.T) {
	// make own http client
	client := makeTestClient(400, map[string]string{})

	cmd, err := MakeRootCmd(client, nil)
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"service", "edit"})
	err = cmd.Execute()
	expected := "ID is not specified"
	if err != nil && err.Error() != expected {
		t.Fatalf("expected vs actual: %v vs %v", expected, err.Error())
	}
}

func TestEditNotFound(t *testing.T) {
	// make own http client
	expected := "kuberlogic service not found: test"
	client := makeTestClient(404, map[string]string{
		"message": expected,
	})

	cmd, err := MakeRootCmd(client, nil)
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"service", "edit",
		"--id", "test",
	})
	err = cmd.Execute()
	if err != nil && err.Error() != expected {
		t.Fatalf("expected vs actual: %v vs %v", expected, err.Error())
	}
}

func TestEditSuccessFormatJson(t *testing.T) {
	// make own http client
	expected := map[string]interface{}{
		"id":         "test",
		"type":       "some-type",
		"domain":     "new-domain",
		"created_at": "0001-01-01T00:00:00.000Z",
	}
	client := makeTestClient(200, expected)
	cmd, err := MakeRootCmd(client, nil)
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"service", "edit",
		"--id", "test",
		"--format", "json",
		"--domain", "new-domain",
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

func TestEditFailType(t *testing.T) {
	// make own http client
	expected := map[string]interface{}{
		"id":         "test",
		"type":       "some-type",
		"domain":     "new-domain",
		"created_at": "0001-01-01T00:00:00.000Z",
	}
	client := makeTestClient(200, expected)
	cmd, err := MakeRootCmd(client, nil)
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"service", "edit",
		"--id", "test",
		"--format", "json",
		"--type", "new-type",
		"--domain", "new-domain",
	})
	err = cmd.Execute()
	if err != nil && err.Error() == "unknown flag: --type" {
		return
	}
	t.Fatal("error has not returned")
}
