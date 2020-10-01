package netutils

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"testing"
)

func skipTest(t *testing.T) {
	if os.Getenv("TESTIPV6") == "" {
		t.Skip("Skipping testing in CI environment")
	}
}

func TestSendRecvHTTPMessageV4(t *testing.T) {
	var body []byte
	// Use Google IPv4 Address
	url := "http://172.217.9.4:80"
	responseCode, err := SendRecvHTTPMessage(url, "", &body)
	if err != nil {
		t.Errorf("Unable to fetch response Error: %v", err)
	}
	if responseCode != http.StatusOK {
		t.Errorf("Response code not 200")
	}
	t.Logf("%s", body)
}

func TestSendRecvHTTPMessageV6(t *testing.T) {
	// Skip when run through make test
	skipTest(t)
	var body []byte
	// TODO: Use localtest server for unit testing
	// Use www.google.com IPV6 address
	url := fmt.Sprintf("http://%s", net.JoinHostPort("2607:f8b0:4005:804::2004", "80"))
	responseCode, err := SendRecvHTTPMessage(url, "", &body)
	if err != nil {
		t.Errorf("Unable to fetch response Error: %v", err)
	}
	if responseCode != http.StatusNotFound {
		t.Errorf("Expected 404. Got:%v", responseCode)
	}
}
