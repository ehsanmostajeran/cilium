package runnables

import (
	"fmt"

	m "github.com/cilium-team/cilium/cilium/messages"
	ucdb "github.com/cilium-team/cilium/cilium/utils/comm/db"
	up "github.com/cilium-team/cilium/cilium/utils/profile"
)

const (
	DockerDaemonCreate = "DockerDaemonCreate"
	DockerSwarmCreate  = "DockerSwarmCreate"
)

type Runnables map[string]PolicyRunnable

var (
	runnables Runnables
)

func init() {
	runnables = make(Runnables)
}

func Register(name string, policyRun PolicyRunnable) error {
	if _, ok := runnables[name]; ok {
		return fmt.Errorf("\"%s\" is already registered, please use a different name", name)
	}
	runnables[name] = policyRun
	return nil
}

func GetRunnables() Runnables {
	return runnables
}

type PolicyRunnable interface {
	GetRunnableFrom(users []up.User, policies []up.PolicySource) PolicyRunnable
	DockerExec(hookType, reqType string, db ucdb.Db, cc *m.DockerCreateConfig) error
}
