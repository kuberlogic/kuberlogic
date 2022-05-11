/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package cli

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func Test_InvalidValidationCommand(t *testing.T) {
	cmd, err := MakeRootCmd()
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"service", "add"})
	err = cmd.Execute()
	expected := "validation error: name in body is required"
	if err != nil && err.Error() != expected {
		t.Fatalf("expected vs actual: %v vs %v", expected, err.Error())
	}
}

func Test_ExecuteCommand(t *testing.T) {
	cmd, err := MakeRootCmd()
	if err != nil {
		t.Fatal(err)
	}
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"service", "add"})
	err = cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
	out, err := ioutil.ReadAll(b)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "testisawesome" {
		t.Fatalf("expected \"%s\" got \"%s\"", "testisawesome", string(out))
	}
}
