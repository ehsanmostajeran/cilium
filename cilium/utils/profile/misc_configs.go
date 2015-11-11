package profile

import (
	"encoding/json"
	"net"
	"regexp"

	d "github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/op/go-logging"
)

var log = logging.MustGetLogger("cilium")

type Coverage struct {
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
}

func NewCoverage() *Coverage {
	return &Coverage{
		Labels: make(map[string]string),
	}
}

// Covers verifies if at least one of the labels from the receiver Coverage has
// a key that matches one of the keys from the labels variable and, if the value
// of that key matches the regex expression of the same key from the labels'
// receiver.
func (c Coverage) Covers(labels map[string]string) bool {
	log.Debug("Checking if %#v covers %#v", c, labels)
	for coverageKey, coverageValue := range c.Labels {
		for labelKey, labelValue := range labels {
			if coverageKey == labelKey {
				if match, _ := regexp.MatchString(coverageValue, labelValue); match {
					log.Debug("Covers")
					return true
				}
			}
		}
	}
	log.Debug("Doesn't cover")
	return false
}

type IPAddress net.IP

type IP struct {
	IPAddress IPAddress
}

// Value marshals the receiver IP into a json string.
func (ip IP) Value() (string, error) {
	if data, err := json.Marshal(ip); err != nil {
		return "", err
	} else {
		return string(data), err
	}
}

// Scan unmarshals the input into the receiver IP.
func (ip *IP) Scan(input string) error {
	return json.Unmarshal([]byte(input), ip)
}

type Links []string

type ContainerLinks struct {
	Container string `json:"container,omitempty" yaml:"container,omitempty"`
	Links     Links  `json:"links,omitempty" yaml:"links,omitempty"`
}

// Value marshals the receiver ContainerLinks into a json string.
func (cl ContainerLinks) Value() (string, error) {
	if data, err := json.Marshal(cl); err != nil {
		return "", err
	} else {
		return string(data), err
	}
}

// Scan unmarshals the input into the receiver ContainerLinks.
func (cl *ContainerLinks) Scan(input string) error {
	return json.Unmarshal([]byte(input), cl)
}

type PortBindings map[d.Port][]d.PortBinding

type ContainerPortBindings struct {
	Container    string       `json:"container,omitempty" yaml:"container,omitempty"`
	PortBindings PortBindings `json:"port-bindings,omitempty" yaml:"port-bindings,omitempty"`
}

// Value marshals the receiver ContainerPortBindings into a json string.
func (cpb ContainerPortBindings) Value() (string, error) {
	if data, err := json.Marshal(cpb); err != nil {
		return "", err
	} else {
		return string(data), err
	}
}

// Scan unmarshals the input into the receiver ContainerPortBindings.
func (cpb *ContainerPortBindings) Scan(input string) error {
	return json.Unmarshal([]byte(input), cpb)
}

type IPs []net.IP
type MACs []string

type Endpoint struct {
	Container string `json:"container,omitempty" yaml:"container,omitempty"`
	IPs       IPs    `json:"ips,omitempty" yaml:"ips,omitempty"`
	MACs      MACs   `json:"macs,omitempty" yaml:"macs,omitempty"`
	Node      string `json:"node,omitempty" yaml:"node,omitempty"`
	Interface string `json:"interface,omitempty" yaml:"interface,omitempty"`
	Group     int    `json:"group,omitempty" yaml:"group,omitempty"`
	BD        int    `json:"bd,omitempty" yaml:"bd,omitempty"`
	Namespace int    `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Service   string `json:"service,omitempty" yaml:"service,omitempty"`
}

// Value marshals the receiver Endpoint into a json string.
func (e Endpoint) Value() (string, error) {
	if data, err := json.Marshal(e); err != nil {
		return "", err
	} else {
		return string(data), err
	}
}

// Scan unmarshals the input into the receiver Endpoint.
func (e *Endpoint) Scan(input string) error {
	return json.Unmarshal([]byte(input), e)
}
