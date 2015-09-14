package posthook

import (
	"testing"

	m "github.com/cilium-team/cilium/cilium/messages"
)

func TestNewPowerstripPostHookResponse(t *testing.T) {
	pphr := NewPowerstripPostHookResponse(validContentType, validServerResponseBody, validCode)
	if pphr.PowerstripProtocolVersion != m.PowerstripProtocolVersion {
		t.Errorf("invalid PowerstripProtocolVersion:\ngot  %d\nwant %d",
			pphr.PowerstripProtocolVersion,
			m.PowerstripProtocolVersion)
	}
	if pphr.ModifiedServerResponse.ContentType != validContentType {
		t.Errorf("invalid ModifiedServerResponse.ContentType:\ngot  %s\nwant %s",
			pphr.ModifiedServerResponse.ContentType,
			validContentType)
	}
	if pphr.ModifiedServerResponse.Code != validCode {
		t.Errorf("invalid ModifiedServerResponse.Code:\ngot  %d\nwant %d",
			pphr.ModifiedServerResponse.Code,
			validCode)
	}
	if pphr.ModifiedServerResponse.Body != validServerResponseBody {
		t.Errorf("invalid ModifiedServerResponse.Body:\ngot  %s\nwant %s",
			pphr.ModifiedServerResponse.Body,
			validServerResponseBody)
	}
}
