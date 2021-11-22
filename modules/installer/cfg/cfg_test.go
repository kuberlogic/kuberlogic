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

package cfg

import (
	logger "github.com/kuberlogic/kuberlogic/modules/installer/log"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"testing"
)

var (
	log = logger.NewLogger(true)
)

func TestInvalidCfg(t *testing.T) {
	rawCfg := `corrupted`
	cfgFile, cleanup, err := testCfgFile(rawCfg)
	defer cleanup()
	if err != nil {
		t.Fatalf("unexpected error during config file preparation: %v", err)
	}

	_, err = NewConfigFromFile(cfgFile, log)
	if err == nil {
		t.Fatalf("config marshaling expected to fail with error")
	}
}

func TestConfigIncomplete(t *testing.T) {
	incompleteConfigs := []string{
		`
---`, `
---
endpoints:
  kuberlogic: example.com`, `
---
endpoints:
  kuberlogic: example.com
  monitoring-console: mc.example.com`, `
---
endpoints:
  kuberlogic: example.com
  monitoring-console: mc.example.com`}
	for _, rawCfg := range incompleteConfigs {
		cfgFile, cleanup, err := testCfgFile(rawCfg)
		if err != nil {
			t.Fatalf("unexpected error during config file preparation: %v", err)
		}
		_, err = NewConfigFromFile(cfgFile, log)
		cleanup()
		if !errors.Is(err, errRequiredParamNotSet) {
			t.Fatalf("config marshaling expected to fail with error")
		}
	}
}

func TestValidCfg(t *testing.T) {
	rawCfg := `
---
endpoints:
  kuberlogic: example.com
  monitoring-console: mc.example.com
namespace: kuberlogic`
	cfgFile, cleanup, err := testCfgFile(rawCfg)
	defer cleanup()
	if err != nil {
		t.Fatalf("unexpected error during config file preparation: %v", err)
	}

	c, err := NewConfigFromFile(cfgFile, log)
	if err != nil || c == nil {
		t.Fatalf("unexpected error during config marshaling: %v", err)
	}
}

// testCfgFile creates a config file with "content" content
// and returns a filename, cleanup function and an error that might have happened during the whole process
func testCfgFile(content string) (string, func(), error) {
	f, err := ioutil.TempFile(".", "test-config")
	if err != nil {
		return "", func() {}, err
	}
	_, err = f.Write([]byte(content))
	return f.Name(), func() { os.Remove(f.Name()) }, err
}
