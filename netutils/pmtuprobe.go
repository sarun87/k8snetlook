package netutils

import (
	"net"

	log "github.com/sarun87/k8snetlook/logutil"
)

const (
	ipHeaderSize   = 20
	icmpHeaderSize = 8
	maxMTUSize     = 9000
)

// PMTUProbeToDestIP runs ICMP pings to destination with varying payload size
// and returns the highest MTU that works. Currently works for IPv4 only
func PMTUProbeToDestIP(dstIP string) (int, error) {
	var maxOkMTU int
	minPayloadSize, maxPayloadSize := (icmpHeaderSize + ipHeaderSize), (maxMTUSize - icmpHeaderSize - ipHeaderSize)

	res, err := SendRecvICMPMessage(dstIP, minPayloadSize, true)
	if err != nil || res == 1 {
		return -1, err
	}
	maxOkMTU = minPayloadSize
	// Use binary search to check for working mtu
	for minPayloadSize <= maxPayloadSize {
		midPayloadSize := (minPayloadSize + maxPayloadSize) / 2
		log.Debug("Trying with mtu size:%d\n", midPayloadSize)
		ret, err := SendRecvICMPMessage(dstIP, midPayloadSize, true)
		if err != nil {
			//fmt.Println("Received error:", err)
			if e, ok := err.(*net.OpError); ok {
				// Check if send failed due to Message too long (i.e. paylod > src if mtu)
				if e.Err.Error() == "sendmsg: message too long" {
					// log.Debug("WARN: Cannot send packet size larger than iface MTU")
					// Go lower
					maxPayloadSize = midPayloadSize - 1
					continue
				}
			} else {
				// Some other error. Not handling this as part of mtu probing
				return -1, err
			}
		}
		if ret == 1 {
			// if return code is 1, icmp reply had fragmentation required. So go loweer
			maxPayloadSize = midPayloadSize - 1
		} else {
			// successful icmp response. Go higher
			log.Debug("  got reflection from %s with payload: %d\n", dstIP, midPayloadSize)
			minPayloadSize = midPayloadSize + 1
			maxOkMTU = midPayloadSize
		}
	}
	return maxOkMTU + ipHeaderSize + icmpHeaderSize, nil
}
