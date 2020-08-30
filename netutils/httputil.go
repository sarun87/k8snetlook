package netutils

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// SendRecvHTTPMessage sends out a HTTP GET request to the url specified
// add token to X-Auth-Token as a Bearer token if token is specified
// Return body from GET response as part of body *[]byte
func SendRecvHTTPMessage(url string, token string, body *[]byte) (int, error) {
	// Transport Layer settings
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	// Create HTTP Client
	client := &http.Client{Transport: tr, Timeout: time.Duration(5) * time.Second}
	// Create GET request
	req, err := http.NewRequest("GET", url, nil)
	// Add Authorization header if token specified
	if token != "" {
		req.Header.Add("Authorization", "Bearer "+token)
	}
	// Send request
	res, err := client.Do(req)
	if err != nil {
		return -1, fmt.Errorf("HTTP request to %s failed: %v", url, err)
	}
	// return body of GET response
	*body, _ = ioutil.ReadAll(res.Body)
	res.Body.Close()
	return res.StatusCode, nil
}
