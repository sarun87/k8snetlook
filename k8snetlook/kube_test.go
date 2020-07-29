package k8snetlook

import "testing"

func TestGetKubernetesClient(t *testing.T) {
	// Specify valid kubeconfig to pass the test
	kubeconfig := "sample_kubeconfig.yaml"
	if _, err := getKubernetesClient(kubeconfig); err != nil {
		t.Errorf("Unable to instantiate clientset. Error: %v", err)
	}
}
