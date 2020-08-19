package k8snetlook

import (
	"fmt"
	"runtime"

	"github.com/vishvananda/netns"
)

var (
	passingPodChecks int
	totalPodChecks   int
)

func RunPodChecks() {
	// 1. Switch to SrcPod network namespace
	// 2. Run checks from within pod network namespace
	// 3. Switch back to host network namespace

	fmt.Println("----------- Pod Checks -----------")

	totalPodChecks = 4
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
	RunKubeAPIServiceIPConnectivityCheck(&passingPodChecks)
	fmt.Println("----> [From SrcPod] Running Kube API Server Endpoint IP connectivity check..")
	RunKubeAPIEndpointIPConnectivityCheck(&passingPodChecks)
	fmt.Println("----> [From SrcPod] Running default gateway connectivity check..")
	RunGatewayConnectivityCheck(&passingPodChecks)
	fmt.Println("----> [From SrcPod] Running DNS lookup test (kubernetes.default)..")
	RunK8sDNSLookupCheck(Cfg.KubeDNSService.IP, "kubernetes", "default",
		Cfg.KubeAPIService.IP, &passingPodChecks)

	if Cfg.DstPod.IP != "" {
		totalPodChecks++
		fmt.Println("----> [From SrcPod] Running DstPod connectivity check..")
		RunDstConnectivityCheck(Cfg.DstPod.IP, &passingPodChecks)
	}

	if Cfg.ExternalIP != "" {
		totalPodChecks++
		fmt.Println("----> [From SrcPod] Running externalIP connectivity check..")
		RunDstConnectivityCheck(Cfg.ExternalIP, &passingPodChecks)
	}

	if Cfg.DstSvc.SvcEndpoint.IP != "" {
		totalPodChecks++
		fmt.Println("----> [From SrcPod] Running DstSvc DNS lookup check..")
		RunK8sDNSLookupCheck(Cfg.KubeDNSService.IP, Cfg.DstSvc.Name, Cfg.DstSvc.Namespace,
			Cfg.DstSvc.SvcEndpoint.IP, &passingPodChecks)
	}

	// Change network ns back to host
	netns.Set(hostNsHandle)

	fmt.Println("----------- Pod Checks Summary -----------")
	fmt.Printf("Passed checks: %d/%d\n", passingPodChecks, totalPodChecks)
}
