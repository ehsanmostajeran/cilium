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

type CreateConfig struct {
	Name string `json:"-" yaml:"-"`
	*d.Config
	ID         string        `json:"-" yaml:"-"`
	State      *d.State      `json:"-" yaml:"-"`
	HostConfig *d.HostConfig `json:"HostConfig,omitempty" yaml:"HostConfig,omitempty"`
}

// NewCreateConfigFromDockerContainer creates a CreateConfig from the giving
// go-dockerclient Container it makes a deep copy of the Config and HostConfig
// pointers in go-dockerclient Container
func NewCreateConfigFromDockerContainer(container d.Container) CreateConfig {
	config := *container.Config
	hostConfig := *container.HostConfig
	state := container.State
	return CreateConfig{
		Name:       container.Name,
		ID:         container.ID,
		Config:     &config,
		HostConfig: &hostConfig,
		State:      &state,
	}
}

// MergeWith merges a CreateConfig (other) with self only if its own values have
// the zero value of its type.
func (cc *CreateConfig) MergeWith(other CreateConfig) {
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

// MergeWithOverwrite merges a CreateConfig (other) with self if other's
// values aren't nil.
func (cc *CreateConfig) MergeWithOverwrite(other CreateConfig) {
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

// UnmarshalCreateClientBody unmarshals the PowerstripRequest into a
// CreateConfig.
func (p PowerstripRequest) UnmarshalCreateClientBody(cc *CreateConfig) error {
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

// Marshal2JSONStr returns on a json string format of the given CreateConfig.
func (cc *CreateConfig) Marshal2JSONStr() (string, error) {
	bytes, err := json.Marshal(cc)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Marshal2JSONStr returns on a json string format of the given d.Config.
func Marshal2JSONStr(dc d.Config) (string, error) {
	bytes, err := json.Marshal(dc)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
