package posthook

import (
	m "github.com/cilium-team/cilium/cilium/messages"
)

type PowerstripPostHookResponse struct {
	m.PowerstripResponse
	ModifiedServerResponse modifiedServerResponse
}

type modifiedServerResponse struct {
	ContentType string
	Body        string
	Code        int
}

// NewPowerstripPostHookResponse creates and returns a new
// PowerstripPostHookResponse from the giving contentType, body and code.
func NewPowerstripPostHookResponse(contentType, body string, code int) *PowerstripPostHookResponse {
	p := new(PowerstripPostHookResponse)
	p.PowerstripProtocolVersion = m.PowerstripProtocolVersion
	p.ModifiedServerResponse = modifiedServerResponse{contentType, body, code}
	return p
}

// GetPowerstripHookResponse returns the receiver itself.
// This function is to force the inheritance of Response messages.
func (p PowerstripPostHookResponse) GetPowerstripHookResponse() interface{} {
	return p
}
