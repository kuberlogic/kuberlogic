package operator

import (
	"fmt"
	"github.com/pkg/errors"
	cloudlinuxv1 "gitlab.com/cloudmanaged/operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"os"
	"strings"
)

type Operator interface {
	Init(cm *cloudlinuxv1.CloudManaged)
	InitFrom(o runtime.Object)
	Update(cm *cloudlinuxv1.CloudManaged)
	AsRuntimeObject() runtime.Object
	AsMetaObject() metav1.Object
	IsEqual(cm *cloudlinuxv1.CloudManaged) bool
	CurrentStatus() string
	GetDefaults() cloudlinuxv1.Defaults
}

func GetOperator(t string) (Operator, error) {
	var operators = map[string]Operator{
		"postgresql": &Postgres{},
		"redis":      &Redis{},
		"mysql":      &Mysql{},
	}

	value, ok := operators[t]
	if !ok {
		return nil, errors.Errorf("Operator %s is not supported", t)
	}
	return value, nil
}

func getImage(base, v string) string {
	repo := os.Getenv("IMAGE_REPO")
	return fmt.Sprintf("%s/%s:%s", strings.TrimSuffix(repo, "/"), base, v)
}

func getImagePullSecret() string {
	return os.Getenv("IMAGE_PULL_SECRET")
}
