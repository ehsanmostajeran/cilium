package kubernetes

import (
	m "github.com/cilium-team/cilium/cilium/messages"
	ucdb "github.com/cilium-team/cilium/cilium/utils/comm/db"
	up "github.com/cilium-team/cilium/cilium/utils/profile"
	upr "github.com/cilium-team/cilium/cilium/utils/profile/runnables"
	upsk "github.com/cilium-team/cilium/cilium/utils/profile/subpolicies/kubernetes"

	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/cilium-team/go-logging"
	k8s "github.com/cilium-team/cilium/Godeps/_workspace/src/k8s.io/kubernetes/pkg/api"
)

var log = logging.MustGetLogger("cilium")

const Name = "kubernetes-runnable"

type KubernetesRunnable struct {
	kubernetescfg *upsk.KubernetesConfig
}

func (kr KubernetesRunnable) GetRunnableFrom(users []up.User, policies []up.PolicySource) upr.PolicyRunnable {
	log.Debug("users %+v\n", users)
	log.Debug("policies %+v\n", policies)
	isDefault := true
	finalKubernetesCfg := upsk.KubernetesConfig{}
	up.OrderUsersByDescendingID(users)
	for _, user := range users {
		log.Debug("user %+v", user)
		userPolicies := up.FilterPoliciesByUser(policies, user)
		if len(userPolicies) == 0 {
			continue
		}
		kubernetesConfigs := up.GetKubernetesConfigs(userPolicies)
		upsk.OrderKubernetesConfigsByAscendingPriority(kubernetesConfigs)
		log.Debug("Filtered kubernetesConfigs: %+v", kubernetesConfigs)
		//TODO: If we want we can make that every rule from a specific use
		//will be overwritten and not merged by prioriry order
		userKubernetesCfg := *upsk.NewKubernetesConfig()
		for _, kConfig := range kubernetesConfigs {
			if isDefault {
				log.Debug("Isn't the default")
				userKubernetesCfg = kConfig
				isDefault = false
			} else {
				log.Debug("coverage before kConfig: %+v", kConfig)
				log.Debug("coverage before finalKubernetesCfg: %+v", userKubernetesCfg)
				userKubernetesCfg.MergeWithOverwrite(kConfig)
			}
			log.Debug("current finalKubernetesCfg: %+v", finalKubernetesCfg)
		}
		finalKubernetesCfg.OverwriteWith(userKubernetesCfg)
		log.Debug("current finalKubernetesCfg: %+v", finalKubernetesCfg)
	}
	log.Debug("final finalKubernetesCfg: %+v", finalKubernetesCfg)

	return KubernetesRunnable{kubernetescfg: &finalKubernetesCfg}
}

func (kr KubernetesRunnable) DockerExec(hookType, reqType string, db ucdb.Db, cc *m.DockerCreateConfig) error {
	return nil
}

func (kr KubernetesRunnable) KubernetesExec(hookType, reqType string, db ucdb.Db, cc *m.KubernetesObjRef) error {
	log.Debug("")
	return cc.MergeWithOverwrite(m.KubernetesObjRef{
		ObjectReference: (k8s.ObjectReference)(kr.kubernetescfg.ObjectReference),
		BodyObj:         (map[string]interface{})(kr.kubernetescfg.BodyObj)})
}
