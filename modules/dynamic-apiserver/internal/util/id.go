/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
