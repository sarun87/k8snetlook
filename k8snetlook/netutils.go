package k8snetlook

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/vishvananda/netlink"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/sys/unix"
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
	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return false, fmt.Errorf("Unable to open icmp socket for ping test: %v", err)
	}
	defer c.Close()
	wm := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: os.Getpid() & 0xffff, Seq: 1,
			Data: []byte("K8SNETLOOK-R-U-THERE"),
		},
	}
	wb, err := wm.Marshal(nil)
	if err != nil {
		return false, fmt.Errorf("Unable to convert icmp echo message to byte string: %v", err)
	}
	fmt.Printf("    ping: Sending echo request to %s   ......", dstIP)
	if _, err := c.WriteTo(wb, &net.IPAddr{IP: net.ParseIP(dstIP)}); err != nil {
		return false, fmt.Errorf("Unable to send icmp echo request to %s:%v", dstIP, err)
	}
	rb := make([]byte, 1500)
	n, peer, err := c.ReadFrom(rb)
	if err != nil {
		return false, fmt.Errorf("Unable to read reply from icmp socket: %v", err)
	}
	rm, err := icmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), rb[:n])
	if err != nil {
		return false, fmt.Errorf("Unable to parse ICMP message:%v", err)
	}
	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		fmt.Printf("    got reflection from %v\n", peer)
		return true, nil
	default:
		fmt.Printf("    got %+v; want echo reply\n", rm)
		return false, nil
	}
}

func sendRecvHTTPMessage(url string, token string, body *[]byte) (int, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr, Timeout: time.Duration(5) * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Bearer "+token)
	res, err := client.Do(req)
	//res, err := client.Get(url)
	if err != nil {
		return -1, fmt.Errorf("HTTP request to %s failed: %v", url, err)
	}
	*body, _ = ioutil.ReadAll(res.Body)
	res.Body.Close()
	return res.StatusCode, nil
}
