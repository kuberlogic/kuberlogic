package cli

import (
	"archive/zip"
	"bytes"
	"io"
	k8stesting "k8s.io/client-go/kubernetes/fake"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"
)

var (
	rgxp, _ = regexp.Compile(`kuberlogic-diag-\d+\.zip`)
)

func TestDiagSuccess(t *testing.T) {
	configDir, _ := os.MkdirTemp("", "install-test")
	defer os.RemoveAll(configDir)

	kubectlBin = "echo"

	cmd, err := MakeRootCmd(nil, k8stesting.NewSimpleClientset())
	if err != nil {
		t.Fatal(err)
	}
	cmd.SetArgs([]string{"diag"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	// search for zip files in current directory
	dirFiles, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range dirFiles {
		if rgxp.MatchString(f.Name()) {
			// zip archive should be not nil and contain diag data
			// not a leak!
			defer os.Remove(f.Name())

			f, err := os.ReadFile(f.Name())
			if err != nil {
				t.Fatal(err)
			}
			z, err := zip.NewReader(bytes.NewReader(f), int64(len(f)))
			if err != nil {
				t.Fatal(err)
			}

			// status file mut be present and not empty
			statusFileHandler, err := z.Open("pods-status.yaml")
			if err != nil {
				t.Fatal(err)
			}
			statusFileContent := &bytes.Buffer{}
			if _, err := io.Copy(statusFileContent, statusFileHandler); err != nil {
				t.Fatal(err)
			}
			expected := "get pods --all-namespaces -o yaml\n"
			if actual := statusFileContent.String(); expected != actual {
				t.Fatalf("actual `%s` does not match expected `%s`", actual, expected)
			}

			// no error files should be found
			for _, f := range z.File {
				if strings.Contains(f.Name, "error") {
					t.Fatalf("error files should not be preset, instead found %s", f.Name)
				}
			}
			return
		}
	}

	t.Fatal("failed to find diag archive")
}

func TestDiagSuccessFailure(t *testing.T) {
	configDir, _ := os.MkdirTemp("", "install-test")
	defer os.RemoveAll(configDir)

	// we need to be sure that there is no race condition between success and failure tests
	time.Sleep(time.Second)
	kubectlBin = "ech"

	cmd, err := MakeRootCmd(nil, k8stesting.NewSimpleClientset())
	if err != nil {
		t.Fatal(err)
	}
	cmd.SetArgs([]string{"diag"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	// search for zip files in current directory
	dirFiles, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range dirFiles {
		if rgxp.MatchString(f.Name()) {
			// zip archive should be not nil and contain diag data
			// not a leak!
			defer os.Remove(f.Name())

			f, err := os.ReadFile(f.Name())
			if err != nil {
				t.Fatal(err)
			}
			z, err := zip.NewReader(bytes.NewReader(f), int64(len(f)))
			if err != nil {
				t.Fatal(err)
			}

			// status file mut be present and not empty
			statusFileHandler, err := z.Open("pods-status.yaml-error")
			if err != nil {
				t.Fatal(err)
			}
			statusFileContent := &bytes.Buffer{}
			if _, err := io.Copy(statusFileContent, statusFileHandler); err != nil {
				t.Fatal(err)
			}
			expected := "exec: \"ech\": executable file not found in $PATH"
			if actual := statusFileContent.String(); expected != actual {
				t.Fatalf("actual `%s` does not match expected `%s`", actual, expected)
			}

			return
		}
	}

	t.Fatal("failed to find diag archive")
}
