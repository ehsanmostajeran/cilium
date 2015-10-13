package docker

import (
	m "github.com/cilium-team/cilium/cilium/messages"
	ucdb "github.com/cilium-team/cilium/cilium/utils/comm/db"
	up "github.com/cilium-team/cilium/cilium/utils/profile"
	upr "github.com/cilium-team/cilium/cilium/utils/profile/runnables"
	upsd "github.com/cilium-team/cilium/cilium/utils/profile/subpolicies/docker"

	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/cilium-team/go-logging"
	d "github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

var log = logging.MustGetLogger("cilium")

const Name = "docker-runnable"

type DockerRunnable struct {
	dockercfg *upsd.DockerConfig
}

func (dr DockerRunnable) GetRunnableFrom(users []up.User, policies []up.PolicySource) upr.PolicyRunnable {
	log.Debug("users %+v\n", users)
	log.Debug("policies %+v\n", policies)
	isDefault := true
	finalDockerCfg := upsd.DockerConfig{}
	up.OrderUsersByDescendingID(users)
	for _, user := range users {
		log.Debug("user %+v", user)
		userPolicies := up.FilterPoliciesByUser(policies, user)
		dockerConfigs := up.GetDockerConfigs(userPolicies)
		upsd.OrderDockerConfigsByAscendingPriority(dockerConfigs)
		log.Debug("Filtered dockerConfigs: %+v", dockerConfigs)
		//TODO: If we want we can make that every rule from a specific use
		//will be overwritten and not merged by prioriry order
		userDockerCfg := *upsd.NewDockerConfig()
		for _, dConfig := range dockerConfigs {
			if isDefault {
				log.Debug("Isn't the default")
				userDockerCfg = dConfig
				isDefault = false
			} else {
				log.Debug("coverage before dConfig: %+v", dConfig)
				log.Debug("coverage before finalDockerCfg: %+v", userDockerCfg)
				userDockerCfg.MergeWithOverwrite(dConfig)
			}
			log.Debug("current userDockerCfg: %+v", finalDockerCfg)
		}
		finalDockerCfg.OverwriteWith(userDockerCfg)
		log.Debug("current finalDockerCfg: %+v", finalDockerCfg)
	}
	log.Debug("final finalDockerCfg: %+v", finalDockerCfg)

	return DockerRunnable{dockercfg: &finalDockerCfg}
}

func (dr DockerRunnable) DockerExec(hookType, reqType string, db ucdb.Db, cc *m.DockerCreateConfig) error {
	cc.MergeWithOverwrite(m.DockerCreateConfig{
		Config:     (*d.Config)(&dr.dockercfg.Config),
		HostConfig: (*d.HostConfig)(&dr.dockercfg.HostConfig)})
	return nil
}
