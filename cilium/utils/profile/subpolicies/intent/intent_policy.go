package intent

import (
	"fmt"
	"sort"

	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/cilium-team/go-logging"
)

var log = logging.MustGetLogger("cilium")

type IntentConfig struct {
	Config   Intent `json:"config,omitempty" yaml:"config,omitempty"`
	Priority int    `json:"priority,omitempty" yaml:"priority,omitempty"`
}

func NewIntentConfig() *IntentConfig {
	return &IntentConfig{Config: *NewIntent()}
}

// GoString is the implementation of the GoStringer interface so we can easily
// print the IntentConfig fields.
func (ic IntentConfig) GoString() string {
	var retStr string
	retStr += fmt.Sprintf("IntentConfig.Priority: %+v, ", ic.Priority)
	retStr += fmt.Sprintf("IntentConfig.Config %#v", ic.Config)
	return retStr
}

// MergeWith merges receiver's values with the `other` IntentConfigs's values.
// `other` overwrites the receiver's values if its values are equal to
// default's.
// Special cases:
// Priority - `other` will replace receiver's priority if it is greater than
// receiver's.
func (ic *IntentConfig) MergeWith(other IntentConfig) {
	if ic.Priority < other.Priority {
		ic.Priority = other.Priority
	}
	ic.Config.MergeWith(other.Config)
}

// MergeWithOverwrite merges receiver's values with the `other` IntentConfig's
// values. `other` overwrites the receiver's values if the other's values are
// different than default's.
func (ic *IntentConfig) MergeWithOverwrite(other IntentConfig) {
	ic.Priority = other.Priority
	ic.Config.MergeWithOverwrite(other.Config)
}

// OverwriteWith overwrites values with the ones from `other` IntentConfig if
// those ones aren't nil and if they have an `omitempty` tag.
func (ic *IntentConfig) OverwriteWith(other IntentConfig) {
	ic.Priority = other.Priority
	ic.Config.OverwriteWith(other.Config)
}

// DeepCopy creates a deep copy of the receiver's IntentConfig.
func (ic IntentConfig) DeepCopy() IntentConfig {
	intent := IntentConfig{}
	intent.Config.OverwriteWith(ic.Config)
	intent.Priority = ic.Priority
	return intent
}

type orderIntentConfigsBy func(i1, i2 *IntentConfig) bool

// OrderDockerConfigsByAscendingPriority orders the slice of IntentConfig in
// ascending priority order.
func OrderIntentConfigsByAscendingPriority(intentConfigs []IntentConfig) {
	ascPriority := func(i1, i2 *IntentConfig) bool {
		return i1.Priority < i2.Priority
	}
	orderIntentConfigsBy(ascPriority).sort(intentConfigs)
}

func (by orderIntentConfigsBy) sort(intentConfigs []IntentConfig) {
	dS := &intentConfigSorter{
		intentConfigs: intentConfigs,
		by:            by,
	}
	sort.Sort(dS)
}

type intentConfigSorter struct {
	intentConfigs []IntentConfig
	by            func(d1, d2 *IntentConfig) bool
}

func (s *intentConfigSorter) Len() int {
	return len(s.intentConfigs)
}

func (s *intentConfigSorter) Swap(i, j int) {
	s.intentConfigs[i], s.intentConfigs[j] = s.intentConfigs[j], s.intentConfigs[i]
}

func (s *intentConfigSorter) Less(i, j int) bool {
	return s.by(&s.intentConfigs[i], &s.intentConfigs[j])
}
