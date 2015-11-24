// Package messages implements all Powerstrip base messages so we can decode
// requests and build responses.
package messages

import (
	"encoding/json"
	"fmt"
)

const (
	PowerstripProtocolVersion = 2
)

type powerstripMessage struct {
	PowerstripProtocolVersion int
}

// DecodeRequest decodes the given content into the given interface.
// If it is a PowerstripMessage message it checks if the
// PowerstripProtocolVersion is the same we support.
func DecodeRequest(content []byte, val interface{}) error {
	if err := json.Unmarshal(content, val); err != nil {
		return err
	}
	if pwm, ok := val.(*powerstripMessage); ok {
		if pwm.PowerstripProtocolVersion != PowerstripProtocolVersion {
			return fmt.Errorf("Unsupported PowerstripProtocolVersion. "+
				"You have %d, we have %d\n",
				pwm.PowerstripProtocolVersion,
				PowerstripProtocolVersion)
		}
	}
	return nil
}
