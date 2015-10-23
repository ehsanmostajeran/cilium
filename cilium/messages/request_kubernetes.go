package messages

import (
	"encoding/json"
	"errors"

	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/cilium-team/mergo"
	k8s "github.com/cilium-team/cilium/Godeps/_workspace/src/k8s.io/kubernetes/pkg/api"
)

type KubernetesObjRef struct {
	k8s.ObjectReference
	BodyObj map[string]interface{}
}

// UnmarshalKubernetesObjRefClientBody unmarshals the PowerstripRequest into a
// KubernetesObjRef.
func (p PowerstripRequest) UnmarshalKubernetesObjRefClientBody(cc *KubernetesObjRef) error {
	if p.ClientRequest.Body == "" {
		return nil
	}
	err := json.Unmarshal([]byte(p.ClientRequest.Body), &cc.ObjectReference)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(p.ClientRequest.Body), &cc.BodyObj)
	if err != nil {
		return err
	}
	return nil
}

func (kor KubernetesObjRef) convertTo(i interface{}) error {
	jsonBytes, err := json.Marshal(kor.BodyObj)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(jsonBytes, i); err != nil {
		return err
	}
	return nil
}

// GetLabels returns the labels of the supported kinds of Kubernetes messages.
func (kor *KubernetesObjRef) GetLabels() (map[string]string, error) {
	switch kor.Kind {
	case "Pod":
		var pod k8s.Pod
		if err := kor.convertTo(&pod); err != nil {
			return nil, err
		}
		return pod.Labels, nil
	case "ReplicationController":
		var rc k8s.ReplicationController
		if err := kor.convertTo(&rc); err != nil {
			return nil, err
		}
		return rc.Labels, nil
	case "Service":
		var s k8s.Service
		if err := kor.convertTo(&s); err != nil {
			return nil, err
		}
		return s.Labels, nil
	}
	return nil, errors.New("unsupported kind")
}

// Marshal2JSONStr returns on a json string format of the given
// KubernetesObjRef.BodyObj.
func (kor *KubernetesObjRef) Marshal2JSONStr() (string, error) {
	bytes, err := json.Marshal(kor.BodyObj)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// MergeWithOverwrite merges a KubernetesObjRef (other) with self.
func (kor *KubernetesObjRef) MergeWithOverwrite(other KubernetesObjRef) {
	mergo.MergeWithOverwrite(&kor.ObjectReference, other.ObjectReference)
	mergo.MergeWithOverwrite(&kor.BodyObj, other.BodyObj)
}
