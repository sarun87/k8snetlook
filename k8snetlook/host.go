package k8snetlook

import (
	"fmt"
)

const (
	totalHostChecks = 4
)

var (
	passingHostChecks int
)

func RunHostChecks() {
	fmt.Println("----------- Host Checks -----------")

	fmt.Println("----> [From Host] Running default gateway connectivity check..")
	RunGatewayConnectivityCheck(&passingHostChecks)
	fmt.Println("----> [From Host] Running Kube service IP connectivity check..")
	RunKubeAPIServiceIPConnectivityCheck(&passingHostChecks)
	fmt.Println("----> [From Host] Running Kube API Server Endpoint IP connectivity check..")
	RunKubeAPIEndpointIPConnectivityCheck(&passingHostChecks)
	fmt.Println("----> [From Host] Running Kube API Server health check..")
	RunAPIServerHealthCheck(&passingHostChecks)

	fmt.Println("----------- Host Checks Summary -----------")
	fmt.Printf("Passed checks: %d/%d\n", passingHostChecks, totalHostChecks)
}
