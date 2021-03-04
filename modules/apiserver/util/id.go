package util

import (
	"errors"
	"fmt"
	"strings"
)

func SplitID(str string) (string, string, error) {
	s := strings.Split(str, ":")
	if len(s) != 2 {
		return "", "", errors.New(fmt.Sprintf("%s is incorrect", str))
	}
	if s[0] == "" || s[1] == "" {
		return "", "", errors.New(fmt.Sprint("name or namespace can't be empty"))
	}

	return s[0], s[1], nil
}

func JoinID(ns, name string) (string, error) {
	if ns == "" || name == "" {
		return "", errors.New("name or namespace can't be empty")
	}
	return fmt.Sprintf("%s:%s", ns, name), nil
}
