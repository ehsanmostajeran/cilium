package posthook

import (
	m "github.com/cilium-team/cilium/cilium/messages"
)

type PowerstripPostHookRequest struct {
	m.PowerstripRequest
	ServerResponse m.ServerResponse
}

type DockerCreateConfig struct {
	ID       string `json:"Id,omitempty" yaml:"Id,omitempty"`
	Warnings string `json:"Warnings,omitempty" yaml:"Warnings,omitempty"`
}
