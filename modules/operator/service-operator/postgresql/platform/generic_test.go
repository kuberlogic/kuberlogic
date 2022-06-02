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

package platform

import (
	v1 "github.com/zalando/postgres-operator/pkg/apis/acid.zalan.do/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

var (
	testPostgres = &PostgresGeneric{
		Spec: &v1.Postgresql{
			ObjectMeta: v12.ObjectMeta{
				Name:      "test",
				Namespace: "test",
			},
		},
	}
)

func TestPostgresGeneric_SetAllowedIPs(t *testing.T) {

	for _, inputList := range [][]string{
		{"1.1.1.1/32"},
		{"2.2.2.2/32", "8.8.8.8/24"},
	} {
		if err := testPostgres.SetAllowedIPs(inputList); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}
}
