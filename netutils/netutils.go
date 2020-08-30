package netutils

import (
	"fmt"

	"golang.org/x/sys/unix"

	"github.com/vishvananda/netlink"
)

// GetHostGatewayIP returns the IP of the default gw as listed in the route list
func GetHostGatewayIP() (string, error) {
	routes, err := netlink.RouteList(nil, unix.AF_INET)
	if err != nil {
		return "", err
	}
	for _, r := range routes {
		// Check if the route is marked as a gw route
		if r.Gw != nil {
			return r.Gw.String(), nil
		}
	}
	return "", fmt.Errorf("unable to find a route with default gw")
}
