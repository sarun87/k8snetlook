package netutils

import (
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/net/ipv6"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const (
	icmpTimeout        = 4
	icmpMessageBody    = "K8SNETLOOK-ICMP-TEST"
	defaultPayloadSize = 64
)

// SendRecvICMPMessage checks if icmp ping is successful.
// Looks at 2 icmp packets for required response
// returncode: 0 - no error. Echo reply received successfully
//			   1 - Fragmentation required
//             2 - got icmp but unknwon type
func SendRecvICMPMessage(dstIP string, payloadSize int, dontFragment bool) (int, error) {
	// Note: Does not handle IPv4 literal in IPv6. TODO later
	ip := net.ParseIP(dstIP)
	if ip.To4() != nil {
		// IPv4
		return sendRecvICMPMessageV4(dstIP, payloadSize, dontFragment)
	}
	// IPv6
	return sendRecvICMPMessageV6(dstIP, payloadSize)

}

// sendRecvICMPMessageV4 checks if icmp ping is successful.
// Looks at 2 icmp packets for required response
// returncode: 0 - no error. Echo reply received successfully
//			   1 - Fragmentation required
//             2 - got icmp but unknwon type
func sendRecvICMPMessageV4(dstIP string, payloadSize int, dontFragment bool) (int, error) {
	// Listen for ICMP reply on all IPs
	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return -1, fmt.Errorf("Unable to open icmp socket for ping test: %v", err)
	}
	defer c.Close()

	// Send ICMP message
	if err := sendICMPMessageV4(dstIP, payloadSize, dontFragment); err != nil {
		return -1, err
	}

	rb := make([]byte, payloadSize)
	c.SetReadDeadline(time.Now().Add(time.Second * icmpTimeout))

	// Read reply. Try twice. Discard echo request if read back on 127.0.0.1 (Needed for unit tests)
	for tries := 0; tries < 2; tries++ {
		n, _, err := c.ReadFrom(rb)
		if err != nil {
			if err.(net.Error).Timeout() {
				return -1, fmt.Errorf("ICMP timeout")
			}
			return -1, fmt.Errorf("Unable to read reply from icmp socket: %v", err)
		}
		// Check if read message is an ICMP message
		rm, err := icmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), rb[:n])
		if err != nil {
			return -1, fmt.Errorf("Unable to parse ICMP message:%v", err)
		}
		// Check to see if ICMP message type is ECHO reply
		switch rm.Type {
		case ipv4.ICMPTypeEchoReply:
			// Reflection received successfully. Return success
			// log.Debug("    got reflection from %v with payload size:%d\n", peer, payloadSize)
			// To check if echo reply is specific to this app, check message
			// b, _ := rm.Body.Marshal(1)  // 1 : ICMPv4 type protocol number
			// icmpMessage == string(b[2:2+len(icmpMessageBody)])  // First two bytes: length of body
			return 0, nil
		case ipv4.ICMPTypeDestinationUnreachable:
			if rm.Code == layers.ICMPv4CodeFragmentationNeeded {
				// log.Debug("   Fragmentation required, and DF flag set\n")
				return 1, nil
			}
		default:
			// log.Debug("    got %+v; want echo reply\n", rm)
			// Try multiple messages
		}
	}
	// Got ICMP type but not an echo reply
	return 2, nil
}

// sendICMPMessage sends a single ICMP packet over the wire
// Picked from https://github.com/ipsecdiagtool/ipsecdiagtool project & modified as necessary
func sendICMPMessageV4(dstIP string, payloadSize int, dontfragment bool) error {
	// If an additional payload size isn't specified, use default
	if payloadSize < defaultPayloadSize {
		payloadSize = defaultPayloadSize
	}
	// IP Layer
	ip := layers.IPv4{
		SrcIP:    net.ParseIP("0.0.0.0"),
		DstIP:    net.ParseIP(dstIP),
		Version:  4,
		TTL:      64,
		Protocol: layers.IPProtocolICMPv4,
	}
	// ICMP Layer
	icmp := layers.ICMPv4{
		TypeCode: layers.CreateICMPv4TypeCode(uint8(ipv4.ICMPTypeEcho), 0),
	}
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}
	ipHeaderBuf := gopacket.NewSerializeBuffer()
	if err := ip.SerializeTo(ipHeaderBuf, opts); err != nil {
		return err
	}
	ipHeader, err := ipv4.ParseHeader(ipHeaderBuf.Bytes())
	if err != nil {
		return err
	}
	// If dontfragment = True, set the bit in IP Header
	if dontfragment {
		ipHeader.Flags |= ipv4.DontFragment
	}
	payloadBuf := gopacket.NewSerializeBuffer()
	//Influence the payload size
	payloadbytes := []byte(icmpMessageBody)
	if payloadSize > len(icmpMessageBody) {
		padding := make([]byte, payloadSize-len(icmpMessageBody))
		payloadbytes = append(payloadbytes, padding...)
	}
	payload := gopacket.Payload(payloadbytes)
	if err := gopacket.SerializeLayers(payloadBuf, opts, &icmp, payload); err != nil {
		return err
	}
	//Send packet
	packetConn, err := net.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return err
	}
	defer packetConn.Close()
	rawConn, err := ipv4.NewRawConn(packetConn)
	if err != nil {
		return err
	}
	defer rawConn.Close()
	return rawConn.WriteTo(ipHeader, payloadBuf.Bytes(), nil)
}

// sendRecvICMPMessageV6 checks if icmp ping is successful.
// Looks at 2 icmp packets for required response
// returncode: 0 - no error. Echo reply received successfully
//			   1 - Fragmentation required
//             2 - got icmp but unknwon type
func sendRecvICMPMessageV6(dstIP string, payloadSize int) (int, error) {
	// Don't fragment is always set for IPv6. IPv6 packets cannot be fragmented
	// Listen for ICMP reply on all IPs
	c, err := icmp.ListenPacket("ip6:ipv6-icmp", "::")
	if err != nil {
		return -1, fmt.Errorf("Unable to open icmp socket for ping test: %v", err)
	}
	defer c.Close()

	// If an additional payload size isn't specified, use default
	if payloadSize < defaultPayloadSize {
		payloadSize = defaultPayloadSize
	}
	payloadbytes := []byte(icmpMessageBody)
	if payloadSize > len(icmpMessageBody) {
		padding := make([]byte, payloadSize-len(icmpMessageBody))
		payloadbytes = append(payloadbytes, padding...)
	}
	wm := icmp.Message{
		Type: ipv6.ICMPTypeEchoRequest, Code: 0,
		Body: &icmp.Echo{
			ID: os.Getpid() & 0xffff, Seq: 1,
			Data: payloadbytes,
		},
	}
	wb, err := wm.Marshal(nil)
	if err != nil {
		return -1, fmt.Errorf("Unable to convert icmp echo message to byte string: %v", err)
	}
	if _, err := c.WriteTo(wb, &net.IPAddr{IP: net.ParseIP(dstIP)}); err != nil {
		return -1, fmt.Errorf("Unable to send icmp echo request to %s:%v", dstIP, err)
	}

	rb := make([]byte, payloadSize)
	c.SetReadDeadline(time.Now().Add(time.Second * icmpTimeout))

	// Read reply. Try twice. Discard echo request if read back on 127.0.0.1 (Needed for unit tests)
	for tries := 0; tries < 2; tries++ {
		n, _, err := c.ReadFrom(rb)
		if err != nil {
			if err.(net.Error).Timeout() {
				return -1, fmt.Errorf("ICMP timeout")
			}
			return -1, fmt.Errorf("Unable to read reply from icmp socket: %v", err)
		}
		// Check if read message is an ICMP message
		rm, err := icmp.ParseMessage(ipv6.ICMPTypeEchoReply.Protocol(), rb[:n])
		if err != nil {
			return -1, fmt.Errorf("Unable to parse ICMP message:%v", err)
		}
		// Check to see if ICMP message type is ECHO reply
		switch rm.Type {
		case ipv6.ICMPTypeEchoReply:
			// Reflection received successfully. Return success
			// log.Debug("    got reflection from %v with payload size:%d\n", peer, payloadSize)
			// To check if echo reply is specific to this app, check message
			// b, _ := rm.Body.Marshal(1)  // 1 : ICMPv4 type protocol number
			// icmpMessage == string(b[2:2+len(icmpMessageBody)])  // First two bytes: length of body
			return 0, nil
		case ipv6.ICMPTypePacketTooBig:
			// log.Debug("   Fragmentation required, and DF flag set\n")
			return 1, nil
		default:
			// log.Debug("    got %+v; want echo reply\n", rm)
			// Try multiple messages
		}
	}
	// Got ICMP type but not an echo reply
	return 2, nil
}
