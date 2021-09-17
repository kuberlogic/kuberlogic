package internal

import (
	"context"
	"github.com/pkg/errors"
	v12 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strings"
)

const (
	releaseName          = "kuberlogic"
	releaseConfigMapName = "kuberlogic-metadata"
	cmStateKey           = "Status"
	cmBannerKey          = "Banner"

	releaseStartedPhase    = "started"
	releaseUpgradePhase    = "upgrading"
	releaseFailedPhase     = "failed"
	releaseSuccessfulPhase = "installed"
)

type ReleaseInfo struct {
	Name        string
	Namespace   string
	Status      string
	bannerLines []string

	cm *v12.ConfigMap
}

func (r ReleaseInfo) getState() (string, error) {
	if r.cm == nil {
		return "", errors.New("release ConfigMap is not found")
	}
	s, f := r.cm.Data[cmStateKey]
	if !f || s == "" {
		return "", errors.New("release state is empty")
	}
	return s, nil
}

func (r *ReleaseInfo) updateState(state string, clientSet *kubernetes.Clientset) error {
	r.cm.Data[cmStateKey] = state
	r.cm.Data[cmBannerKey] = r.Banner()

	cm, err := clientSet.CoreV1().ConfigMaps(r.Namespace).Update(context.TODO(), r.cm, v1.UpdateOptions{})
	if err != nil {
		return err
	}
	r.cm = cm
	return nil
}

func (r *ReleaseInfo) AddBannerLines(lines ...string) {
	r.bannerLines = append(r.bannerLines, lines...)
}

func (r ReleaseInfo) Banner() string {
	return strings.Join(r.bannerLines, "\n")
}

func (r ReleaseInfo) ShowBanner() bool {
	return len(r.bannerLines) != 0
}

func (r *ReleaseInfo) UpgradeRelease(clientSet *kubernetes.Clientset) error {
	err := r.updateState(releaseUpgradePhase, clientSet)
	return err
}

func (r *ReleaseInfo) FinishRelease(clientSet *kubernetes.Clientset) error {
	err := r.updateState(releaseSuccessfulPhase, clientSet)
	return err
}

func StartRelease(namespace string, clientSet *kubernetes.Clientset) (*ReleaseInfo, error) {
	if _, f, err := DiscoverReleaseInfo(namespace, clientSet); err != nil {
		return nil, errors.Wrap(err, "error checking for release")
	} else if f {
		return nil, errors.New("release already exists")
	}
	cm := &v12.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      releaseConfigMapName,
			Namespace: namespace,
		},
		Data: map[string]string{
			cmStateKey: releaseStartedPhase,
		},
	}

	cm, err := clientSet.CoreV1().ConfigMaps(namespace).Create(context.TODO(), cm, v1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	r := &ReleaseInfo{
		Name:      releaseName,
		Namespace: namespace,
		cm:        cm,
	}
	return r, err
}

func FailRelease(namespace string, clientSet *kubernetes.Clientset) (*ReleaseInfo, error) {
	rel, found, err := DiscoverReleaseInfo(namespace, clientSet)
	if !found {
		return nil, errors.New("release is not found")
	}
	if err != nil {
		return nil, errors.Wrap(err, "error finding release")
	}
	err = rel.updateState(releaseFailedPhase, clientSet)
	return rel, err
}

func DiscoverReleaseInfo(namespace string, clientSet *kubernetes.Clientset) (*ReleaseInfo, bool, error) {
	r := &ReleaseInfo{
		Name:        releaseName,
		Namespace:   namespace,
		bannerLines: make([]string, 0),
	}
	var err error
	r.cm, err = clientSet.CoreV1().ConfigMaps(namespace).Get(context.TODO(), releaseConfigMapName, v1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	r.Status, err = r.getState()
	r.bannerLines = strings.Split(r.cm.Data[cmBannerKey], "\n")
	return r, true, nil
}

func CleanupReleaseInfo(namespace string, clientSet *kubernetes.Clientset) error {
	rel, found, err := DiscoverReleaseInfo(namespace, clientSet)
	// release info is not found, exiting
	if !found {
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "error finding release")
	}

	err = clientSet.CoreV1().ConfigMaps(namespace).Delete(context.TODO(), rel.cm.Name, v1.DeleteOptions{})
	return err
}
