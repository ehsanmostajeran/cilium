package kubernetes_policy

import (
	"encoding/json"
	"errors"
	"sort"

	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/cilium-team/mergo"
	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/op/go-logging"
	k8s "github.com/cilium-team/cilium/Godeps/_workspace/src/k8s.io/kubernetes/pkg/api"
)

var log = logging.MustGetLogger("cilium")

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

func (kc KubernetesConfig) ConvertBodyObjTo(i interface{}) error {
	jsonBytes, err := json.Marshal(kc.BodyObj)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(jsonBytes, i); err != nil {
		return err
	}
	return nil
}

func (kc KubernetesConfig) convertObjRefTo(i interface{}) error {
	jsonBytes, err := json.Marshal(kc.ObjectReference)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(jsonBytes, i); err != nil {
		return err
	}
	return nil
}

func (kc *KubernetesConfig) mapToBodyObj(i interface{}) error {
	jsonBytes, err := json.Marshal(i)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(jsonBytes, &kc.BodyObj); err != nil {
		return err
	}
	return nil
}

// MergeWithOverwrite merges receiver's values with the `other`
// KubernetesConfig's values if self's and `other`'s ObjectReference.Kind are
// the same. `other` overwrites the receiver's values if the other's values are
// different than default's.
// Special cases:
// Priority - `other` will overwrite receiver's Priorty no matter the value.
func (kc *KubernetesConfig) MergeWithOverwrite(other KubernetesConfig) error {
	if kc.ObjectReference.Kind != kc.ObjectReference.Kind {
		return nil
	}

	if err := mergo.MergeWithOverwrite(&kc.ObjectReference, other.ObjectReference); err != nil {
		return err
	}

	// We won't merge the map, we convert it to the specific object and then
	// merge it.
	switch kc.ObjectReference.Kind {
	case "Pod":
		var orig, otherPod k8s.Pod
		if err := kc.ConvertBodyObjTo(&orig); err != nil {
			return err
		}
		if err := other.ConvertBodyObjTo(&otherPod); err != nil {
			return err
		}
		if err := mergo.MergeWithOverwrite(&orig, otherPod); err != nil {
			return err
		}
		if err := kc.mapToBodyObj(orig); err != nil {
			return err
		}
	case "ReplicationController":
		var orig, otherRC k8s.ReplicationController
		if err := kc.ConvertBodyObjTo(&orig); err != nil {
			return err
		}
		if err := other.ConvertBodyObjTo(&otherRC); err != nil {
			return err
		}
		if err := mergo.MergeWithOverwrite(&orig, otherRC); err != nil {
			return err
		}
		if err := kc.mapToBodyObj(orig); err != nil {
			return err
		}
	case "Service":
		var orig, otherService k8s.Service
		if err := kc.ConvertBodyObjTo(&orig); err != nil {
			return err
		}
		if err := other.ConvertBodyObjTo(&otherService); err != nil {
			return err
		}
		if err := mergo.MergeWithOverwrite(&orig, otherService); err != nil {
			return err
		}
		if err := kc.mapToBodyObj(orig); err != nil {
			return err
		}
	default:
		return errors.New("unsupported kind")
	}

	// We have to make sure that we return the BodyObj with the values from
	// ObjectReference
	var objRefMap map[string]interface{}
	if err := kc.convertObjRefTo(&objRefMap); err != nil {
		return err
	}
	if err := mergo.MapWithOverwrite(&kc.BodyObj, objRefMap); err != nil {
		return err
	}

	kc.Priority = other.Priority
	return nil
}

// OverwriteWith overwrites values with the ones from `other` KubernetesConfig
// if those ones aren't nil and if they have an `omitempty` tag.
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
