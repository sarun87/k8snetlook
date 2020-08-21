package k8snetlook

import (
	"fmt"
	"runtime"

	"github.com/vishvananda/netns"
)

var (
	PassingPodChecks int
	TotalPodChecks   int
)

func RunPodChecks() {
	// 1. Switch to SrcPod network namespace
	// 2. Run checks from within pod network namespace
	// 3. Switch back to host network namespace

	fmt.Println("----------- Pod Checks -----------")

	TotalPodChecks = 4
	// Lock OS thread to prevent ns change
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Get current network ns (Should be the host network ns)
	hostNsHandle, err := netns.Get()
	if err != nil {
		fmt.Println("Unable to get handle to current netns", err)
		return
	}
	defer hostNsHandle.Close()

	// Change network ns to SrcPod network ns
	if err := netns.Set(Cfg.SrcPod.NsHandle); err != nil {
		fmt.Println("Unable to switch to pod network namespace. Error:", err)
		return
	}

	// Execute checks from within the Pod network ns
	fmt.Println("----> [From SrcPod] Running Kube service IP connectivity check..")
	RunKubeAPIServiceIPConnectivityCheck(&PassingPodChecks)
	fmt.Println("----> [From SrcPod] Running Kube API Server Endpoint IP connectivity check..")
	RunKubeAPIEndpointIPConnectivityCheck(&PassingPodChecks)
	fmt.Println("----> [From SrcPod] Running default gateway connectivity check..")
	RunGatewayConnectivityCheck(&PassingPodChecks)
	fmt.Println("----> [From SrcPod] Running DNS lookup test (kubernetes.default)..")
	RunK8sDNSLookupCheck(Cfg.KubeDNSService.IP, "kubernetes", "default",
		Cfg.KubeAPIService.IP, &PassingPodChecks)

	if Cfg.DstPod.IP != "" {
		TotalPodChecks++
		fmt.Println("----> [From SrcPod] Running DstPod connectivity check..")
		RunDstConnectivityCheck(Cfg.DstPod.IP, &PassingPodChecks)
		TotalPodChecks++
		fmt.Println("----> [From SrcPod] Running pmtud check for dstIP..")
		RunMTUProbeToDstIPCheck(Cfg.DstPod.IP, &PassingPodChecks)
	}

	if Cfg.ExternalIP != "" {
		TotalPodChecks++
		fmt.Println("----> [From SrcPod] Running externalIP connectivity check..")
		RunDstConnectivityCheck(Cfg.ExternalIP, &PassingPodChecks)
		TotalPodChecks++
		fmt.Println("----> [From SrcPod] Running pmtud check for externalIP..")
		RunMTUProbeToDstIPCheck(Cfg.ExternalIP, &PassingPodChecks)
	}

	if Cfg.DstSvc.SvcEndpoint.IP != "" {
		TotalPodChecks++
		fmt.Println("----> [From SrcPod] Running DstSvc DNS lookup check..")
		RunK8sDNSLookupCheck(Cfg.KubeDNSService.IP, Cfg.DstSvc.Name, Cfg.DstSvc.Namespace,
			Cfg.DstSvc.SvcEndpoint.IP, &PassingPodChecks)
	}

	// Change network ns back to host
	netns.Set(hostNsHandle)

	fmt.Println("-----------------------------------")
}
