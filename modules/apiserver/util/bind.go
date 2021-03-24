package util

import (
	"context"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"strconv"
)

func BackupConfigResourceToModel(resource *v1.Secret) *models.BackupConfig {
	awsAccessKeyID := string(resource.Data["aws-access-key-id"])
	awsSecretAccessKey := string(resource.Data["aws-secret-access-key"])
	bucket := string(resource.Data["bucket"])
	endpoint := string(resource.Data["endpoint"])

	enabled, _ := strconv.ParseBool(string(resource.Data["enabled"]))
	schedule := string(resource.Data["schedule"])

	return &models.BackupConfig{
		AwsAccessKeyID:     &awsAccessKeyID,
		AwsSecretAccessKey: &awsSecretAccessKey,
		Bucket:             &bucket,
		Endpoint:           &endpoint,
		Enabled:            &enabled,
		Schedule:           &schedule,
	}
}

func BackupConfigModelToResource(model *models.BackupConfig) *v1.Secret {
	accessKey := *model.AwsAccessKeyID
	secretKey := *model.AwsSecretAccessKey
	bucket := *model.Bucket
	endpoint := *model.Endpoint
	enabled := *model.Enabled
	schedule := *model.Schedule

	return &v1.Secret{
		StringData: map[string]string{
			"aws-access-key-id":     accessKey,
			"aws-secret-access-key": secretKey,
			"bucket":                bucket,
			"bucket-scope-suffix":   "",
			"endpoint":              endpoint,
			"region":                "",
			"sse":                   "AES256",
			"enabled":               strconv.FormatBool(enabled),
			"schedule":              schedule,
		},
	}
}

// TODO: Should be moved into operator's package/repo
func CreateBackupResource(client *rest.RESTClient, ns, name, schedule string) error {
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
		Resource("kuberlogicbackupschedules").
		Body(&resource).
		Do(context.TODO()).
		Into(&result)
	if err != nil {
		return err
	}
	return nil
}

func UpdateBackupResource(client *rest.RESTClient, ns, name, schedule string) error {
	resource, err := GetBackupResource(client, ns, name)
	if err != nil {
		return err
	}

	if resource.Spec.Schedule != schedule {
		resource.Spec.Schedule = schedule

		result := kuberlogicv1.KuberLogicBackupSchedule{}
		err := client.Put().
			Namespace(ns).
			Resource("kuberlogicbackupschedules").
			Body(&resource).
			Do(context.TODO()).
			Into(&result)
		if err != nil {
			return err
		}
		return nil
	}

	return nil
}

func DeleteBackupResource(client *rest.RESTClient, ns, name string) error {
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
		return err
	}
	return nil
}

func GetBackupResource(client *rest.RESTClient, ns, name string) (*kuberlogicv1.KuberLogicBackupSchedule, error) {
	item := kuberlogicv1.KuberLogicBackupSchedule{}
	err := client.Get().
		Namespace(ns).
		Resource("kuberlogicbackupschedules").
		Name(name).
		Do(context.TODO()).
		Into(&item)

	if err != nil {
		return nil, err
	}
	return &item, nil
}

func CreateBackupRestoreResource(client *rest.RESTClient, ns, name, backup, database string) error {
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
		return err
	}
	return nil
}
