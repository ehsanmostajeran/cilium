// Package server contains all necessary components for cilium to process
// all powerstrip requests and process them.
package server

import (
	"errors"
	"sync"

	"github.com/cilium-team/cilium/cilium/hook/posthook"
	"github.com/cilium-team/cilium/cilium/hook/prehook"
	m "github.com/cilium-team/cilium/cilium/messages"
)

var (
	preHook      prehook.PreHook
	preHookOnce  sync.Once
	postHook     posthook.PostHook
	postHookOnce sync.Once
)

// GetHook is a factory of Hooks. Returns a valid hook based on the reqString
// value. Returns "Unsupported hook type" if the reqString doesn't match to
// any of the available hooks.
func GetHook(reqString string) (Hook, error) {
	switch reqString {
	case prehook.Type:
		preHookOnce.Do(func() {
			preHook = prehook.NewPreHook()
		})
		return preHook, nil
	case posthook.Type:
		postHookOnce.Do(func() {
			postHook = posthook.NewPostHook()
		})
		return postHook, nil
	default:
		return nil, errors.New("Unsupported hook type")
	}
}

// Hook type that has the ProcessRequest func to perfom actions on the client or
// server requests.
type Hook interface {
	ProcessRequest(baseAddr string, req string, cont []byte) (m.Response, error)
}
