package messages

import (
	"encoding/json"
	"net/url"

	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/cilium-team/mergo"
	d "github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

type PowerstripRequest struct {
	powerstripMessage
	Type          string
	ClientRequest ClientRequest
}

type ClientRequest struct {
	Method  string
	Request string
	Body    string
}

type DockerCreateConfig struct {
	Name string `json:"-" yaml:"-"`
	*d.Config
	ID         string        `json:"-" yaml:"-"`
	State      *d.State      `json:"-" yaml:"-"`
	HostConfig *d.HostConfig `json:"HostConfig,omitempty" yaml:"HostConfig,omitempty"`
}

// NewDockerCreateConfigFromDockerContainer creates a CreateConfig from the
// giving go-dockerclient Container it makes a deep copy of the Config and
// HostConfig pointers in go-dockerclient Container
func NewDockerCreateConfigFromDockerContainer(container d.Container) DockerCreateConfig {
	config := *container.Config
	hostConfig := *container.HostConfig
	state := container.State
	return DockerCreateConfig{
		Name:       container.Name,
		ID:         container.ID,
		Config:     &config,
		HostConfig: &hostConfig,
		State:      &state,
	}
}

// MergeWith merges a DockerCreateConfig (other) with self only if its own
// values have the zero value of its type.
func (cc *DockerCreateConfig) MergeWith(other DockerCreateConfig) {
	if cc.Name == "" {
		cc.Name = other.Name
	}
	if cc.Config != nil {
		mergo.Merge(cc.Config, other.Config)
	} else {
		cc.Config = other.Config
	}
	if cc.HostConfig != nil {
		mergo.Merge(cc.HostConfig, other.HostConfig)
	} else {
		cc.HostConfig = other.HostConfig
	}
}

// MergeWithOverwrite merges a DockerCreateConfig (other) with self if other's
// values aren't nil.
func (cc *DockerCreateConfig) MergeWithOverwrite(other DockerCreateConfig) {
	if other.Name != "" {
		cc.Name = other.Name
	}
	if cc.Config != nil {
		mergo.MergeWithOverwrite(cc.Config, other.Config)
	} else {
		if other.Config != nil {
			config := *other.Config
			cc.Config = &config
		}
	}
	if cc.HostConfig != nil {
		mergo.MergeWithOverwrite(cc.HostConfig, other.HostConfig)
	} else {
		if other.HostConfig != nil {
			hostConfig := *other.HostConfig
			cc.HostConfig = &hostConfig
		}
	}
}

// UnmarshalDockerCreateClientBody unmarshals the PowerstripRequest into a
// DockerCreateConfig.
func (p PowerstripRequest) UnmarshalDockerCreateClientBody(cc *DockerCreateConfig) error {
	if p.ClientRequest.Body == "" {
		return nil
	}
	err := json.Unmarshal([]byte(p.ClientRequest.Body), cc)
	if err != nil {
		return err
	}
	urlreq, err := url.ParseRequestURI(p.ClientRequest.Request)
	if err != nil {
		return err
	}
	// The "/" is because docker inserts them as well on the container's names
	cc.Name = "/" + urlreq.Query().Get("name")

	return nil
}

// UnmarshalClientBody unmarshals the PowerstripRequest into a go-dockerclient
// Config.
func (p PowerstripRequest) UnmarshalClientBody(config *d.Config) error {
	if p.ClientRequest.Body == "" {
		return nil
	}
	if err := json.Unmarshal([]byte(p.ClientRequest.Body), config); err != nil {
		return err
	}
	return nil
}

// Marshal2JSONStr returns on a json string format of the given DockerCreateConfig.
func (cc *DockerCreateConfig) Marshal2JSONStr() (string, error) {
	bytes, err := json.Marshal(cc)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
