package cli

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
)

func TestListLogs(t *testing.T) {
	input := models.Logs{
		&models.Log{
			ContainerName: "test",
			Logs:          "successfully started container...",
		},
	}
	expected := fmt.Sprintf("%s:\t%s", input[0].ContainerName, input[0].Logs)
	client := makeTestClient(200, input)
	cmd, err := MakeRootCmd(client, nil)
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"service", "logs", "--service_id", "test"})
	err = cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
	actual, err := ioutil.ReadAll(b)
	if err != nil {
		t.Fatal(err)
	}

	if strings.TrimSpace(expected) != strings.TrimSpace(string(actual)) {
		t.Fatalf("expected vs actual: %s vs %s", expected, actual)
	}
}
