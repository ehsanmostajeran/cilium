// Package prehook provides all necessary handling for the post hook powerstrip
// messages.
package posthook

import (
	"regexp"
	"strings"
	"time"

	m "github.com/cilium-team/cilium/cilium/messages"
	uc "github.com/cilium-team/cilium/cilium/utils/comm"
	ucdb "github.com/cilium-team/cilium/cilium/utils/comm/db"
	up "github.com/cilium-team/cilium/cilium/utils/profile"
	upr "github.com/cilium-team/cilium/cilium/utils/profile/runnables"

	dclient "github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/docker/engine-api/client"
	dtypes "github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/docker/engine-api/types"
	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/docker/engine-api/types/container"
	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/op/go-logging"
)

var log = logging.MustGetLogger("cilium")

const (
	Type       = upr.PostHook
	Docker     = "Docker"
	Kubernetes = "Kubernetes"
)

type PostHook struct {
	dbConn     ucdb.Db
	dockerConn uc.Docker
	handlers   map[string]string
}

// NewPostHook creates a PostHook instance and gets a New Connection to the
// available DB and Docker daemon.
func NewPostHook() PostHook {
	var p PostHook
	dbConn, err := ucdb.NewConn()
	if err != nil {
		log.Panicf("Error while getting a new connection to DB: %s", err)
	}
	dockerConn, err := uc.NewDockerClient()
	if err != nil {
		log.Panicf("Error while getting a new connection of DockerClient: %s", err)
	}
	p.dockerConn = dockerConn
	p.dbConn = dbConn
	p.handlers = map[string]string{}
	for _, runnable := range upr.GetRunnables() {
		runHandlers := runnable.GetHandlers(Type)
		for k, v := range runHandlers {
			p.handlers[k] = v
		}
	}
	return p
}

// ProcessRequest processes the incoming requests and returns the appropriate
// message response for that request.
func (p PostHook) ProcessRequest(baseAddr string, req string, cont []byte) (m.Response, error) {
	log.Debug("Request: %+v", req)
	return p.postHook(p.parseRequest(baseAddr, req), cont)
}

// parseRequest parses de request address and returns the proper request type
// that are understood by the runnables.
func (p PostHook) parseRequest(baseAddr string, req string) string {
	for k, v := range p.handlers {
		log.Debug("k: %+v; v: %+v; baseAddr+req : %+v", k, v, baseAddr+req)
		if match, _ := regexp.MatchString(k, baseAddr+req); match {
			return v
		}
	}
	return "Default"
}

// defaultRequest only parses the given request and returns a
// PowerstripPostHookResponse without modifying the request. Useful for debug.
func defaultRequest(cont []byte) (m.Response, error) {
	log.Debug("Posthook")

	var pphreq PowerstripPostHookRequest
	if err := m.DecodeRequest(cont, &pphreq); err != nil {
		return &PowerstripPostHookResponse{}, err
	}

	var clientBody container.Config
	if err := pphreq.UnmarshalClientBody(&clientBody); err != nil {
		return &PowerstripPostHookResponse{}, err
	}

	log.Debug("ClientBody: %+v", clientBody)

	return NewPowerstripPostHookResponse(pphreq.ServerResponse.ContentType,
			pphreq.ServerResponse.Body,
			pphreq.ServerResponse.Code),
		nil
}

// getDockerIDFrom extracts the Docker ID from a request.
func getDockerIDFrom(req string) string {
	log.Debug("")
	// Docker doesn't allow '/' so we can't make sure this won't catch docker
	// names.
	// See: https://github.com/docker/docker/blob/v1.7.1/daemon/daemon.go#L57.
	r, _ := regexp.Compile("/[[:xdigit:]]{64}")
	return strings.Replace(r.FindString(req), "/", "", -1)
}

// getDockerContainer gets the container information from docker, if the
// container doesn't exist it retries 10 more times before giving up.
func (p PostHook) getDockerContainer(containerID string) (dtypes.ContainerJSON, error) {
	var (
		dockerContainer dtypes.ContainerJSON
		err             error
	)
	attempts := 1
	maxAttempds := 10
	for attempts <= maxAttempds {
		dockerContainer, err = p.dockerConn.ContainerInspect(containerID)
		if err != nil {
			if dclient.IsErrContainerNotFound(err) {
				log.Info("Container %s not found, attempt %d/%d...", containerID, attempts, maxAttempds)
				const delay = 3 * time.Second
				time.Sleep(delay)
				attempts++
			} else {
				return dtypes.ContainerJSON{}, err
			}
		} else {
			return dockerContainer, nil
		}
	}
	return dockerContainer, err
}

// postHook takes care of preparing necessary requirements so it can call all
// Runnables available under server/utils/profile/runnables.
// It returns an error if it isn't possible to decode a request. All remaining
// failures are hidden but they are logged.
func (p PostHook) postHook(endPoint string, cont []byte) (m.Response, error) {
	log.Debug("")

	var pphreq PowerstripPostHookRequest
	if err := m.DecodeRequest(cont, &pphreq); err != nil {
		return &PowerstripPostHookResponse{}, err
	}
	log.Debug("PowerstripPostHookRequest %+v", pphreq)

	users, err := p.dbConn.GetUsers()
	if err != nil {
		// If we can't connect to DB we just sent the response without any
		// modification but we still log it.
		log.Error("Error: %+v", err)
		return defaultRequest(cont)
	}

	if strings.HasPrefix(endPoint, Docker) {
		return p.postHookDocker(endPoint, pphreq, users, cont)
	} else if strings.HasPrefix(endPoint, Kubernetes) {
		return p.postHookKubernetes(endPoint, pphreq, users, cont)
	}

	return defaultRequest(cont)
}

// postHookDocker deals with post-hook requests that are docker specific.
func (p PostHook) postHookDocker(endPoint string, pphreq PowerstripPostHookRequest,
	users []up.User, cont []byte) (m.Response, error) {
	log.Debug("")

	containerID := getDockerIDFrom(pphreq.ClientRequest.Request)

	dockerContainer, err := p.getDockerContainer(containerID)
	if err != nil {
		return &PowerstripPostHookResponse{}, err
	}

	createConfig := m.NewDockerCreateConfigFromDockerContainer(dockerContainer)

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
		log.Info("Loaded and merged policy for container '%s': %#v", createConfig.ID, runnable)
		if err = runnable.DockerExec(Type, endPoint, p.dbConn, &createConfig); err != nil {
			return &PowerstripPostHookResponse{}, err
		}
	}

	log.Debug("Response ClientBody Config: %+v", createConfig.Config)
	log.Debug("Response ClientBody HostConfig: %+v", createConfig.HostConfig)

	log.Info("Posthook successfully executed for container '%s'", createConfig.ID)
	return NewPowerstripPostHookResponse(pphreq.ServerResponse.ContentType,
			pphreq.ServerResponse.Body,
			pphreq.ServerResponse.Code),
		nil
}

// postHookKubernetes deals with post-hook requests that are kubernetes
// specific.
func (p PostHook) postHookKubernetes(endPoint string, pphreq PowerstripPostHookRequest,
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

	// We want to make sure the policies that we get are for the same Kind for
	// the request we have received or we could end up trying to apply a Pod
	// configuration into a ReplicationController for example.
	policiesKind := up.FilterPoliciesByKubernetesKind(policies, kubernetesObjRef.Kind)
	if policiesKind == nil || len(policiesKind) == 0 {
		log.Info("There aren't any policies for the giving kind.")
		return defaultRequest(cont)
	}

	for _, runnables := range upr.GetRunnables() {
		runnable := runnables.GetRunnableFrom(users, policies)
		log.Info("Loaded and merged policy for kubernetesObjRef '%s': %#v", kubernetesObjRef.Name, runnable)
		if err = runnable.KubernetesExec(Type, endPoint, p.dbConn, &kubernetesObjRef); err != nil {
			return PowerstripPostHookResponse{}, err
		}
	}

	log.Debug("Response kubernetesObjRef: %+v", kubernetesObjRef)

	log.Info("Posthook successfully executed for kubernetesObjRef '%s'", kubernetesObjRef.Name)
	return NewPowerstripPostHookResponse(pphreq.ServerResponse.ContentType,
			pphreq.ServerResponse.Body,
			pphreq.ServerResponse.Code),
		nil
}
