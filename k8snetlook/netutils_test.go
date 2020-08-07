package k8snetlook

import (
	"testing"
)

func TestSendRcvICMPMessageSuccess(t *testing.T) {
	ret, err := sendRecvICMPMessage("127.0.0.1")
	if err != nil {
		t.Errorf("ICMP reply expected from localhost. Received error: %s", err)
		return
	}
	if ret != true {
		t.Errorf("Expected (true, nil) but received (false, nil). received nil for error")
	}
}

func TestSendRcvICMPMessageFailure(t *testing.T) {
	// Using arbitary IP for failure test
	_, err := sendRecvICMPMessage("192.192.192.192")
	if err == nil {
		t.Errorf("Expected ICMP to arbitary IP to fail with a timeout")
	}
	if err.Error() != "ICMP timeout" {
		t.Errorf("Received a different error than expected. Received: %v", err)
	}
}
