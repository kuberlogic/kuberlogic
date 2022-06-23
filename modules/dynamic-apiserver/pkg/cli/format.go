/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package cli

import (
	"github.com/pkg/errors"
)

type format string

const (
	jsonFormat   format = "json"
	yamlFormat   format = "yaml"
	stringFormat format = "string"
)

// String is used both by fmt.Print and by Cobra in help text
func (f *format) String() string {
	return string(*f)
}

// Set must have pointer receiver so it doesn't change the value of a copy
func (f *format) Set(v string) error {
	switch v {
	case "json", "yaml", "string":
		*f = format(v)
		return nil
	default:
		return errors.Errorf(`must be one of [%s, %s, %s]"`, jsonFormat, yamlFormat, stringFormat)
	}
}

// Type is only used in help text
func (f *format) Type() string {
	return "string"
}
