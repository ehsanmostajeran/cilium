package profile

import (
	"net"
	"regexp"

	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/cilium-team/go-logging"
	d "github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
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

type Links []string

type ContainerLinks struct {
	Container string `json:"container,omitempty" yaml:"container,omitempty"`
	Links     Links  `json:"links,omitempty" yaml:"links,omitempty"`
}

type PortBindings map[d.Port][]d.PortBinding

type ContainerPortBindings struct {
	Container    string       `json:"container,omitempty" yaml:"container,omitempty"`
	PortBindings PortBindings `json:"port-bindings,omitempty" yaml:"port-bindings,omitempty"`
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
