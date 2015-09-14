package prehook

import (
	m "github.com/cilium-team/cilium/cilium/messages"
)

type PowerstripPreHookRequest struct {
	m.PowerstripRequest
}
