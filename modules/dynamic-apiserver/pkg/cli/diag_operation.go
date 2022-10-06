package cli

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func makeDiagCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diag",
		Short: "Gather diagnostic information about Kuberlogic installation",
		RunE:  runDiag(),
	}

	return cmd
}

func runDiag() func(command *cobra.Command, args []string) error {
	return func(command *cobra.Command, args []string) error {

		buf := &bytes.Buffer{}
		zipWriter := zip.NewWriter(buf)

		var cmd string
		command.Println("Gathering pods status")
		cmd = fmt.Sprintf("%s get pods --all-namespaces -o yaml", kubectlBin)
		if err := saveDiagInfo(cmd, "pods-status.yaml", zipWriter); err != nil {
			return errors.Wrapf(err, "failed to save %s info", cmd)
		}

		command.Println("Gathering Kuberlogic pods logs")
		cmd = fmt.Sprintf("%s -n kuberlogic logs -l control-plane=controller-manager --all-containers=true", kubectlBin)
		if err := saveDiagInfo(cmd, "kuberlogic-logs.txt", zipWriter); err != nil {
			return errors.Wrapf(err, "failed to save %s info", cmd)
		}

		command.Println("Gathering Kuberlogic configuration information")
		cmd = fmt.Sprintf("%s -n kuberlogic get secrets -l kl-config=true -o yaml", kubectlBin)
		if err := saveDiagInfo(cmd, "kuberlogic-configs.yaml", zipWriter); err != nil {
			return errors.Wrapf(err, "failed to save %s info", cmd)
		}

		command.Println("Gathering Kuberlogic endpoints info")
		cmd = fmt.Sprintf("%s -n kuberlogic get ep,svc -o yaml", kubectlBin)
		if err := saveDiagInfo(cmd, "kuberlogic-endpoints.yaml", zipWriter); err != nil {
			return errors.Wrapf(err, "failed to save %s info", cmd)
		}

		command.Println("Gathering Kuberlogic resources status")
		cmd = fmt.Sprintf("%s get kuberlogic -o yaml", kubectlBin)
		if err := saveDiagInfo(cmd, "kuberlogic-resources.yaml", zipWriter); err != nil {
			return errors.Wrapf(err, "failed to save %s info", cmd)
		}

		// save to diag archive file
		diagArchiveName := fmt.Sprintf("kuberlogic-diag-%d.zip", time.Now().Unix())
		diagArchive, err := os.Create(diagArchiveName)
		if err != nil {
			return errors.Wrapf(err, "failed to create diagnostics archive %s", diagArchiveName)
		}
		defer diagArchive.Close()

		zipWriter.Close()
		if _, err := io.Copy(diagArchive, buf); err != nil {
			return errors.Wrapf(err, "failed to save diagnostics archive data")
		}

		command.Println("Diagnostics information is saved into " + diagArchiveName)
		return nil
	}
}

func saveDiagInfo(command, fname string, archive *zip.Writer) error {
	cmd := strings.Split(command, " ")
	if len(cmd) < 2 {
		return errors.New("failed to parse diag command")
	}
	cmdOut, cmdErr := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
	if cmdErr != nil {
		w, _ := archive.Create(fname + "-error")
		_, _ = io.Copy(w, strings.NewReader(cmdErr.Error()))
	}

	w, err := archive.Create(fname)
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, bytes.NewReader(cmdOut)); err != nil {
		return err
	}
	return nil
}
