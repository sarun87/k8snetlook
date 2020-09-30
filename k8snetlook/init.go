package k8snetlook

import (
	"context"
	"os"
	"strings"

	"github.com/docker/docker/client"
	log "github.com/sarun87/k8snetlook/logutil"
	"github.com/sarun87/k8snetlook/netutils"
	"github.com/vishvananda/netns"
)

const (
	defaultInternetEgressTestIP = "8.8.8.8"
)

// Pod struct specifies properties required for Pod network debugging
type Pod struct {
	Name      string
	Namespace string
	IP        string
	NsHandle  netns.NsHandle // Initializes this with an open FD to the netns file /proc/<pid>/ns/net
}

// Service struct specifies properties required for decribing a K8s service
type Service struct {
	Name         string
	Namespace    string
	ClusterIP    Endpoint
	SvcEndpoints []Endpoint
}

// Endpoint struct specifies properties that an Endpoint represents
type Endpoint struct {
	IP   string
	Port int32
}

// Config struct represents the properties required by k8snetlook to run checks
// most properties are populated from user input
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

// Check describes the reporting structure for a network check
type Check struct {
	Name     string `json:"name"`
	Success  bool   `json:"success"`
	ErrorMsg error  `json:"error_msg"`
}

// Checker stores check names and results for all of the checks
type Checker struct {
	PodChecks  []Check `json:"pod_checks,omitempty"`
	HostChecks []Check `json:"host_checks"`
}

// allChecks holds information about the checks being run by k8snetlook
var allChecks Checker

// Cfg is an instance of Config struct
var Cfg Config

// Init initializes k8snetlook
func Init(kubeconfigPath string) error {
	if err := initKubernetesClient(kubeconfigPath); err != nil {
		return err
	}
	return initK8sInfo()
}

// initK8sInfo initializes information related to pods, services by querying k8s api
func initK8sInfo() error {
	var err error
	// If we aren't able to talk to the k8sapiserver endpoint via ip specified by kubeconfig,
	// return. Do not execute further
	if Cfg.KubeAPIService, err = getServiceClusterIP("default", "kubernetes"); err != nil {
		return err
	}
	Cfg.HostGatewayIP, _ = netutils.GetHostGatewayIP()
	Cfg.KubeDNSService, _ = getServiceClusterIP("kube-system", "kube-dns")
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
		Cfg.DstSvc.ClusterIP, _ = getServiceClusterIP(Cfg.DstSvc.Namespace, Cfg.DstSvc.Name)
		Cfg.DstSvc.SvcEndpoints = getEndpointsFromService(Cfg.DstSvc.Namespace, Cfg.DstSvc.Name)
	}
	return nil
}

func getPodNetnsHandle(namespace string, podName string) netns.NsHandle {
	containerID := getContainerIDFromPod(namespace, podName)
	if containerID == "" {
		log.Error("Unable to fetch container id for pod %s. Exiting..\n", podName)
		Cleanup()
		os.Exit(1)
	}
	containerID = strings.TrimPrefix(containerID, "docker://")
	log.Debug("ContainerID:%s\n", containerID)
	cli, err := client.NewClient(client.DefaultDockerHost, client.DefaultVersion, nil, nil)
	if err != nil {
		log.Error("Unable to create docker client: %v", err)
		os.Exit(1)
	}
	containerJSON, err := cli.ContainerInspect(context.Background(), containerID)
	if err != nil {
		log.Error("Unable to inspect container: %v", err)
		os.Exit(1)
	}
	log.Debug("Pid of container: %d\n", containerJSON.State.Pid)
	nshandle, err := netns.GetFromPid(containerJSON.State.Pid)
	if err != nil {
		log.Error("Unable to fetch netns handle for pod %s. Error: %v Exiting..\n", podName, err)
		Cleanup()
		os.Exit(1)
	}
	return nshandle
}

// Cleanup closes all of the open network namespaces handles
func Cleanup() {
	if Cfg.SrcPod.NsHandle.IsOpen() {
		Cfg.SrcPod.NsHandle.Close()
	}
}
