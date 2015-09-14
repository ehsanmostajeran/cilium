package db

import (
	"net"

	uc "github.com/cilium-team/cilium/cilium/utils/comm"
	upl "github.com/cilium-team/cilium/cilium/utils/plugins/loadbalancer"
	up "github.com/cilium-team/cilium/cilium/utils/profile"

	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/cilium-team/go-logging"
)

var log = logging.MustGetLogger("cilium")

const defaultDB = "elastic"

const (
	TNDNSconfig              = "dnsconfig"
	TNEndpoint               = "endpoint"
	TNHAProxyconfig          = "haproxyconfig"
	TNIPsinUse               = "ipsinuse"
	TNLinksConfig            = "dockerlinks"
	TNLinksConfigTemp        = "dockerlinkstemp"
	TNPolicySource           = "policies"
	TNPortBindingsConfig     = "dockerportbindings"
	TNPortBindingsConfigTemp = "dockerportbindingstemp"
	TNUsers                  = "users"
)

func InitDb(dbType string) error {
	switch dbType {
	case "elastic":
		return InitElasticDb()
	default:
		return InitElasticDb()
	}
}

func FlushConfig(dbType string) error {
	switch dbType {
	case "elastic":
		return ElasticFlushConfig()
	default:
		return ElasticFlushConfig()
	}
}

func NewConn() (Db, error) {
	return NewElasticConn()
}

func NewConnTo(driver, ip, port string) (Db, error) {
	switch driver {
	case "elastic":
		return NewElasticConnTo(ip, port)
	default:
		return NewElasticConnTo(ip, port)
	}
}

type Db interface {
	Close()
	GetDNSConfig() (uc.DNSClient, error)
	GetDockerLinksOfContainer(string) (up.ContainerLinks, error)
	GetDockerLinksOfContainerTemp(string) (up.ContainerLinks, error)
	GetPoliciesThatCovers(map[string]string) ([]up.PolicySource, error)
	GetUsers() ([]up.User, error)
	PutDNSConfig(uc.DNSClient) error
	PutDockerLinksOfContainer(up.ContainerLinks) error
	PutDockerLinksOfContainerTemp(up.ContainerLinks) error
	PutDockerPortBindingsOfContainerTemp(up.ContainerPortBindings) error
	PutDockerPortBindingsOfContainer(up.ContainerPortBindings) error
	GetDockerPortBindingsOfContainerTemp(string) (up.ContainerPortBindings, error)
	GetDockerPortBindingsOfContainer(string) (up.ContainerPortBindings, error)
	PutUser(userName string) (bool, error)
	PutPolicy(up.PolicySource) error
	PutHAProxyConfig(upl.HAProxyClient) error
	GetHAProxyConfig() (upl.HAProxyClient, error)

	PutIP(net.IP) error
	DeleteIP(net.IP) error
	PutEndpoint(up.Endpoint) error
	DeleteEndpoint(string) error
	GetEndpoint(string) (up.Endpoint, error)
}
