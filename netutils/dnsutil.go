package netutils

import (
	"errors"

	"github.com/miekg/dns"
)

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

// RunDNSLookupUsingCustomResolver sends a dns lookup for hostFQDN
// to DNS server specified by nameserver.
// run dns lookup using github.com/miekg/dns
// code referenced from: https://github.com/bogdanovich/dns_resolver
// nameserver string format: "ip:port"
// hostFQDN string format: "abc.def.ghi."
func RunDNSLookupUsingCustomResolver(nameserver, hostFQDN string) ([]string, error) {
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
