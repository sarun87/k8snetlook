package netutils

import (
	"golang.org/x/sys/unix"

	"github.com/vishvananda/netlink"
)

func GetHostGatewayIP() string {
	routes, err := netlink.RouteList(nil, unix.AF_INET)
	if err != nil {
		return ""
	}
	for _, r := range routes {
		if r.Gw != nil {
			return r.Gw.String()
		}
	}
	return ""
}
