package netutils

import (
	"fmt"
	"net"
	"time"

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

// Picked from https://github.com/ipsecdiagtool/ipsecdiagtool project & modified as necessary
func sendICMPMessage(dstIP string, payloadSize int, dontfragment bool) error {
	if payloadSize < defaultPayloadSize {
		payloadSize = defaultPayloadSize
	}
	//IP Layer
	ip := layers.IPv4{
		SrcIP:    net.ParseIP("0.0.0.0"),
		DstIP:    net.ParseIP(dstIP),
		Version:  4,
		TTL:      64,
		Protocol: layers.IPProtocolICMPv4,
	}
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
	if dontfragment {
		//Set "Don't Fragment"-Flag in Header
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

// SendRecvICMPMessage checks if icmp ping is successful.
// Looks at 2 icmp packets for required response
// returncode: 0 - no error. Echo reply received successfully
//			   1 - Fragmentation required
//             2 - got icmp but unknwon type
func SendRecvICMPMessage(dstIP string, payloadSize int, dontFragment bool) (int, error) {
	// Listen on all IPs
	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return -1, fmt.Errorf("Unable to open icmp socket for ping test: %v", err)
	}
	defer c.Close()

	if err := sendICMPMessage(dstIP, payloadSize, dontFragment); err != nil {
		return -1, err
	}
	rb := make([]byte, payloadSize)
	c.SetReadDeadline(time.Now().Add(time.Second * icmpTimeout))

	// Read reply. Try twice. Discard echo request if read back on 127.0.0.1 (Needed for unit tests)
	for tries := 0; tries < 2; tries++ {
		n, peer, err := c.ReadFrom(rb)
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
			fmt.Printf("    got reflection from %v with payload size:%d\n", peer, payloadSize)
			// To check if echo reply is specific to this app, check message
			// b, _ := rm.Body.Marshal(1)  // 1 : ICMPv4 type protocol number
			// icmpMessage == string(b[2:2+len(icmpMessageBody)])  // First two bytes: length of body
			return 0, nil
		case ipv4.ICMPTypeDestinationUnreachable:
			if rm.Code == layers.ICMPv4CodeFragmentationNeeded {
				// fmt.Printf("   Fragmentation required, and DF flag set\n")
				return 1, nil
			}
		default:
			//fmt.Printf("    got %+v; want echo reply\n", rm)
			// Try multiple messages
		}
	}
	// Got ICMP type but not an echo reply
	return 2, nil
}
