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

package interfaces

import (
	"github.com/kuberlogic/kuberlogic/modules/operator/api/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type OperatorInterface interface {
	Name(cm *v1.KuberLogicService) string
	Init(cm *v1.KuberLogicService, platform string)
	InitFrom(o runtime.Object)
	Update(cm *v1.KuberLogicService) error
	AsRuntimeObject() runtime.Object
	AsMetaObject() metav1.Object
	AsClientObject() client.Object
	IsReady() (bool, string)

	GetBackupSchedule() BackupSchedule
	GetBackupRestore() BackupRestore
	GetInternalDetails() InternalDetails
	GetSession(cm *v1.KuberLogicService, client kubernetes.Interface, db string) (Session, error)
}

// PlatformOperator interface holds methods for platform specific actions
// For example: managing a firewall, interacting with load balancers, etc
// SetAllowedIPs method configures a service to accepts traffic from a list of CIDR ranges passed as []string
type PlatformOperator interface {
	SetAllowedIPs([]string) error
}

type BackupSchedule interface {
	New(backup *v1.KuberLogicBackupSchedule) batchv1beta1.CronJob
	Init(*v1.KuberLogicBackupSchedule)
	InitFrom(*batchv1beta1.CronJob)
	IsEqual(cm *v1.KuberLogicBackupSchedule) bool
	Update(cm *v1.KuberLogicBackupSchedule)
	GetCronJob() *batchv1beta1.CronJob
	IsRunning(job *batchv1.Job) bool
	IsFinished(job *batchv1.Job) bool
	IsFailed(job *batchv1.Job) bool

	SetBackupImage(repo, version string)
	SetBackupEnv(cm *v1.KuberLogicBackupSchedule)
	SetServiceAccount(name string)
}

type BackupRestore interface {
	New(backup *v1.KuberLogicBackupRestore) batchv1.Job
	Init(*v1.KuberLogicBackupRestore)
	InitFrom(*batchv1.Job)
	GetJob() *batchv1.Job
	IsFailed() bool
	IsRunning() bool
	IsFinished() bool

	SetRestoreImage(repo, version string)
	SetRestoreEnv(cm *v1.KuberLogicBackupRestore)
	SetServiceAccount(name string)
}

type InternalDetails interface {
	GetCredentialsSecret() (*corev1.Secret, error)

	GetPodReplicaSelector() map[string]string
	GetPodMasterSelector() map[string]string

	GetMasterService() string
	GetReplicaService() string
	GetAccessPort() int

	GetDefaultConnectionPassword() (string, string)
	GetDefaultConnectionUser() string
	GetMainPodContainer() string
}

type Session interface {
	GetDatabase() Database
	GetUser() User
	CreateTable(table string) error
	ConnectionString(host, db string) string

	GetMasterIP() string
	GetReplicaIPs() []string
}

type Database interface {
	List() ([]string, error)
	Create(name string) error
	Drop(name string) error
	IsProtected(name string) bool
}

type User interface {
	List() ([]string, error)
	Create(name, password string) error
	Delete(name string) error
	Edit(name, password string) error
	IsProtected(name string) bool
}
