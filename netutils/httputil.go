package netutils

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func SendRecvHTTPMessage(url string, token string, body *[]byte) (int, error) {
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
