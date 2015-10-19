// Package prehook provides all necessary handling for the pre hook powerstrip
// messages.
package prehook

import (
	"regexp"
	"strings"

	m "github.com/cilium-team/cilium/cilium/messages"
	ucdb "github.com/cilium-team/cilium/cilium/utils/comm/db"
	up "github.com/cilium-team/cilium/cilium/utils/profile"
	upr "github.com/cilium-team/cilium/cilium/utils/profile/runnables"

	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/cilium-team/go-logging"
	d "github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

var log = logging.MustGetLogger("cilium")

const (
	Type       = "pre-hook"
	Docker     = "Docker"
	Kubernetes = "Kubernetes"
)

var handlers = map[string]string{
	`/docker/swarm/cilium-adapter/.*/containers/create(\?.*)?`:           upr.DockerSwarmCreate,
	`/docker/daemon/cilium-adapter/.*/containers/create(\?.*)?`:          upr.DockerDaemonCreate,
	`/kubernetes/master/cilium-adapter/api/v1/namespaces/.*/pods(\?.*)?`: upr.KubernetesMasterPodCreate,
}

type PreHook struct {
	dbConn ucdb.Db
}

// NewPreHook creates a PreHook instance and gets a New Connection to the
// available DB.
func NewPreHook() PreHook {
	var p PreHook
	dbConn, err := ucdb.NewConn()
	if err != nil {
		log.Panicf("Error while getting a new connection to DB: %s", err)
	}
	p.dbConn = dbConn
	return p
}

// ProcessRequest processes the incoming requests and returns the appropriate
// message response for that request.
func (p PreHook) ProcessRequest(baseAddr string, req string, cont []byte) (m.Response, error) {
	log.Debug("Request: %+v", req)
	return p.preHook(parseRequest(baseAddr, req), cont)
}

// parseRequest parses de request address and returns the proper request type
// that are understood by the runnables.
func parseRequest(baseAddr string, req string) string {
	for k, v := range handlers {
		if match, _ := regexp.MatchString(k, baseAddr+req); match {
			return v
		}
	}
	return "Default"
}

// defaultRequest only parses the given request and returns a
// PowerstripPreHookResponse without modifying the request. Useful for debug.
func defaultRequest(cont []byte) (m.Response, error) {
	log.Debug("Prehook")

	var pphreq PowerstripPreHookRequest
	if err := m.DecodeRequest(cont, &pphreq); err != nil {
		return PowerstripPreHookResponse{}, err
	}

	var clientBody d.Config
	if err := pphreq.UnmarshalClientBody(&clientBody); err != nil {
		return PowerstripPreHookResponse{}, err
	}

	log.Debug("ClientBody: %+v", clientBody)

	return NewPowerstripPreHookResponse(pphreq.ClientRequest.Method,
			pphreq.ClientRequest.Request,
			pphreq.ClientRequest.Body),
		nil
}

// preHook takes care of preparing necessary requirements so it can call all
// Runnables available under server/utils/profile/runnables.
// It returns an error if it isn't possible to decode a request. All remaining
// failures are hidden but they are logged.
func (p PreHook) preHook(endPoint string, cont []byte) (m.Response, error) {
	log.Debug("")

	var pphreq PowerstripPreHookRequest
	if err := m.DecodeRequest(cont, &pphreq); err != nil {
		return PowerstripPreHookResponse{}, err
	}
	log.Debug("PowerstripPreHookRequest %+v", pphreq)

	users, err := p.dbConn.GetUsers()
	if err != nil {
		// If we can't connect to DB we just sent the response without any
		// modification but we still log it.
		log.Error("Error: %+v", err)
		return defaultRequest(cont)
	}

	if strings.HasPrefix(endPoint, Docker) {
		return p.preHookDocker(endPoint, pphreq, users, cont)
	} else if strings.HasPrefix(endPoint, Kubernetes) {
		return p.preHookKubernetes(endPoint, pphreq, users, cont)
	}

	return defaultRequest(cont)
}

// preHookDocker deals with pre-hook requests that are docker specific.
func (p PreHook) preHookDocker(endPoint string, pphreq PowerstripPreHookRequest,
	users []up.User, cont []byte) (m.Response, error) {
	log.Debug("")

	var createConfig m.DockerCreateConfig
	if err := pphreq.UnmarshalDockerCreateClientBody(&createConfig); err != nil {
		log.Error("Error: %+v", err)
		return defaultRequest(cont)
	}

	log.Debug("ClientBody: %+v", createConfig)
	log.Debug("ClientBody.Config: %+v", createConfig.Config)
	log.Debug("ClientBody.HostConfig: %+v", createConfig.HostConfig)

	if createConfig.Config == nil || createConfig.Config.Labels == nil {
		log.Info("Request has empty config or empty labels.")
		return defaultRequest(cont)
	}

	policies, err := p.dbConn.GetPoliciesThatCovers(createConfig.Config.Labels)
	if err != nil {
		log.Error("Error: %+v", err)
		return defaultRequest(cont)
	}
	if policies == nil || len(policies) == 0 {
		log.Info("There aren't any policies for the giving labels.")
		return defaultRequest(cont)
	}

	for _, runnables := range upr.GetRunnables() {
		runnable := runnables.GetRunnableFrom(users, policies)
		log.Info("Loaded and merged policy for container %s: %#v", createConfig.Name, runnable)
		if err = runnable.DockerExec(Type, endPoint, p.dbConn, &createConfig); err != nil {
			return PowerstripPreHookResponse{}, err
		}
	}

	log.Debug("Response ClientBody Config: %+v", createConfig.Config)
	log.Debug("Response ClientBody HostConfig: %+v", createConfig.HostConfig)

	respCreateConfig, err := createConfig.Marshal2JSONStr()
	if err != nil {
		return PowerstripPreHookResponse{}, err
	}

	log.Info("Response created for container %s: %#v", createConfig.Name, respCreateConfig)
	return NewPowerstripPreHookResponse(pphreq.ClientRequest.Method,
			pphreq.ClientRequest.Request,
			respCreateConfig),
		nil
}

// preHookKubernetes deals with pre-hook requests that are kubernetes specific.
func (p PreHook) preHookKubernetes(endPoint string, pphreq PowerstripPreHookRequest,
	users []up.User, cont []byte) (m.Response, error) {
	log.Debug("")

	var kubernetesObjRef m.KubernetesObjRef
	if err := pphreq.UnmarshalKubernetesObjRefClientBody(&kubernetesObjRef); err != nil {
		log.Error("Error: %+v", err)
		return defaultRequest(cont)
	}

	log.Debug("kubernetesObjRef: %+v", kubernetesObjRef)

	labels, err := kubernetesObjRef.GetLabels()
	if err != nil {
		log.Info("Request has empty labels: %+v", err)
		return defaultRequest(cont)
	}

	policies, err := p.dbConn.GetPoliciesThatCovers(labels)
	if err != nil {
		log.Error("Error: %+v", err)
		return defaultRequest(cont)
	}
	if policies == nil || len(policies) == 0 {
		log.Info("There aren't any policies for the giving labels.")
		return defaultRequest(cont)
	}

	for _, runnables := range upr.GetRunnables() {
		runnable := runnables.GetRunnableFrom(users, policies)
		log.Info("Loaded and merged policy for kubernetesObjRef '%s': %#v", kubernetesObjRef.Name, runnable)
		if err = runnable.KubernetesExec(Type, endPoint, p.dbConn, &kubernetesObjRef); err != nil {
			return PowerstripPreHookResponse{}, err
		}
	}

	log.Debug("Response kubernetesObjRef: %+v", kubernetesObjRef)

	respCreateConfig, err := kubernetesObjRef.Marshal2JSONStr()
	if err != nil {
		return PowerstripPreHookResponse{}, err
	}

	log.Info("Response created for kubernetesObjRef '%s': %#v", kubernetesObjRef.Name, respCreateConfig)
	return NewPowerstripPreHookResponse(pphreq.ClientRequest.Method,
			pphreq.ClientRequest.Request,
			respCreateConfig),
		nil
}
