package prehook

import (
	"testing"

	m "github.com/cilium-team/cilium/cilium/messages"
)

func TestNewPowerstripPreHookResponse(t *testing.T) {
	pphr := NewPowerstripPreHookResponse(validMethod, validRequest, validBody, validServerIP, validServerPort)
	if pphr.PowerstripProtocolVersion != m.PowerstripProtocolVersion {
		t.Errorf("invalid PowerstripProtocolVersion:\ngot  %d\nwant %d",
			pphr.PowerstripProtocolVersion,
			m.PowerstripProtocolVersion)
	}
	if pphr.ModifiedClientRequest.Method != validMethod {
		t.Errorf("invalid ModifiedClientRequest.Method:\ngot  %s\nwant %s",
			pphr.ModifiedClientRequest.Method,
			validMethod)
	}
	if pphr.ModifiedClientRequest.Request != validRequest {
		t.Errorf("invalid ModifiedClientRequest.Request:\ngot  %s\nwant %s",
			pphr.ModifiedClientRequest.Request,
			validRequest)
	}
	if pphr.ModifiedClientRequest.Body != validBody {
		t.Errorf("invalid ModifiedClientRequest.Body:\ngot  %s\nwant %s",
			pphr.ModifiedClientRequest.Body,
			validBody)
	}
}
