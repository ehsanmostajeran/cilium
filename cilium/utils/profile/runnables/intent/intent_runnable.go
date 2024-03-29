package intent

import (
	m "github.com/cilium-team/cilium/cilium/messages"
	ucdb "github.com/cilium-team/cilium/cilium/utils/comm/db"
	up "github.com/cilium-team/cilium/cilium/utils/profile"
	upr "github.com/cilium-team/cilium/cilium/utils/profile/runnables"
	upsi "github.com/cilium-team/cilium/cilium/utils/profile/subpolicies/intent"

	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/op/go-logging"
)

const (
	Name = "intent-runnable"

	DockerSwarmCreate      = "DockerSwarmCreate"
	DockerDaemonCreate     = "DockerDaemonCreate"
	DockerDaemonStart      = "DockerDaemonStart"
	DockerDaemonRestart    = "DockerDaemonRestart"
	KubernetesMasterCreate = "KubernetesMasterCreate"
)

var (
	log = logging.MustGetLogger("cilium")

	preHookHandlers = map[string]string{
		`/docker/daemon/cilium-adapter/.*/containers/create(\?.*)?`:                            DockerDaemonCreate,
		`/docker/swarm/cilium-adapter/.*/containers/create(\?.*)?`:                             DockerSwarmCreate,
		`/kubernetes/master/cilium-adapter/api/v1/namespaces/.*/pods(\?.*)?`:                   KubernetesMasterCreate,
		`/kubernetes/master/cilium-adapter/api/v1/namespaces/.*/replicationcontrollers(\?.*)?`: KubernetesMasterCreate,
		`/kubernetes/master/cilium-adapter/api/v1/namespaces/.*/service(\?.*)?`:                KubernetesMasterCreate,
	}
	postHookHandlers = map[string]string{
		`/docker/daemon/cilium-adapter/.*/containers/.*/start(\?.*)?`:   DockerDaemonStart,
		`/docker/daemon/cilium-adapter/.*/containers/.*/restart(\?.*)?`: DockerDaemonRestart,
	}

	dockerHookHandlers = map[string]func(ucdb.Db, *upsi.Intent, *m.DockerCreateConfig) error{
		upr.PreHook + DockerDaemonCreate:   preHookDockerDaemonCreate,
		upr.PreHook + DockerSwarmCreate:    preHookDockerSwarmCreate,
		upr.PostHook + DockerDaemonStart:   postHookDockerDaemonStart,
		upr.PostHook + DockerDaemonRestart: postHookDockerDaemonStart,
	}
	kubernetesHookHandlers = map[string]func(ucdb.Db, *upsi.Intent, *m.KubernetesObjRef) error{
		upr.PreHook + KubernetesMasterCreate: preHookKubernetesMasterCreate,
	}
)

type IntentRunnable struct {
	intent *upsi.Intent
}

func (ir IntentRunnable) GetHandlers(typ string) map[string]string {
	switch typ {
	case upr.PreHook:
		return preHookHandlers
	case upr.PostHook:
		return postHookHandlers
	default:
		return nil
	}
}

func (ir IntentRunnable) DockerExec(hookType, reqType string, db ucdb.Db, cc *m.DockerCreateConfig) error {
	if f, ok := dockerHookHandlers[hookType+reqType]; ok {
		return f(db, ir.intent, cc)
	}
	return nil
}

func (ir IntentRunnable) KubernetesExec(hookType, reqType string, db ucdb.Db, cc *m.KubernetesObjRef) error {
	if f, ok := kubernetesHookHandlers[hookType+reqType]; ok {
		return f(db, ir.intent, cc)
	}
	return nil
}

func (ir IntentRunnable) GetRunnableFrom(users []up.User, policies []up.PolicySource) upr.PolicyRunnable {
	log.Debug("users %+v", users)
	isDefaultIntentConfig := true
	lastUserIntentCfgCovered := upsi.NewIntentConfig()
	usersIntentCfg := upsi.NewIntentConfig()
	usersIntentCfg.Config = upsi.Intent{}
	up.OrderUsersByDescendingID(users)
	for _, user := range users {
		log.Debug("user %+v", user)
		userPolicies := up.FilterPoliciesByUser(policies, user)
		if len(userPolicies) == 0 {
			continue
		}
		intentConfigs := up.GetIntentConfigs(userPolicies)
		upsi.OrderIntentConfigsByAscendingPriority(intentConfigs)
		for i, iConfig := range intentConfigs {
			intentConfigs[i] = iConfig.DeepCopy()
			log.Debug("Filtered userIntentConfigs: %+v", iConfig)
		}
		//TODO: If we want we can make that every rule from a specific user
		//would be overwritten, and not merged, by a prioritary order
		userIntentCfg := upsi.NewIntentConfig()
		//TODO: If we can use what is below as it is we don't
		//need to use "NewIntentConfig()" above
		userIntentCfg.Config = upsi.Intent{}
		for _, iConfig := range intentConfigs {
			log.Debug("It's covered")
			if isDefaultIntentConfig {
				log.Debug("Isn't the default")
				isDefaultIntentConfig = false
				userIntentCfg.MergeWithOverwrite(iConfig)
			} else {
				userIntentCfg.MergeWith(iConfig)
			}
			lastUserIntentCfgCovered.Config = userIntentCfg.Config
			log.Debug("current userIntentCfg.Config: %#v", userIntentCfg.Config)
		}
		log.Debug("New userIntentCfg %#v", userIntentCfg.Config)
		log.Debug("New usersIntentCfg %#v", usersIntentCfg.Config)
		usersIntentCfg.OverwriteWith(*userIntentCfg)
		log.Debug("current usersIntentCfg.Config: %#v", usersIntentCfg.Config)
	}
	/*
		We still have to create a new "default" IntentConfig
		This way we will make sure that every configuration that
		was not configured by users will have a default value set.
	*/
	finalIntentCfg := upsi.NewIntentConfig()
	finalIntentCfg.MergeWithOverwrite(*usersIntentCfg)
	log.Info("Final intent loaded: %#v", finalIntentCfg.Config)
	return IntentRunnable{intent: &finalIntentCfg.Config}
}
