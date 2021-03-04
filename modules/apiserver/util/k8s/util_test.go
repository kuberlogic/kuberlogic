package k8s

import "testing"

func TestMapToStrSelector(t *testing.T) {
	k8sSelector := map[string]string{"app": "test", "env": "testing"}
	k8sSelectorExpected := "app=test,env=testing"
	k8sSelectorActual := MapToStrSelector(k8sSelector)

	if k8sSelectorActual != k8sSelectorExpected {
		t.Errorf("failed, got %s, want %s", k8sSelectorActual, k8sSelectorExpected)
	}
}
