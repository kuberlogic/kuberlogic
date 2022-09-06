package cli

import (
	"bytes"
	"errors"
	"io/ioutil"
	"testing"
)

func TestCredentialsUpdateOK(t *testing.T) {
	client := makeTestClient(200, nil)
	cmd, err := MakeRootCmd(client, nil)
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"service", "credentials-update",
		"token=secret",
	})
	err = cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
	_, err = ioutil.ReadAll(b)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCredentialsUpdateInvalidData(t *testing.T) {
	client := makeTestClient(200, nil)
	cmd, err := MakeRootCmd(client, nil)
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"service", "credentials-update",
		"token",
	})
	if !errors.Is(cmd.Execute(), errPartialCredentials) {
		t.Fatalf("error is not %v", errPartialCredentials)
	}

	cmd.SetArgs([]string{"service", "credentials-update"})
	if !errors.Is(cmd.Execute(), errCredentialsNotFound) {
		t.Fatalf("error is not %v", errCredentialsNotFound)
	}

	cmd.SetArgs([]string{"service", "credentials-update", "token=test", "token=test2"})
	if !errors.Is(cmd.Execute(), errDuplicateCredentialsPairs) {
		t.Fatalf("error is not %v", errDuplicateCredentialsPairs)
	}
}
