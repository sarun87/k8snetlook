package netutils

import (
	"testing"
)

func TestSendRcvICMPMessageSuccess(t *testing.T) {
	ret, err := SendRecvICMPMessage("127.0.0.1", 64, true)
	if err != nil {
		t.Errorf("ICMP reply expected from localhost. Received error: %s", err)
		return
	}
	if ret != 0 {
		t.Errorf("Expected (0, nil) but received (%d, nil).", ret)
	}
}

func TestSendRcvICMPMessageFailure(t *testing.T) {
	// Using arbitary IP for failure test
	_, err := SendRecvICMPMessage("192.192.192.192", 64, true)
	if err == nil {
		t.Errorf("Expected ICMP to arbitary IP to fail with a timeout")
	}
	if err.Error() != "ICMP timeout" {
		t.Errorf("Received a different error than expected. Received: %v", err)
	}
}
