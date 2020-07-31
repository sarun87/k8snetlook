package k8snetlook

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var clientset *kubernetes.Clientset

func getKubernetesClient(kubeconfigPath string) (*kubernetes.Clientset, error) {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

func getServiceClusterIP(namespace string, serviceName string) Endpoint {
	service, err := clientset.CoreV1().Services(namespace).Get(serviceName, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("Error fetching %s service in %s ns. Error: %v", serviceName, namespace, err)
		return Endpoint{}
	}
	// Return one port only
	return Endpoint{IP: service.Spec.ClusterIP, Port: service.Spec.Ports[0].Port}
}

func getPodIPFromName(namespace string, podName string) string {
	// if namespace == "" i.e metav1.NamespaceAll, then all pods are listed
	pod, err := clientset.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("Error fetching %s pod in %s ns. Error: %v", podName, namespace, err)
		return ""
	}
	return pod.Status.PodIP
}

func getContainerIDFromPod(namespace string, podName string) string {
	// if namespace == "" i.e metav1.NamespaceAll, then all pods are listed
	pod, err := clientset.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("Error fetching %s pod in %s ns. Error: %v", podName, namespace, err)
		return ""
	}
	// Pod should have alteast one container (pause)
	return pod.Status.ContainerStatuses[0].ContainerID
}

func getEndpointsFromService(namespace string, serviceName string) []Endpoint {
	var ret []Endpoint
	endpoints, err := clientset.CoreV1().Endpoints(namespace).Get(serviceName, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("Error fetching %s service endpoints in %s ns. Error: %v", serviceName, namespace, err)
		return ret
	}
	for _, subset := range endpoints.Subsets {
		for _, ip := range subset.Addresses {
			for _, port := range subset.Ports {
				ret = append(ret, Endpoint{IP: ip.IP, Port: port.Port})
			}
		}
	}
	return ret
}

func getSvcAccountToken() (string, error) {
	svcAccount, err := clientset.CoreV1().ServiceAccounts("default").Get("default", metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("Error fetching default service acccount. Error: %s", err)
	}
	secret, err := clientset.CoreV1().Secrets("default").Get(svcAccount.Secrets[0].Name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("Error fetching secret for service account. Error: %s", err)
	}
	return string(secret.Data["token"]), nil
}
