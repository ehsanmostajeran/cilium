package comm

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

const (
	defaultPort = "80"
)

type DNSClient struct {
	IP   string `json:"ip,omitempty" yaml:"ip,omitempty"`
	Port string `json:"port,omitempty" yaml:"port,omitempty"`
}

// Value marshals the receiver DNSClient into a json string.
func (d DNSClient) Value() (string, error) {
	if data, err := json.Marshal(d); err != nil {
		return "", err
	} else {
		return string(data), err
	}
}

// Scan unmarshals the input into the receiver DNSClient.
func (d *DNSClient) Scan(input string) error {
	return json.Unmarshal([]byte(input), d)
}

func NewDNSClientToIP(ip string) DNSClient {
	return DNSClient{IP: ip, Port: defaultPort}
}

type domainsList struct {
	Domains []string `json:"domains,omitempty"`
	Ips     []string `json:"ips,omitempty"`
}

func (dc *DNSClient) SendToDNS(domains, ips []string) error {
	ipsJSON := domainsList{Ips: ips}
	cDbytes, err := json.Marshal(ipsJSON)
	if err != nil {
		return err
	}
	for _, domain := range domains {
		addrReq := "http://" + dc.IP + ":" + dc.Port + "/domain/" + domain
		if err := sendRequest(addrReq, cDbytes); err != nil {
			return err
		}
	}
	return nil
}

func sendRequest(addrReq string, cDbytes []byte) error {
	request, err := http.NewRequest("PUT", addrReq, bytes.NewBuffer(cDbytes))
	if err != nil {
		log.Debug("request %+v", request)
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	log.Debug("response Status: %+v", resp.Status)
	log.Debug("response Headers: %+v", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	log.Debug("response Body: %+v", string(body))
	return nil
}
