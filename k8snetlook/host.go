package k8snetlook

import (
	log "github.com/sarun87/k8snetlook/logutil"
)

// RunHostChecks runs checks from host network namespace
func RunHostChecks() {
	log.Debug("----------- Host Checks -----------")

	log.Debug("----> [From Host] Running default gateway connectivity check..")
	pass, err := RunGatewayConnectivityCheck()
	allChecks.HostChecks = append(allChecks.HostChecks, Check{
		Name: "Default gateway connectivity check", Success: pass, ErrorMsg: err})

	log.Debug("----> [From Host] Running Kube service IP connectivity check..")
	pass, err = RunKubeAPIServiceIPConnectivityCheck()
	allChecks.HostChecks = append(allChecks.HostChecks, Check{
		Name: "Kube service IP connectivity check", Success: pass, ErrorMsg: err})

	log.Debug("----> [From Host] Running Kube API Server Endpoint IP connectivity check..")
	pass, err = RunKubeAPIEndpointIPConnectivityCheck()
	allChecks.HostChecks = append(allChecks.HostChecks, Check{
		Name: "Kube API Server Endpoint IP connectivity check", Success: pass, ErrorMsg: err})

	log.Debug("----> [From Host] Running Kube API Server health check..")
	pass, err = RunAPIServerHealthCheck()
	allChecks.HostChecks = append(allChecks.HostChecks, Check{
		Name: "Kube API Server health check", Success: pass, ErrorMsg: err})

	log.Debug("-----------------------------------")
}
