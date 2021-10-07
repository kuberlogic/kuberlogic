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
	v1 "k8s.io/api/core/v1"
	"os"
)

func BucketVariables(secret string) []v1.EnvVar {
	return []v1.EnvVar{
		{
			Name:      "LOGICAL_BACKUP_S3_BUCKET",
			ValueFrom: FromSecret(secret, "bucket"),
		},
		{
			Name:      "LOGICAL_BACKUP_S3_REGION",
			ValueFrom: FromSecret(secret, "region"),
		},
		{
			Name:      "LOGICAL_BACKUP_S3_ENDPOINT",
			ValueFrom: FromSecret(secret, "endpoint"),
		},
		{
			Name:      "LOGICAL_BACKUP_S3_SSE",
			ValueFrom: FromSecret(secret, "sse"),
		},
		{
			Name:      "LOGICAL_BACKUP_S3_BUCKET_SCOPE_SUFFIX",
			ValueFrom: FromSecret(secret, "bucket-scope-suffix"),
		},
	}
}

func S3Credentials(secret string) []v1.EnvVar {
	return []v1.EnvVar{
		{
			Name:      "AWS_ACCESS_KEY_ID",
			ValueFrom: FromSecret(secret, "aws-access-key-id"),
		},
		{
			Name:      "AWS_SECRET_ACCESS_KEY",
			ValueFrom: FromSecret(secret, "aws-secret-access-key"),
		},
	}
}

func SentryEnv() []v1.EnvVar {
	var env []v1.EnvVar
	if dsn := os.Getenv("SENTRY_DSN"); dsn != "" {
		env = append(env, v1.EnvVar{
			Name:  "SENTRY_DSN",
			Value: os.Getenv("SENTRY_DSN"),
		})
	}
	return env
}
