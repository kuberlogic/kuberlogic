package internal

import (
	"context"
	"encoding/hex"
	"github.com/pkg/errors"
	v12 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strings"

	"crypto/rand"
)

const (
	releaseName          = "kuberlogic"
	releaseConfigMapName = "kuberlogic-metadata"
	cmStateKey           = "Status"
	cmBannerKey          = "Banner"
	cmSharedSecretKey    = "SharedSecret"

	releaseStartedPhase    = "started"
	releaseUpgradePhase    = "upgrading"
	releaseFailedPhase     = "failed"
	releaseSuccessfulPhase = "installed"

	// length of internal password used via Kuberlogic installation process
	internalPasswordLen = 10
)

type ReleaseInfo struct {
	Name        string
	Namespace   string
	Status      string
	bannerLines []string

	secret *v12.Secret
}

func (r ReleaseInfo) getState() (string, error) {
	if r.secret == nil {
		return "", errors.New("release Secret is not found")
	}
	s, f := r.secret.Data[cmStateKey]
	if !f || string(s) == "" {
		return "", errors.New("release state is empty")
	}
	return string(s), nil
}

func (r *ReleaseInfo) updateState(state string, clientSet *kubernetes.Clientset) error {
	r.secret.Data[cmStateKey] = []byte(state)

	cm, err := clientSet.CoreV1().Secrets(r.Namespace).Update(context.TODO(), r.secret, v1.UpdateOptions{})
	if err != nil {
		return err
	}
	r.secret = cm
	return nil
}

func (r *ReleaseInfo) updateBanner(clientSet *kubernetes.Clientset) error {
	r.secret.Data[cmBannerKey] = []byte(r.Banner())

	secret, err := clientSet.CoreV1().Secrets(r.Namespace).Update(context.TODO(), r.secret, v1.UpdateOptions{})
	if err != nil {
		return err
	}
	r.secret = secret
	return nil
}

func (r ReleaseInfo) InternalPassword() string {
	return string(r.secret.Data[cmSharedSecretKey])
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
	r.secret.Data[cmBannerKey] = []byte("")
	return err
}

func (r *ReleaseInfo) FinishRelease(clientSet *kubernetes.Clientset) error {
	err := r.updateBanner(clientSet)
	err = r.updateState(releaseSuccessfulPhase, clientSet)
	return err
}

func StartRelease(namespace string, clientSet *kubernetes.Clientset) (*ReleaseInfo, error) {
	if _, f, err := DiscoverReleaseInfo(namespace, clientSet); err != nil {
		return nil, errors.Wrap(err, "error checking for release")
	} else if f {
		return nil, errors.New("release already exists")
	}
	secret := &v12.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      releaseConfigMapName,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			cmStateKey:        []byte(releaseStartedPhase),
			cmSharedSecretKey: []byte(generateRandomString(internalPasswordLen)),
		},
	}

	s, err := clientSet.CoreV1().Secrets(namespace).Create(context.TODO(), secret, v1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	r := &ReleaseInfo{
		Name:      releaseName,
		Namespace: namespace,
		secret:    s,
	}

	if r.InternalPassword() == "" {
		return nil, errors.New("internal password can't be empty")
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
	r.secret, err = clientSet.CoreV1().Secrets(namespace).Get(context.TODO(), releaseConfigMapName, v1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	if r.InternalPassword() == "" {
		return nil, false, errors.New("internal password can't be empty")
	}

	r.Status, err = r.getState()
	r.bannerLines = strings.Split(string(r.secret.Data[cmBannerKey]), "\n")
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

	err = clientSet.CoreV1().ConfigMaps(namespace).Delete(context.TODO(), rel.secret.Name, v1.DeleteOptions{})
	return err
}

func generateRandomString(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
