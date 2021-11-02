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
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/kuberlogic/kuberlogic/modules/apiserver/internal/generated/models"
	kuberlogicv1 "github.com/kuberlogic/kuberlogic/modules/operator/api/v1"
	errors2 "github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"strconv"
)

func BackupConfigResourceToModel(resource *v1.Secret) *models.BackupConfig {
	enabled, _ := strconv.ParseBool(resource.StringData["enabled"])
	return &models.BackupConfig{
		AwsAccessKeyID:     aws.String(resource.StringData["aws-access-key-id"]),
		AwsSecretAccessKey: aws.String(resource.StringData["aws-secret-access-key"]),
		Bucket:             aws.String(resource.StringData["bucket"]),
		Endpoint:           aws.String(resource.StringData["endpoint"]),
		Enabled:            aws.Bool(enabled),
		Schedule:           aws.String(resource.StringData["schedule"]),
		Region:             aws.String(resource.StringData["region"]),
	}
}

func BackupConfigModelToResource(model *models.BackupConfig) *v1.Secret {
	return &v1.Secret{
		StringData: map[string]string{
			"aws-access-key-id":     aws.StringValue(model.AwsAccessKeyID),
			"aws-secret-access-key": aws.StringValue(model.AwsSecretAccessKey),
			"bucket":                aws.StringValue(model.Bucket),
			"bucket-scope-suffix":   "",
			"endpoint":              aws.StringValue(model.Endpoint),
			"region":                aws.StringValue(model.Region),
			"sse":                   "AES256",
			"enabled":               strconv.FormatBool(aws.BoolValue(model.Enabled)),
			"schedule":              aws.StringValue(model.Schedule),
		},
	}
}

// TODO: Should be moved into operator's package/repo
func CreateBackupResource(client rest.Interface, ns, name, schedule string) error {
	if _, err := GetBackupResource(client, ns, name); !errors.IsNotFound(err) {
		return err // backup already exists
	}

	resource := kuberlogicv1.KuberLogicBackupSchedule{
		ObjectMeta: v12.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: kuberlogicv1.KuberLogicBackupScheduleSpec{
			Type:        "s3",
			Schedule:    schedule,
			ClusterName: name,
			SecretName:  name, // name of secret the same as name of resource
		},
	}

	result := kuberlogicv1.KuberLogicBackupSchedule{}
	err := client.Post().
		Namespace(ns).
		Name(name).
		Resource("kuberlogicbackupschedules").
		Body(&resource).
		Do(context.TODO()).
		Into(&result)
	if err != nil {
		return errors2.Wrap(err, "cannot create kuberlogicbackupschedule resource")
	}
	return nil
}

func UpdateBackupResource(client rest.Interface, ns, name, schedule string) error {
	resource, err := GetBackupResource(client, ns, name)
	if err != nil {
		return err
	}

	if resource.Spec.Schedule != schedule {
		resource.Spec.Schedule = schedule

		result := kuberlogicv1.KuberLogicBackupSchedule{}
		err := client.Put().
			Namespace(ns).
			Name(name).
			Resource("kuberlogicbackupschedules").
			Body(resource).
			Do(context.TODO()).
			Into(&result)
		if err != nil {
			return errors2.Wrap(err, "cannot update kuberlogicbackupschedule resource")
		}
		return nil
	}

	return nil
}

func DeleteBackupResource(client rest.Interface, ns, name string) error {
	_, err := GetBackupResource(client, ns, name)
	if errors.IsNotFound(err) {
		return nil // cluster does not exist
	} else if err != nil {
		return err
	}

	err = client.Delete().
		Namespace(ns).
		Resource("kuberlogicbackupschedules").
		Name(name).
		Body(&v12.DeleteOptions{}).
		Do(context.TODO()).
		Error()
	if err != nil {
		return errors2.Wrap(err, "cannot delete kuberlogicbackupschedule resource")
	}
	return nil
}

func GetBackupResource(client rest.Interface, ns, name string) (*kuberlogicv1.KuberLogicBackupSchedule, error) {
	item := kuberlogicv1.KuberLogicBackupSchedule{}
	err := client.Get().
		Namespace(ns).
		Resource("kuberlogicbackupschedules").
		Name(name).
		Do(context.TODO()).
		Into(&item)

	if err != nil {
		return nil, errors2.Wrap(err, "cannot receive kuberlogicbackupschedule resource")
	}
	return &item, nil
}

func CreateBackupRestoreResource(client rest.Interface, ns, name, backup, database string) error {
	resource := kuberlogicv1.KuberLogicBackupRestore{
		ObjectMeta: v12.ObjectMeta{
			GenerateName: name,
			Namespace:    ns,
		},
		Spec: kuberlogicv1.KuberLogicBackupRestoreSpec{
			Type:        "s3",
			ClusterName: name,
			SecretName:  name, // name of secret the same as name of resource
			Backup:      backup,
			Database:    database,
		},
	}

	result := kuberlogicv1.KuberLogicBackupRestore{}
	err := client.Post().
		Namespace(ns).
		Resource("kuberlogicbackuprestores").
		Body(&resource).
		Do(context.TODO()).
		Into(&result)
	if err != nil {
		return errors2.Wrap(err, "cannot create kuberlogicbackuprestore resource")
	}
	return nil
}
