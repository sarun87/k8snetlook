package k8snetlook

import (
	"fmt"
	"os"
	"strings"

	"github.com/sarun87/k8snetlook/netutils"
	"github.com/vishvananda/netns"
)

const (
	defaultInternetEgressTestIP = "8.8.8.8"
)

type Pod struct {
	Name      string
	Namespace string
	IP        string
	NsHandle  netns.NsHandle // Initializes this with an open FD to the netns file /proc/<pid>/ns/net
}

type Service struct {
	Name        string
	Namespace   string
	SvcEndpoint Endpoint
}

type Endpoint struct {
	IP   string
	Port int32
}

type Config struct {
	SrcPod         Pod
	DstPod         Pod
	DstSvc         Service
	ExternalIP     string
	KubeconfigPath string

	KubeAPIService Endpoint
	KubeDNSService Endpoint
	HostGatewayIP  string
}

var Cfg Config

func InitKubeClient(kubeconfigPath string) {
	var err error
	clientset, err = getKubernetesClient(kubeconfigPath)
	if err != nil {
		fmt.Printf("Unable to initialize kubernetes client. Error: %v\n", err)
		os.Exit(1)
	}
}

func InitK8sInfo() {
	Cfg.KubeAPIService = getServiceClusterIP("default", "kubernetes")
	Cfg.KubeDNSService = getServiceClusterIP("kube-system", "kube-dns")
	Cfg.HostGatewayIP = netutils.GetHostGatewayIP()
	Cfg.SrcPod.NsHandle = netns.NsHandle(-1)
	if Cfg.SrcPod.Name != "" && Cfg.SrcPod.Namespace != "" {
		Cfg.SrcPod.IP = getPodIPFromName(Cfg.SrcPod.Namespace, Cfg.SrcPod.Name)
		Cfg.SrcPod.NsHandle = getPodNetnsHandle(Cfg.SrcPod.Namespace, Cfg.SrcPod.Name)
	}
	Cfg.DstPod.NsHandle = netns.NsHandle(-1)
	if Cfg.DstPod.Name != "" && Cfg.DstPod.Namespace != "" {
		Cfg.DstPod.IP = getPodIPFromName(Cfg.DstPod.Namespace, Cfg.DstPod.Name)
	}
	if Cfg.DstSvc.Name != "" && Cfg.DstSvc.Namespace != "" {
		Cfg.DstSvc.SvcEndpoint = getServiceClusterIP(Cfg.DstSvc.Namespace, Cfg.DstSvc.Name)
	}
}

func getPodNetnsHandle(namespace string, podName string) netns.NsHandle {
	containerID := getContainerIDFromPod(namespace, podName)
	if containerID == "" {
		fmt.Printf("Unable to fetch container id for pod %s. Exiting..\n", podName)
		Cleanup()
		os.Exit(1)
	}
	fmt.Println("src pod container id:", containerID)
	nshandle, err := netns.GetFromDocker(strings.TrimPrefix(containerID, "docker://"))
	if err != nil {
		fmt.Printf("Unable to fetch netns handle for pod %s. Error: %v Exiting..\n", podName, err)
		Cleanup()
		os.Exit(1)
	}
	return nshandle
}

func Cleanup() {
	if Cfg.SrcPod.NsHandle.IsOpen() {
		Cfg.SrcPod.NsHandle.Close()
	}
}
