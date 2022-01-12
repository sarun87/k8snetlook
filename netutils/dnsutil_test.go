package netutils

import (
	"testing"
)

func TestDNSLookupGoogle(t *testing.T) {
	res, err := RunDNSLookupUsingCustomResolver("8.8.8.8:53", "www.google.com")
	if err != nil {
		t.Errorf("Unable to resolve Google using Google DNS! Error:%v", err)
	}
	t.Logf("Resolved IPs:%v\n", res)
	// Should have an IPv4 and an IPv6 address for Google
	if len(res) < 2 {
		t.Errorf("DNS resolution for www.google.com into ipv4 and ipv6 addresses failed with Google DNS")
	}
}
