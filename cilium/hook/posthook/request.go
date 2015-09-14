package posthook

import (
	m "github.com/cilium-team/cilium/cilium/messages"
)

type PowerstripPostHookRequest struct {
	m.PowerstripRequest
	ServerResponse serverResponse
}

type serverResponse struct {
	ContentType string
	Body        string
	Code        int
}

type CreateConfig struct {
	ID       string `json:"Id,omitempty" yaml:"Id,omitempty"`
	Warnings string `json:"Warnings,omitempty" yaml:"Warnings,omitempty"`
}
