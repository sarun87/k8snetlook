package k8snetlook

import (
	"fmt"
	"net/http"
)

const (
	totalHostChecks = 4
)

var (
	passingHostChecks int
)

func RunHostChecks() {
	RunGatewayConnectivityCheck()
	RunKubeAPIServiceIPConnectivityCheck()
	RunKubeAPIEndpointIPConnectivityCheck()
	RunAPIServerHealthCheck()

	fmt.Println("----> Summary -----")
	fmt.Printf("Passed checks: %d/%d\n", passingHostChecks, totalHostChecks)
}

func RunGatewayConnectivityCheck() {
	fmt.Println("----> Running Gateway connectivity check..")
	pass, err := sendRecvICMPMessage(Cfg.HostGatewayIP)
	if err != nil {
		fmt.Printf("  (Failed) Error running RunGatewayConnectivityCheck. Error: %v\n", err)
		return
	}
	if pass {
		passingHostChecks++
		fmt.Println("  (Passed) Gateway connectivity check completed successfully")
	} else {
		fmt.Println("  (Failed) Gateway connectivity check failed")
	}
}

func RunKubeAPIServiceIPConnectivityCheck() {
	// TODO: Handle secure/non-secure api-servers

	// HTTP 401 return code is a successful check
	url := fmt.Sprintf("https://%s:%d", Cfg.KubeAPIService.IP, Cfg.KubeAPIService.Port)
	fmt.Println("----> Running Kube Service IP connectivity check..")
	var body []byte
	responseCode, err := sendRecvHTTPMessage(url, "", &body)
	if err != nil {
		fmt.Printf("  (Failed) Error running RunKubeAPIServiceIPConnectivityCheck. Error: %v\n", err)
		return
	}
	if responseCode == http.StatusUnauthorized {
		fmt.Println("  (Passed) Kube API Service IP connectivity check completed successfully")
	} else {
		fmt.Println("  (Passed) Kube API Service IP connectivity check returned a non 401 HTTP Code")
	}
	passingHostChecks++
}

func RunKubeAPIEndpointIPConnectivityCheck(){
	// TODO: Handle secure/non-secure api-servers

	// HTTP 401 return code is a successful check
	endpoints := getEndpointsFromService("default", "kubernetes")
	fmt.Println("----> Running Kube API Server Endpoint IP connectivity check..")
	passedCount := 0
	totalCount := len(endpoints)
	for _, ep := range endpoints {
		url := fmt.Sprintf("https://%s:%d", ep.IP, ep.Port)
		fmt.Printf("  checking endpoint: %s ........", url)
		var body []byte
		responseCode, err := sendRecvHTTPMessage(url, "", &body)
		if err != nil {
			fmt.Printf("    failed connectivity check. Error: %v\n", err)
			continue
		}
		if responseCode == http.StatusUnauthorized {
			fmt.Println("    passed connectivity check")
		} else {
			fmt.Println("    passed connectivity check. Retured non 401 code though")
		}
		passedCount++
	}
	if passedCount == totalCount {
		fmt.Println("  (Passed) Kube API Endpoint IP connectivity check")
		passingHostChecks++
	} else {
		fmt.Println("  (Failed) Kube API Endoint IP connectivity check for one or more endpoints")
	}
}

func RunAPIServerHealthCheck() {
	url := fmt.Sprintf("https://%s:%d/livez?verbose", Cfg.KubeAPIService.IP, Cfg.KubeAPIService.Port)
	fmt.Println("----> Running Kube API Server health check..")
	svcAccountToken, err := getSvcAccountToken()
	if err != nil {
		fmt.Println("  (Failed) ", err)
		return
	}
	var body []byte
	responseCode, err := sendRecvHTTPMessage(url, svcAccountToken, &body)
	if err != nil {
		fmt.Printf("    Unable to fetch api server check. Error: %v\n", err)
		return
	}
	if responseCode != http.StatusOK {
		fmt.Printf("  (Failed) status check returned non-200 http code of %d\n", responseCode)
	} else {
		fmt.Printf("%s", body)
		fmt.Println("  (Passed) please check above statuses for (ok)")
		passingHostChecks++
	}
}