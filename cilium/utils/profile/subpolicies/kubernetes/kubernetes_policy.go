package kubernetes_policy

import (
	"encoding/json"
	"sort"

	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/cilium-team/mergo"
	k8s "github.com/cilium-team/cilium/Godeps/_workspace/src/k8s.io/kubernetes/pkg/api"
)

type KubernetesConfig struct {
	ObjectReference ObjectReference `json:"object-reference,omitempty" yaml:"object-reference,omitempty"`
	BodyObj         BodyObj         `json:"body-obj,omitempty" yaml:"body-obj,omitempty"`
	Priority        int             `json:"priority,omitempty" yaml:"priority,omitempty"`
}

type ObjectReference k8s.ObjectReference
type BodyObj map[string]interface{}

// Value marshals the receiver ObjectReference into a json string.
func (or ObjectReference) Value() (string, error) {
	if data, err := json.Marshal(or); err != nil {
		return "", err
	} else {
		return string(data), err
	}
}

// Scan unmarshals the input into the receiver ObjectReference.
func (or *ObjectReference) Scan(input interface{}) error {
	return json.Unmarshal([]byte(input.(string)), or)
}

// OverwriteWith overwrites values with the ones from `other` ObjectReference if
// those ones aren't nil and if they have an `omitempty` tag.
func (or *ObjectReference) OverwriteWith(other ObjectReference) error {
	if strOther, err := other.Value(); err != nil {
		return err
	} else {
		or.Scan(strOther)
		return nil
	}
}

// Value marshals the receiver BodyObj into a json string.
func (bo BodyObj) Value() (string, error) {
	if data, err := json.Marshal(bo); err != nil {
		return "", err
	} else {
		return string(data), err
	}
}

// Scan unmarshals the input into the receiver BodyObj.
func (bo *BodyObj) Scan(input interface{}) error {
	return json.Unmarshal([]byte(input.(string)), bo)
}

// OverwriteWith overwrites values with the ones from `other` BodyObj if those
// ones aren't nil and if they have an `omitempty` tag.
func (bo *BodyObj) OverwriteWith(other BodyObj) error {
	if strOther, err := other.Value(); err != nil {
		return err
	} else {
		bo.Scan(strOther)
		return nil
	}
}

func NewKubernetesConfig() *KubernetesConfig {
	return &KubernetesConfig{}
}

// MergeWithOverwrite merges receiver's values with the `other`
// KubernetesConfig's values. `other` overwrites the receiver's values if the
// other's values are different than default's.
// Special cases:
// Priority - `other` will overwrite receiver's Priorty no matter the value.
func (kc *KubernetesConfig) MergeWithOverwrite(other KubernetesConfig) error {
	if err := mergo.MergeWithOverwrite(kc, other); err != nil {
		return err
	}
	kc.Priority = other.Priority
	return nil
}

// OverwriteWith overwrites values with the ones from `other` KubernetesConfig.
func (kc *KubernetesConfig) OverwriteWith(other KubernetesConfig) error {
	if err := kc.ObjectReference.OverwriteWith(other.ObjectReference); err != nil {
		return err
	}
	if err := kc.BodyObj.OverwriteWith(other.BodyObj); err != nil {
		return err
	}
	kc.Priority = other.Priority
	return nil
}

type orderKubernetesConfigsBy func(k1, k2 *KubernetesConfig) bool

// OrderKubernetesConfigsByAscendingPriority orders the slice of
// KubernetesConfig in ascending priority order.
func OrderKubernetesConfigsByAscendingPriority(kubernetesConfigs []KubernetesConfig) {
	ascPriority := func(d1, d2 *KubernetesConfig) bool {
		return d1.Priority < d2.Priority
	}
	orderKubernetesConfigsBy(ascPriority).sort(kubernetesConfigs)
}

func (by orderKubernetesConfigsBy) sort(kubernetesConfigs []KubernetesConfig) {
	kCS := &kubernetesConfigSorter{
		kubernetesConfigs: kubernetesConfigs,
		by:                by,
	}
	sort.Sort(kCS)
}

type kubernetesConfigSorter struct {
	kubernetesConfigs []KubernetesConfig
	by                func(d1, d2 *KubernetesConfig) bool
}

func (s *kubernetesConfigSorter) Len() int {
	return len(s.kubernetesConfigs)
}

func (s *kubernetesConfigSorter) Swap(i, j int) {
	s.kubernetesConfigs[i], s.kubernetesConfigs[j] = s.kubernetesConfigs[j], s.kubernetesConfigs[i]
}

func (s *kubernetesConfigSorter) Less(i, j int) bool {
	return s.by(&s.kubernetesConfigs[i], &s.kubernetesConfigs[j])
}
