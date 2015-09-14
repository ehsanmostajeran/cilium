package utils

import (
	"errors"
	"net"

	uc "github.com/cilium-team/cilium/cilium/utils/comm"
	upl "github.com/cilium-team/cilium/cilium/utils/plugins/loadbalancer"
	up "github.com/cilium-team/cilium/cilium/utils/profile"
)

type FakeDB struct {
	OnClose                                func()
	OnGetDNSConfig                         func() (uc.DNSClient, error)
	OnGetDockerLinksOfContainer            func(string) (up.ContainerLinks, error)
	OnGetDockerLinksOfContainerTemp        func(string) (up.ContainerLinks, error)
	OnGetPoliciesThatCovers                func(map[string]string) ([]up.PolicySource, error)
	OnGetUsers                             func() ([]up.User, error)
	OnPutDNSConfig                         func(uc.DNSClient) error
	OnPutDockerLinksOfContainer            func(up.ContainerLinks) error
	OnPutDockerLinksOfContainerTemp        func(up.ContainerLinks) error
	OnPutDockerPortBindingsOfContainerTemp func(up.ContainerPortBindings) error
	OnPutDockerPortBindingsOfContainer     func(up.ContainerPortBindings) error
	OnGetDockerPortBindingsOfContainerTemp func(string) (up.ContainerPortBindings, error)
	OnGetDockerPortBindingsOfContainer     func(string) (up.ContainerPortBindings, error)
	OnPutUser                              func(userName string) (bool, error)
	OnPutPolicy                            func(up.PolicySource) error
	OnPutHAProxyConfig                     func(upl.HAProxyClient) error
	OnGetHAProxyConfig                     func() (upl.HAProxyClient, error)
	OnPutIP                                func(net.IP) error
	OnDeleteIP                             func(net.IP) error
	OnPutEndpoint                          func(up.Endpoint) error
	OnDeleteEndpoint                       func(string) error
	OnGetEndpoint                          func(string) (up.Endpoint, error)
}

func (f FakeDB) Close() {
}

func (f FakeDB) GetUsers() ([]up.User, error) {
	if f.OnGetUsers != nil {
		return f.OnGetUsers()
	}
	return nil, errors.New("GetUsers should not have been called")
}
func (f FakeDB) GetDNSConfig() (uc.DNSClient, error) {
	if f.OnGetDNSConfig != nil {
		return f.OnGetDNSConfig()
	}
	return uc.DNSClient{}, errors.New("GetDNSConfig should not have been called")
}

func (f FakeDB) GetHAProxyConfig() (upl.HAProxyClient, error) {
	if f.OnGetHAProxyConfig != nil {
		return f.OnGetHAProxyConfig()
	}
	return upl.HAProxyClient{}, errors.New("GetHAProxyConfig should not have been called")
}

func (f FakeDB) GetDockerLinksOfContainerTemp(containerName string) (up.ContainerLinks, error) {
	if f.OnGetDockerLinksOfContainerTemp != nil {
		return f.OnGetDockerLinksOfContainerTemp(containerName)
	}
	return up.ContainerLinks{}, errors.New("GetDockerLinksOfContainerTemp should not have been called")
}

func (f FakeDB) GetDockerLinksOfContainer(containerID string) (up.ContainerLinks, error) {
	if f.OnGetDockerLinksOfContainer != nil {
		return f.OnGetDockerLinksOfContainer(containerID)
	}
	return up.ContainerLinks{}, errors.New("GetDockerLinksOfContainer should not have been called")
}

func (f FakeDB) GetEndpoint(containerID string) (up.Endpoint, error) {
	if f.OnGetEndpoint != nil {
		return f.OnGetEndpoint(containerID)
	}
	return up.Endpoint{}, errors.New("GetEndpoint should not have been called")
}

func (f FakeDB) GetDockerPortBindingsOfContainerTemp(containerID string) (up.ContainerPortBindings, error) {
	if f.OnGetDockerPortBindingsOfContainerTemp != nil {
		return f.OnGetDockerPortBindingsOfContainerTemp(containerID)
	}
	return up.ContainerPortBindings{}, errors.New("GetDockerPortBindingsOfContainerTemp should not have been called")
}

func (f FakeDB) GetDockerPortBindingsOfContainer(containerID string) (up.ContainerPortBindings, error) {
	if f.OnGetDockerPortBindingsOfContainer != nil {
		return f.OnGetDockerPortBindingsOfContainer(containerID)
	}
	return up.ContainerPortBindings{}, errors.New("GetDockerPortBindingsOfContainer should not have been called")
}

func (f FakeDB) PutUser(userName string) (bool, error) {
	if f.OnPutUser != nil {
		return f.OnPutUser(userName)
	}
	return false, errors.New("PutUser should not have been called")
}

func (f FakeDB) PutDNSConfig(dnsConfig uc.DNSClient) error {
	if f.OnPutDNSConfig != nil {
		return f.OnPutDNSConfig(dnsConfig)
	}
	return errors.New("PutDNSConfig should not have been called")
}

func (f FakeDB) PutHAProxyConfig(haProxyClient upl.HAProxyClient) error {
	if f.OnPutHAProxyConfig != nil {
		return f.OnPutHAProxyConfig(haProxyClient)
	}
	return errors.New("PutHAProxyConfig should not have been called")
}

func (f FakeDB) PutDockerLinksOfContainer(containerLinks up.ContainerLinks) error {
	if f.OnPutDockerLinksOfContainer != nil {
		return f.OnPutDockerLinksOfContainer(containerLinks)
	}
	return errors.New("PutDockerLinksOfContainer should not have been called")
}

func (f FakeDB) PutDockerLinksOfContainerTemp(containerLinks up.ContainerLinks) error {
	if f.OnPutDockerLinksOfContainerTemp != nil {
		return f.OnPutDockerLinksOfContainerTemp(containerLinks)
	}
	return errors.New("PutDockerLinksOfContainerTemp should not have been called")
}

func (f FakeDB) PutDockerPortBindingsOfContainerTemp(portBindings up.ContainerPortBindings) error {
	if f.OnPutDockerPortBindingsOfContainerTemp != nil {
		return f.OnPutDockerPortBindingsOfContainerTemp(portBindings)
	}
	return errors.New("PutDockerPortBindingsOfContainerTemp should not have been called")
}

func (f FakeDB) PutDockerPortBindingsOfContainer(portBindings up.ContainerPortBindings) error {
	if f.OnPutDockerPortBindingsOfContainer != nil {
		return f.OnPutDockerPortBindingsOfContainer(portBindings)
	}
	return errors.New("PutDockerPortBindingsOfContainer should not have been called")
}

func (f FakeDB) PutIP(ip net.IP) error {
	if f.OnPutIP != nil {
		return f.OnPutIP(ip)
	}
	return errors.New("PutIP should not have been called")
}

func (f FakeDB) DeleteIP(ip net.IP) error {
	if f.OnDeleteIP != nil {
		return f.OnDeleteIP(ip)
	}
	return errors.New("DeleteIP should not have been called")
}

func (f FakeDB) PutEndpoint(endpoint up.Endpoint) error {
	if f.OnPutEndpoint != nil {
		return f.OnPutEndpoint(endpoint)
	}
	return errors.New("PutEndpoint should not have been called")
}

func (f FakeDB) DeleteEndpoint(containerID string) error {
	if f.OnDeleteEndpoint != nil {
		return f.OnDeleteEndpoint(containerID)
	}
	return errors.New("DeleteEndpoint should not have been called")
}

func (f FakeDB) GetPoliciesThatCovers(labels map[string]string) ([]up.PolicySource, error) {
	if f.OnGetPoliciesThatCovers != nil {
		return f.OnGetPoliciesThatCovers(labels)
	}
	return nil, errors.New("GetPoliciesThatCovers should not have been called")
}

func (f FakeDB) PutPolicy(policies up.PolicySource) error {
	if f.OnPutPolicy != nil {
		return f.OnPutPolicy(policies)
	}
	return errors.New("PutPolicy should not have been called")
}
