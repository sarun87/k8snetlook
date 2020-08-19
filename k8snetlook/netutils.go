package k8snetlook

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/sys/unix"

	"github.com/miekg/dns"
	"github.com/vishvananda/netlink"
)

const (
	icmpTimeout = 4
)

func getHostGatewayIP() string {
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

func sendRecvICMPMessage(dstIP string) (bool, error) {
	// Listen on all IPs
	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return false, fmt.Errorf("Unable to open icmp socket for ping test: %v", err)
	}
	defer c.Close()
	// Create icmp message
	wm := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: os.Getpid() & 0xffff, Seq: 1,
			Data: []byte("K8SNETLOOK-R-U-THERE"),
		},
	}
	// convert icmp message to byte string
	wb, err := wm.Marshal(nil)
	if err != nil {
		return false, fmt.Errorf("Unable to convert icmp echo message to byte string: %v", err)
	}

	fmt.Printf("    ping: Sending echo request to %s   ......", dstIP)
	if _, err := c.WriteTo(wb, &net.IPAddr{IP: net.ParseIP(dstIP)}); err != nil {
		return false, fmt.Errorf("Unable to send icmp echo request to %s:%v", dstIP, err)
	}
	rb := make([]byte, 1500)
	c.SetReadDeadline(time.Now().Add(time.Second * icmpTimeout))
	// Read reply. Try twice. Discard echo request if read back on 127.0.0.1 (Needed for unit tests)
	for tries := 0; tries < 2; tries++ {
		n, peer, err := c.ReadFrom(rb)
		if err != nil {
			if err.(net.Error).Timeout() {
				return false, fmt.Errorf("ICMP timeout")
			}
			return false, fmt.Errorf("Unable to read reply from icmp socket: %v", err)
		}
		// Check if read message is an ICMP message
		rm, err := icmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), rb[:n])
		if err != nil {
			return false, fmt.Errorf("Unable to parse ICMP message:%v", err)
		}
		// Check to see if ICMP message type is ECHO reply
		switch rm.Type {
		case ipv4.ICMPTypeEchoReply:
			fmt.Printf("    got reflection from %v\n", peer)
			return true, nil
		default:
			fmt.Printf("    got %+v; want echo reply\n", rm)
			// Try multiple messages
		}
	}
	// Got ICMP type but not an echo reply
	return false, nil
}

func sendRecvHTTPMessage(url string, token string, body *[]byte) (int, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr, Timeout: time.Duration(5) * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if token != "" {
		req.Header.Add("Authorization", "Bearer "+token)
	}
	res, err := client.Do(req)
	if err != nil {
		return -1, fmt.Errorf("HTTP request to %s failed: %v", url, err)
	}
	*body, _ = ioutil.ReadAll(res.Body)
	res.Body.Close()
	return res.StatusCode, nil
}

// Does not work, not sure why :(
/*func runDNSLookupUsingCustomResolver(dnsFQDN) ([]net.IPAddr, error) {
	// Create a custom resolver since /etc/resolv.conf is the host's configuration
	// and not the pods config. This is because the below code is executed within
	// the Pod netns but not the file system of the pod.
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(10000),
			}
			//return net.Dial(network, net.JoinHostPort(dnsServerIP, "53"))
			return d.DialContext(ctx, network, net.JoinHostPort(dnsServerIP, "53"))
		},
	}
	// Lookup for svcname.svcnamespace eg: kubernetes.default
	dnsFQDN := fmt.Sprintf("%s.%s", dstSvcName, dstSvcNamespace)
	ips, err := r.LookupIPAddr(context.Background(), dnsFQDN)
	if err != nil {
		fmt.Printf("  (Failed) dns lookup to %s failed. Error: %v\n", dnsFQDN, err)
		return nil, err
	}
	return ips, nil
}
*/

// run dns lookup using github.com/miekg/dns
// code referenced from: https://github.com/bogdanovich/dns_resolver
// nameserver string format: "ip:port"
// hostFQDN string format: "abc.def.ghi."
func runDNSLookupUsingCustomResolver(nameserver, hostFQDN string) ([]string, error) {
	// TODO: Add retries

	// Create DNS Message with single question
	msg := &dns.Msg{
		MsgHdr: dns.MsgHdr{
			Id:               dns.Id(),
			RecursionDesired: true,
		},
		Question: []dns.Question{{Name: dns.Fqdn(hostFQDN), Qtype: dns.TypeA, Qclass: dns.ClassINET}},
	}

	// Send question to nameserver and wait for answer
	in, err := dns.Exchange(msg, nameserver)
	if err != nil {
		return nil, err
	}

	result := []string{}

	if in != nil && in.Rcode != dns.RcodeSuccess {
		// Return error code
		return result, errors.New(dns.RcodeToString[in.Rcode])
	}

	// Fetch IP Addresses in DNS Answer and return
	for _, record := range in.Answer {
		if t, ok := record.(*dns.A); ok {
			result = append(result, t.A.String())
		}
	}
	return result, err
}
