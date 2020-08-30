package k8snetlook

import (
	"fmt"
	"runtime"

	log "github.com/sarun87/k8snetlook/logutil"
	"github.com/vishvananda/netns"
)

// RunPodChecks runs checks from within the Pod network namespace
func RunPodChecks() {
	// 1. Switch to SrcPod network namespace
	// 2. Run checks from within pod network namespace
	// 3. Switch back to host network namespace

	log.Debug("----------- Pod Checks -----------")

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
		log.Error("Unable to switch to pod network namespace:%v\n", err)
		return
	}

	// Execute checks from within the Pod network ns
	log.Debug("----> [From SrcPod] Running Kube service IP connectivity check..")
	pass, err := RunKubeAPIServiceIPConnectivityCheck()
	allChecks.PodChecks = append(allChecks.PodChecks, Check{
		Name: "Kube service IP connectivity check", Success: pass, ErrorMsg: err})

	log.Debug("----> [From SrcPod] Running Kube API Server Endpoint IP connectivity check..")
	pass, err = RunKubeAPIEndpointIPConnectivityCheck()
	allChecks.PodChecks = append(allChecks.PodChecks, Check{
		Name: "Kube API Server Endpoint IP connectivity check", Success: pass, ErrorMsg: err})

	log.Debug("----> [From SrcPod] Running default gateway connectivity check..")
	pass, err = RunGatewayConnectivityCheck()
	allChecks.PodChecks = append(allChecks.PodChecks, Check{
		Name: "Default gateway connectivity check", Success: pass, ErrorMsg: err})

	log.Debug("----> [From SrcPod] Running DNS lookup test (kubernetes.default)..")
	pass, err = RunK8sDNSLookupCheck(Cfg.KubeDNSService.IP, "kubernetes", "default",
		Cfg.KubeAPIService.IP)
	allChecks.PodChecks = append(allChecks.PodChecks, Check{
		Name: "DNS lookup check for kubernetes.default", Success: pass, ErrorMsg: err})

	if Cfg.DstPod.IP != "" {
		log.Debug("----> [From SrcPod] Running DstPod connectivity check..")
		pass, err = RunDstConnectivityCheck(Cfg.DstPod.IP)
		allChecks.PodChecks = append(allChecks.PodChecks, Check{
			Name: "DstPod connectivity check", Success: pass, ErrorMsg: err})

		log.Debug("----> [From SrcPod] Running pmtud check for dstIP..")
		pass, err = RunMTUProbeToDstIPCheck(Cfg.DstPod.IP)
		allChecks.PodChecks = append(allChecks.PodChecks, Check{
			Name: "pMTU check for DstIP", Success: pass, ErrorMsg: err})
	}

	if Cfg.ExternalIP != "" {
		log.Debug("----> [From SrcPod] Running externalIP connectivity check..")
		pass, err = RunDstConnectivityCheck(Cfg.ExternalIP)
		allChecks.PodChecks = append(allChecks.PodChecks, Check{
			Name: "ExternalIP connectivity check", Success: pass, ErrorMsg: err})

		log.Debug("----> [From SrcPod] Running pmtud check for externalIP..")
		pass, err = RunMTUProbeToDstIPCheck(Cfg.ExternalIP)
		allChecks.PodChecks = append(allChecks.PodChecks, Check{
			Name: "pMTU check for ExternalIP", Success: pass, ErrorMsg: err})
	}

	if Cfg.DstSvc.SvcEndpoint.IP != "" {
		log.Debug("----> [From SrcPod] Running DstSvc DNS lookup check..")
		pass, err = RunK8sDNSLookupCheck(Cfg.KubeDNSService.IP, Cfg.DstSvc.Name, Cfg.DstSvc.Namespace,
			Cfg.DstSvc.SvcEndpoint.IP)
		allChecks.PodChecks = append(allChecks.PodChecks, Check{
			Name: "DNS lookup for DstSvc", Success: pass, ErrorMsg: err})
	}

	// Change network ns back to host
	netns.Set(hostNsHandle)
}
