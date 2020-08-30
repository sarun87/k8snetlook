package k8snetlook

import (
	"encoding/json"

	log "github.com/sarun87/k8snetlook/logutil"
)

// PrintReport prints the summary of the checks
func PrintReport() {
	log.Info("----------------k8snetlook-----------------")
	log.Info("")
	log.Info("----> Host Checks")
	var hostPassCount, podPassCount int
	for _, ch := range allChecks.HostChecks {
		symbol := "fail"
		if ch.Success {
			symbol = " ok "
			hostPassCount++
		}
		log.Info(" %s\t%s\n", symbol, ch.Name)
	}
	log.Info("")
	if len(allChecks.PodChecks) > 0 {
		log.Info("----> Pod Checks (from within SrcPod)")
		for _, ch := range allChecks.PodChecks {
			symbol := "fail"
			if ch.Success {
				symbol = " ok "
				podPassCount++
			}
			log.Info(" %s\t%s\n", symbol, ch.Name)
		}
	}
	log.Info("")
	log.Info("-------------Summary-------------------")
	log.Info("")
	log.Info("  Host Checks: %d/%d\n", hostPassCount, len(allChecks.HostChecks))
	log.Info("   Pod Checks: %d/%d\n", podPassCount, len(allChecks.PodChecks))
	log.Info(" Total Checks: %d/%d\n", hostPassCount+podPassCount, len(allChecks.HostChecks)+len(allChecks.PodChecks))
	log.Info("")
	log.Info("---------------------------------------")
}

// GetReportJSON returns allChecks object as a JSON string
func GetReportJSON() string {
	jsonResult, err := json.Marshal(allChecks)
	if err != nil {
		return "Unable to return results as JSON string"
	}
	return string(jsonResult)
}
