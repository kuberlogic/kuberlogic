/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package cli

import (
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"strings"
)

type WithPayload interface {
	GetPayload() *models.Error
}

func getString(cmd *cobra.Command, flag string) (value *string, err error) {
	if cmd.Flags().Changed(flag) {
		value, err := cmd.Flags().GetString(flag)
		if err != nil {
			return nil, err
		}
		return &value, nil
	}
	return
}

func getBool(cmd *cobra.Command, flag string) (value *bool, err error) {
	if cmd.Flags().Changed(flag) {
		value, err := cmd.Flags().GetBool(flag)
		if err != nil {
			return nil, err
		}
		return &value, nil
	}
	return
}

func setInt64(cmd *cobra.Command, flag string) (value *int64, err error) {
	if cmd.Flags().Changed(flag) {
		value, err := cmd.Flags().GetInt64(flag)
		if err != nil {
			return nil, err
		}
		return &value, nil
	}
	return
}

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

func humanizeError(err error) error {
	response, ok := err.(WithPayload)
	if ok {
		e := response.GetPayload()
		if e != nil {
			return errors.New(e.Message)
		}
	}
	return err
}

func getSelectPrompt(cmd *cobra.Command, parameter, defaultValue string, items []string) (string, error) {
	if len(items) == 0 {
		return "", errors.New("no items found")
	}

	flag := cmd.Flag(parameter)

	nonInteractive, err := getBool(cmd, "non-interactive")
	if err != nil {
		return "", err
	}
	if nonInteractive == nil {
		currentPos := 0
		for ix, i := range items {
			if defaultValue == i {
				currentPos = ix
			}
		}

		prompt := promptui.Select{
			Label:     flag.Usage,
			Items:     items,
			CursorPos: currentPos,
			Size:      len(items),
		}
		_, result, err := prompt.Run()
		return result, err
	}
	val, err := getString(cmd, parameter)
	if err != nil {
		return "", err
	}
	if val == nil {
		val = &defaultValue
	}
	for _, item := range items {
		if item == *val {
			return *val, nil
		}
	}

	return "", errors.New(*val + " is not available. Available: " + strings.Join(items, ","))
}

func getStringPrompt(cmd *cobra.Command, parameter, defaultValue string, validatef promptui.ValidateFunc) (string, error) {
	if validatef == nil {
		validatef = func(s string) error {
			return nil
		}
	}

	flag := cmd.Flag(parameter)

	nonInteractive, err := getBool(cmd, "non-interactive")
	if err != nil {
		return "", err
	}
	if nonInteractive == nil {
		prompt := promptui.Prompt{
			Label:    flag.Usage,
			Validate: validatef,
			Default:  defaultValue,
		}
		result, err := prompt.Run()
		return result, err
	}
	val, err := getString(cmd, parameter)
	if err != nil {
		return "", err
	}
	if val != nil {
		return *val, validatef(*val)
	}
	return defaultValue, validatef(defaultValue)
}

func getBoolPrompt(cmd *cobra.Command, defaultValue bool, parameter string) (bool, error) {
	flag := cmd.Flag(parameter)

	nonInteractive, err := getBool(cmd, "non-interactive")
	if err != nil {
		return false, err
	}
	if nonInteractive == nil {
		promptDefault := "N"
		suffix := " [y/N]"
		if defaultValue {
			promptDefault = "y"
			suffix = " [Y/n]"
		}

		prompt := promptui.Prompt{
			Label:   flag.Usage + suffix,
			Default: promptDefault,
			Validate: func(s string) error {
				if lower := strings.ToLower(s); lower != "y" && lower != "n" {
					return errors.New("Please answer with [y]es or [n]o")
				}
				return nil
			},
		}
		result, err := prompt.Run()
		isTrue := strings.ToLower(result) == "y"
		return isTrue, err
	}
	val, err := getBool(cmd, parameter)
	if val != nil {
		return *val, err
	}
	return defaultValue, err
}
