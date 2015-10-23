package intent

import (
	"encoding/json"

	m "github.com/cilium-team/cilium/cilium/messages"
	ucdb "github.com/cilium-team/cilium/cilium/utils/comm/db"
	upsi "github.com/cilium-team/cilium/cilium/utils/profile/subpolicies/intent"

	k8s "github.com/cilium-team/cilium/Godeps/_workspace/src/k8s.io/kubernetes/pkg/api"
)

func convertMapTo(obj map[string]interface{}, i interface{}) error {
	bt, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(bt, i); err != nil {
		return err
	}
	return nil
}

func preHookKubernetesMasterCreate(conn ucdb.Db, intent *upsi.Intent, kor *m.KubernetesObjRef) error {
	log.Debug("kor.Kind = %+v", kor.Kind)
	switch kor.Kind {
	case "ReplicationController":
		var rc k8s.ReplicationController
		if err := convertMapTo(kor.BodyObj, &rc); err != nil {
			return err
		}
		preHookKubernetesMasterRCCreate(conn, intent, rc)
	case "Pod":
		var pod k8s.Pod
		if err := convertMapTo(kor.BodyObj, &pod); err != nil {
			return err
		}
		preHookKubernetesMasterPodCreate(conn, intent, pod)
	case "Service":
		var service k8s.Service
		if err := convertMapTo(kor.BodyObj, &service); err != nil {
			return err
		}
		preHookKubernetesMasterServiceCreate(conn, intent, service)
	default:
		return nil
	}
	return nil
}

func preHookKubernetesMasterRCCreate(conn ucdb.Db, intent *upsi.Intent, rc k8s.ReplicationController) error {
	log.Info("I could have changed a replication controller message")
	log.Info("K8s %+v", rc)
	return nil
}

func preHookKubernetesMasterPodCreate(conn ucdb.Db, intent *upsi.Intent, p k8s.Pod) error {
	log.Info("I could have changed a pod message")
	log.Info("K8s %+v", p)
	return nil
}

func preHookKubernetesMasterServiceCreate(conn ucdb.Db, intent *upsi.Intent, s k8s.Service) error {
	log.Info("I could have changed a service message")
	log.Info("K8s %+v", s)
	return nil
}
