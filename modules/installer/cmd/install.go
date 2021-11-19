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

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

const (
	AllArg         = "all"
	certManagerArg = "cert-manager"
	depsArg        = "dependencies"
	kuberlogicArg  = "kuberlogic"
)

// newInstallCmd returns the "install" command
func newInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:       fmt.Sprintf("install [%s | %s | %s | %s]", AllArg, certManagerArg, depsArg, kuberlogicArg),
		ValidArgs: []string{AllArg, certManagerArg, depsArg, kuberlogicArg},
		Args:      cobra.OnlyValidArgs,
		Short:     "installs a Kuberlogic release",
		Run: func(cmd *cobra.Command, args []string) {
			kuberlogicInstaller.Exit(kuberlogicInstaller.Install(cmd, args))
		},
	}
}
