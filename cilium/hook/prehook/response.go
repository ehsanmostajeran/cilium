package prehook

import (
	m "github.com/cilium-team/cilium/cilium/messages"
)

type PowerstripPreHookResponse struct {
	m.PowerstripResponse
	ModifiedClientRequest modifiedClientRequest
}

type modifiedClientRequest struct {
	m.ClientRequest
}

// NewPowerstripPreHookResponse creates and returns a new
// PowerstripPreHookResponse from the giving method, request and body.
func NewPowerstripPreHookResponse(method, request, body, serverIP string, serverPort int) *PowerstripPreHookResponse {
	p := new(PowerstripPreHookResponse)
	p.PowerstripProtocolVersion = m.PowerstripProtocolVersion
	p.ModifiedClientRequest = modifiedClientRequest{m.ClientRequest{method, request, body, serverIP, serverPort}}
	return p
}

// GetPowerstripHookResponse returns the receiver itself.
// This function is to force the inheritance of Response messages.
func (p PowerstripPreHookResponse) GetPowerstripHookResponse() interface{} {
	return p
}
