package docker_policy

import (
	"encoding/json"
	"sort"

	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/cilium-team/mergo"
	d "github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

type HostConfig d.HostConfig

// Value marshals the receiver HostConfig into a json string.
func (hc HostConfig) Value() (string, error) {
	if data, err := json.Marshal(hc); err != nil {
		return "", err
	} else {
		return string(data), err
	}
}

// Scan unmarshals the input into the receiver HostConfig.
func (hc *HostConfig) Scan(input interface{}) error {
	return json.Unmarshal([]byte(input.(string)), hc)
}

// OverwriteWith overwrites values with the ones from `other` HostConfig if
// those ones aren't nil.
func (hc *HostConfig) OverwriteWith(other HostConfig) error {
	if strOther, err := other.Value(); err != nil {
		return err
	} else {
		hc.Scan(strOther)
		return nil
	}
}

type DockerConfig struct {
	Config     Config     `json:"config,omitempty" yaml:"config,omitempty"`
	HostConfig HostConfig `json:"host-config,omitempty" yaml:"host-config,omitempty"`
	Priority   int        `json:"priority,omitempty" yaml:"priority,omitempty"`
}

type Config d.Config

// Value marshals the receiver Config into a json string.
func (c Config) Value() (string, error) {
	if data, err := json.Marshal(c); err != nil {
		return "", err
	} else {
		return string(data), err
	}
}

// Scan unmarshals the input into the receiver Config.
func (c *Config) Scan(input interface{}) error {
	return json.Unmarshal([]byte(input.(string)), c)
}

// OverwriteWith overwrites values with the ones from `other` Config if those
// ones aren't nil and if they have an `omitempty` tag.
func (c *Config) OverwriteWith(other Config) error {
	if strOther, err := other.Value(); err != nil {
		return err
	} else {
		c.Scan(strOther)
		return nil
	}
}

func NewDockerConfig() *DockerConfig {
	return &DockerConfig{}
}

// MergeWithOverwrite merges receiver's values with the `other` DockerConfig's
// values. `other` overwrites the receiver's values if the other's values are
// different than default's.
// Special cases:
// Priority - `other` will overwrite receiver's Priorty no matter the value.
func (dc *DockerConfig) MergeWithOverwrite(other DockerConfig) error {
	if err := mergo.MergeWithOverwrite(dc, other); err != nil {
		return err
	}
	dc.Priority = other.Priority
	return nil
}

// OverwriteWith overwrites values with the ones from `other` DockerConfig if
// those ones aren't nil.
func (dc *DockerConfig) OverwriteWith(other DockerConfig) error {
	if err := dc.Config.OverwriteWith(other.Config); err != nil {
		return err
	}
	if err := dc.HostConfig.OverwriteWith(other.HostConfig); err != nil {
		return err
	}
	dc.Priority = other.Priority
	return nil
}

type orderDockerConfigsBy func(d1, d2 *DockerConfig) bool

// OrderDockerConfigsByAscendingPriority orders the slice of DockerConfig in
// ascending priority order.
func OrderDockerConfigsByAscendingPriority(dockerConfigs []DockerConfig) {
	ascPriority := func(d1, d2 *DockerConfig) bool {
		return d1.Priority < d2.Priority
	}
	orderDockerConfigsBy(ascPriority).sort(dockerConfigs)
}

func (by orderDockerConfigsBy) sort(dockerConfigs []DockerConfig) {
	dS := &dockerConfigSorter{
		dockerConfigs: dockerConfigs,
		by:            by,
	}
	sort.Sort(dS)
}

type dockerConfigSorter struct {
	dockerConfigs []DockerConfig
	by            func(d1, d2 *DockerConfig) bool
}

func (s *dockerConfigSorter) Len() int {
	return len(s.dockerConfigs)
}

func (s *dockerConfigSorter) Swap(i, j int) {
	s.dockerConfigs[i], s.dockerConfigs[j] = s.dockerConfigs[j], s.dockerConfigs[i]
}

func (s *dockerConfigSorter) Less(i, j int) bool {
	return s.by(&s.dockerConfigs[i], &s.dockerConfigs[j])
}
