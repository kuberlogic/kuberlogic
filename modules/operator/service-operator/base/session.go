package base

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	client2 "sigs.k8s.io/controller-runtime/pkg/client"
)

type BaseSession struct {
	ClusterName      string
	ClusterNamespace string

	Database string
	Port     int
	Username string
	Password string

	ClusterCredentialsSecret string

	MasterIP   string
	ReplicaIPs []string
}

func (session *BaseSession) GetPods(client *kubernetes.Clientset, matchingLabels client2.MatchingLabels) (*v1.PodList, error) {
	labelMap, err := metav1.LabelSelectorAsMap(&metav1.LabelSelector{
		MatchLabels: matchingLabels,
	})
	if err != nil {
		return nil, err
	}

	options := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labelMap).String(),
	}

	pods, err := client.CoreV1().
		Pods("").
		List(context.TODO(), options)
	if err != nil {
		return nil, err
	}
	return pods, nil
}

func (session *BaseSession) GetMasterIP() string {
	return session.MasterIP
}

func (session *BaseSession) GetReplicaIPs() []string {
	return session.ReplicaIPs
}
