package k8snetlook

import (
	"fmt"
	"net"
	"net/http"
	"strconv"

	log "github.com/sarun87/k8snetlook/logutil"
	"github.com/sarun87/k8snetlook/netutils"
)

// RunGatewayConnectivityCheck checks connectivity to default gw
func RunGatewayConnectivityCheck() (bool, error) {
	log.Debug("Sending ICMP message to gw IP:%s", Cfg.HostGatewayIP)
	pass, err := netutils.SendRecvICMPMessage(Cfg.HostGatewayIP, 64, true)
	if err != nil {
		log.Debug("  (Failed) Error running RunGatewayConnectivityCheck. Error: %v\n", err)
		return false, err
	}
	if pass == 0 {
		log.Debug("  (Passed) Gateway connectivity check completed successfully")
		return true, nil
	}
	log.Debug("  (Failed) Gateway connectivity check failed")
	return false, nil
}

// RunDstConnectivityCheck checks connectivity to destination specified by dstIP
func RunDstConnectivityCheck(dstIP string) (bool, error) {
	pass, err := netutils.SendRecvICMPMessage(dstIP, 64, true)
	if err != nil {
		log.Debug("  (Failed) Error running connectivity check to %s. Error: %v\n", dstIP, err)
		return false, err
	}
	if pass == 0 {
		log.Debug("  (Passed) Connectivity check to destination %s completed successfully\n", dstIP)
		return true, nil
	}
	log.Debug("  (Failed) Connectivity check to destination %s failed\n", dstIP)
	return false, nil
}

// RunKubeAPIServiceIPConnectivityCheck checks connectivity to K8s api service via clusterIP
func RunKubeAPIServiceIPConnectivityCheck() (bool, error) {
	// TODO: Handle secure/non-secure api-servers
	// HTTP 401 return code is a successful check
	url := fmt.Sprintf("https://%s", net.JoinHostPort(Cfg.KubeAPIService.IP, strconv.Itoa(int(Cfg.KubeAPIService.Port))))
	var body []byte
	responseCode, err := netutils.SendRecvHTTPMessage(url, "", &body)
	if err != nil {
		log.Debug("  (Failed) Error running RunKubeAPIServiceIPConnectivityCheck. Error: %v\n", err)
		return false, err
	}
	if responseCode == http.StatusUnauthorized {
		log.Debug("  (Passed) Kube API Service IP connectivity check completed successfully")
	} else {
		log.Debug("  (Passed) Kube API Service IP connectivity check returned a non 401 HTTP Code")
	}
	return true, nil
}

// RunKubeAPIEndpointIPConnectivityCheck checks connectivity to k8s api server via each endpoint (nodeIP)
func RunKubeAPIEndpointIPConnectivityCheck() (bool, error) {
	// TODO: Handle secure/non-secure api-servers
	// HTTP 401 return code is a successful check
	endpoints := getEndpointsFromService("default", "kubernetes")
	totalCount := len(endpoints)
	if totalCount == 0 {
		return false, fmt.Errorf("could not fetch endpoints for k8s api server")
	}
	passedCount := 0
	for _, ep := range endpoints {
		url := fmt.Sprintf("https://%s", net.JoinHostPort(ep.IP, strconv.Itoa(int(ep.Port))))
		log.Debug("  checking endpoint: %s ........", url)
		var body []byte
		responseCode, err := netutils.SendRecvHTTPMessage(url, "", &body)
		if err != nil {
			log.Debug("    failed connectivity check. Error: %v\n", err)
			continue
		}
		if responseCode == http.StatusUnauthorized {
			log.Debug("    passed connectivity check")
		} else {
			log.Debug("    passed connectivity check. Retured non 401 code though")
		}
		passedCount++
	}
	if passedCount == totalCount {
		log.Debug("  (Passed) Kube API Endpoint IP connectivity check")
		return true, nil
	}
	log.Debug("  (Failed) Kube API Endoint IP connectivity check for one or more endpoints")
	return false, nil
}

// RunAPIServerHealthCheck checks api server health using livez endpoint
func RunAPIServerHealthCheck() (bool, error) {
	url := fmt.Sprintf("https://%s/livez?verbose", net.JoinHostPort(Cfg.KubeAPIService.IP, strconv.Itoa(int(Cfg.KubeAPIService.Port))))
	svcAccountToken, err := getSvcAccountToken()
	if err != nil {
		log.Debug("  (Failed) ", err)
		return false, err
	}
	var body []byte
	responseCode, err := netutils.SendRecvHTTPMessage(url, svcAccountToken, &body)
	if err != nil {
		log.Debug("    Unable to fetch api server check. Error: %v\n", err)
		return false, err
	}
	if responseCode != http.StatusOK {
		log.Debug("  (Failed) status check returned non-200 http code of %d\n", responseCode)
		return false, nil
	}
	log.Debug("%s", body)
	log.Debug("  (Passed) please check above statuses for (ok)")
	return true, nil
}

// RunK8sDNSLookupCheck checks DNS lookup functionality for a given K8s service
func RunK8sDNSLookupCheck(dnsServerIP, dstSvcName, dstSvcNamespace, dstSvcExpectedIP string) (bool, error) {
	dnsServerURL := net.JoinHostPort(dnsServerIP, "53")
	// TODO: Fetch domain information from cluster
	svcfqdn := fmt.Sprintf("%s.%s.svc.cluster.local.", dstSvcName, dstSvcNamespace)
	ips, err := netutils.RunDNSLookupUsingCustomResolver(dnsServerURL, svcfqdn)
	if err != nil {
		log.Debug("  (Failed) Unable to run dns lookup to %s, error: %v\n", svcfqdn, err)
		return false, err
	}
	// Check if the resolved IP matches with the IP reported by K8s
	for _, ip := range ips {
		if ip == dstSvcExpectedIP {
			log.Debug("  (Passed) dns lookup to %s returned: %s. Expected: %s\n", svcfqdn, ip, dstSvcExpectedIP)
			return true, nil
		}
	}
	log.Debug("  (Failed) Lookup of %s retured: %v, expected: %s\n", svcfqdn, ips, dstSvcExpectedIP)
	return false, nil
}

// RunMTUProbeToDstIPCheck checks path-MTU by probing the traffic path using icmp messages
func RunMTUProbeToDstIPCheck(dstIP string) (bool, error) {
	supportedMTU, err := netutils.PMTUProbeToDestIP(dstIP)
	if err != nil {
		log.Debug("   (Failed) Unable to run pmtud for %s. Error: %v\n", dstIP, err)
		return false, err
	}
	log.Debug("   Maximum MTU that works for destination IP: %s is %d\n", dstIP, supportedMTU)
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Debug("   Unable to fetch network interfaces. Error: %v\n", err)
		return false, err
	}
	for _, iface := range ifaces {
		// If loopback device, skip
		if iface.Flags&net.FlagLoopback == net.FlagLoopback {
			continue
		}
		if iface.MTU > supportedMTU {
			log.Debug("  Iface %s has higher mtu than supported path mtu. Has: %d, should be less than %d\n", iface.Name, iface.MTU, supportedMTU)
		}
	}
	// TODO: Check for the outgoing interface mtu and compare
	log.Info("   (Passed) Retured MTU for destination IP: %s = %d\n", dstIP, supportedMTU)
	return true, nil
}

// RunDstSvcEndpointsConnectivityCheck checks connectivity from SrcPod to all IPs provided to this checker
func RunDstSvcEndpointsConnectivityCheck(endpoints []Endpoint) (bool, error) {
	totalCount := len(endpoints)
	if totalCount == 0 {
		return false, fmt.Errorf("could not fetch endpoints for k8s api server")
	}
	passedCount := 0
	for _, ep := range endpoints {
		log.Debug("  checking endpoint: %s ........", ep.IP)
		pass, err := netutils.SendRecvICMPMessage(ep.IP, 64, true)
		if err != nil {
			log.Debug("  (Failed) Error running connectivity check to %s. Error: %v\n", ep.IP, err)
		}
		if pass == 0 {
			log.Debug("  (Passed) Connectivity check to destination %s completed successfully\n", ep.IP)
			passedCount++
		} else {
			log.Debug("  (Failed) Connectivity check to destination %s failed\n", ep.IP)
		}
	}
	if passedCount == totalCount {
		log.Debug("  (Passed) DstSvc Endpoints IP connectivity check")
		return true, nil
	}
	log.Debug("  (Failed) DstSvc Endoints IP connectivity check for one or more endpoints")
	return false, nil
}
