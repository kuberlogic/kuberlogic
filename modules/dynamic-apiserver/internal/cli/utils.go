/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package cli

import (
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func printResult(cmd *cobra.Command, formatResponse format, payload interface{}) error {
	var result []byte
	var err error
	switch formatResponse {
	case jsonFormat:
		result, err = json.MarshalIndent(payload, "", "\t")
	case yamlFormat:
		result, err = yaml.Marshal(payload)
	default:
		return errors.Errorf("undefined format: %v", formatResponse)
	}

	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", result)
	if err != nil {
		return err
	}
	return nil
}

func isDefaultPrintFormat(formatResponse format) bool {
	return formatResponse == "" || formatResponse == stringFormat
}
