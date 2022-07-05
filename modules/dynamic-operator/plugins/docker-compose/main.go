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

package main

import (
	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/plugin/commons"
	"os"
)

func main() {
	cfg, err := getConfig()
	if err != nil {
		panic(err)
	}
	cfgRaw, err := os.ReadFile(cfg.ComposeFile)
	if err != nil {
		panic(err)
	}
	config, err := loader.ParseYAML(cfgRaw)
	if err != nil {
		panic(err)
	}
	project, err := loader.Load(types.ConfigDetails{
		WorkingDir: "/tmp",
		ConfigFiles: []types.ConfigFile{
			{
				Filename: cfg.ComposeFile,
				Config:   config,
			},
		},
	})
	if err != nil {
		panic(err)
	}
	commons.ServePlugin("docker-compose", newDockerComposeServicePlugin(project))
}
