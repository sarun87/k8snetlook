package k8snetlook

import (
	"fmt"
)

const (
	TotalHostChecks = 4
)

var (
	PassingHostChecks int
)

func RunHostChecks() {
	fmt.Println("----------- Host Checks -----------")

	fmt.Println("----> [From Host] Running default gateway connectivity check..")
	RunGatewayConnectivityCheck(&PassingHostChecks)
	fmt.Println("----> [From Host] Running Kube service IP connectivity check..")
	RunKubeAPIServiceIPConnectivityCheck(&PassingHostChecks)
	fmt.Println("----> [From Host] Running Kube API Server Endpoint IP connectivity check..")
	RunKubeAPIEndpointIPConnectivityCheck(&PassingHostChecks)
	fmt.Println("----> [From Host] Running Kube API Server health check..")
	RunAPIServerHealthCheck(&PassingHostChecks)

	fmt.Println("-----------------------------------")
}
