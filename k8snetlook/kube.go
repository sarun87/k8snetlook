package k8snetlook

import (
	"fmt"
	"time"

	log "github.com/sarun87/k8snetlook/logutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"golang.org/x/net/context"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var clientset *kubernetes.Clientset

func initKubernetesClient(kubeconfigPath string) error {
	// check if running in-cluster. If so initialize client-set using incluster method
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Debug("Not running in Pod or unable to fetch config via incluster method. Error:%v", err)
		log.Debug("Trying from kubeconfig specified via command line flag")
		// Fall back to using config provided as part of command line arguments
		// use the current context in kubeconfig
		if kubeconfigPath == "" {
			return fmt.Errorf("Not running in Pod & kubeconfig not specified using -config flag")
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return err
		}
	}
	config.Timeout = time.Second * 4
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	return nil
}

func getServiceClusterIP(namespace string, serviceName string) (Endpoint, error) {
	service, err := clientset.CoreV1().Services(namespace).Get(context.TODO(), serviceName, metav1.GetOptions{})
	if err != nil {
		log.Error("Error fetching %s service in %s ns. Error: %v", serviceName, namespace, err)
		return Endpoint{}, err
	}
	// Return one port only
	return Endpoint{IP: service.Spec.ClusterIP, Port: service.Spec.Ports[0].Port}, nil
}

func getPodIPFromName(namespace string, podName string) string {
	// if namespace == "" i.e metav1.NamespaceAll, then all pods are listed
	pod, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		log.Error("Error fetching %s pod in %s ns. Error: %v", podName, namespace, err)
		return ""
	}
	return pod.Status.PodIP
}

func getContainerIDFromPod(namespace string, podName string) string {
	// if namespace == "" i.e metav1.NamespaceAll, then all pods are listed
	pod, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		log.Error("Error fetching %s pod in %s ns. Error: %v", podName, namespace, err)
		return ""
	}
	// Pod should have alteast one container (pause)
	return pod.Status.ContainerStatuses[0].ContainerID
}

func getEndpointsFromService(namespace string, serviceName string) []Endpoint {
	var ret []Endpoint
	endpoints, err := clientset.CoreV1().Endpoints(namespace).Get(context.TODO(), serviceName, metav1.GetOptions{})
	if err != nil {
		log.Error("Error fetching %s service endpoints in %s ns. Error: %v", serviceName, namespace, err)
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
	svcAccount, err := clientset.CoreV1().ServiceAccounts("default").Get(context.TODO(), "default", metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("Error fetching default service acccount. Error: %s", err)
	}
	secret, err := clientset.CoreV1().Secrets("default").Get(context.TODO(), svcAccount.Secrets[0].Name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("Error fetching secret for service account. Error: %s", err)
	}
	return string(secret.Data["token"]), nil
}
