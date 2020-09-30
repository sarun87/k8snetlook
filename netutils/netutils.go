package netutils

import (
	"fmt"

	"golang.org/x/sys/unix"

	log "github.com/sarun87/k8snetlook/logutil"
	"github.com/vishvananda/netlink"
)

// GetHostGatewayIP returns the IP of the default gw as listed in the route list
func GetHostGatewayIP() (string, error) {
	gwIP, err := getHostGatewayIPUsingFamily(unix.AF_INET)
	if err != nil {
		log.Debug("Error: %v", err)
		// If we are here, there was a problem returning list v4 routes. Try v6 routes
		gwIP, err = getHostGatewayIPUsingFamily(unix.AF_INET6)
		if err != nil {
			// v6 gw route not found either. Return nil
			return "", err
		}
	}
	return gwIP, nil
}

func getHostGatewayIPUsingFamily(family int) (string, error) {
	routes, err := netlink.RouteList(nil, family)
	if err != nil {
		return "", err
	}
	for _, r := range routes {
		// Check if the route is marked as a gw route
		if r.Gw != nil && r.Dst == nil {
			return r.Gw.String(), nil
		}
	}
	return "", fmt.Errorf("unable to find a route with default gw")
}
