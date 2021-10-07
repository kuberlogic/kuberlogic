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

package store

// Error type for 1st level customer facing Service methods
type ServiceError struct {
	ClientMsg string
	Client    bool
	Err       error
}

func NewServiceError(clientMsg string, client bool, err error) *ServiceError {
	return &ServiceError{
		ClientMsg: clientMsg,
		Client:    client,
		Err:       err,
	}
}
