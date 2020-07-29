package k8snetlook

import (
	"fmt"
	"os"

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
	Name      string
	Namespace string
	IP        string
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

	KubeAPIServiceIP string
	HostGatewayIP    string
	KubeDNSServiceIP string
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
	Cfg.KubeAPIServiceIP = getServiceClusterIP("default", "kubernetes")
	Cfg.KubeDNSServiceIP = getServiceClusterIP("kube-system", "kube-dns")
	Cfg.HostGatewayIP = getHostGatewayIP()
	Cfg.SrcPod.NsHandle = netns.NsHandle(-1)
	if Cfg.SrcPod.Name != "" {
		Cfg.SrcPod.IP = getPodIPFromName(Cfg.SrcPod.Namespace, Cfg.SrcPod.Name)
		Cfg.SrcPod.NsHandle = getPodNetnsHandle(Cfg.SrcPod.Namespace, Cfg.SrcPod.Name)
	}
	Cfg.DstPod.NsHandle = netns.NsHandle(-1)
	if Cfg.DstPod.Name != "" {
		Cfg.DstPod.IP = getPodIPFromName(Cfg.DstPod.Namespace, Cfg.DstPod.Name)
		Cfg.DstPod.NsHandle = getPodNetnsHandle(Cfg.DstPod.Namespace, Cfg.DstPod.Name)
	}
	if Cfg.DstSvc.Name != "" {
		Cfg.DstSvc.IP = getServiceClusterIP(Cfg.DstSvc.Namespace, Cfg.DstSvc.Name)
	}
}

func getPodNetnsHandle(namespace string, podName string) netns.NsHandle {
	containerID := getContainerIDFromPod(namespace, podName)
	if containerID == "" {
		fmt.Printf("Unable to fetch container id for pod %s. Exiting..\n", podName)
		Cleanup()
		os.Exit(1)
	}
	nshandle, error := netns.GetFromDocker(containerID)
	if error != nil {
		fmt.Printf("Unable to fetch netns handle for pod %s. Exiting..\n", podName)
		Cleanup()
		os.Exit(1)
	}
	return nshandle
}

func Cleanup() {
	if Cfg.SrcPod.NsHandle.IsOpen() {
		Cfg.SrcPod.NsHandle.Close()
	}
	if Cfg.DstPod.NsHandle.IsOpen() {
		Cfg.DstPod.NsHandle.Close()
	}
}
