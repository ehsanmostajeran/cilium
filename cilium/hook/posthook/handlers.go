// Package prehook provides all necessary handling for the post hook powerstrip
// messages.
package posthook

import (
	"regexp"
	"strings"

	m "github.com/cilium-team/cilium/cilium/messages"
	uc "github.com/cilium-team/cilium/cilium/utils/comm"
	ucdb "github.com/cilium-team/cilium/cilium/utils/comm/db"
	upr "github.com/cilium-team/cilium/cilium/utils/profile/runnables"

	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/cilium-team/go-logging"
	d "github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

var log = logging.MustGetLogger("cilium")

const (
	Type = "post-hook"
)

var handlers = map[string]string{
	`/daemon/cilium-adapter/.*/containers/create(\?.*)?`:     "DaemonCreate",
	`/daemon/cilium-adapter/.*/containers/.*/start(\?.*)?`:   "DaemonStart",
	`/daemon/cilium-adapter/.*/containers/.*/restart(\?.*)?`: "DaemonRestart",
}

type PostHook struct {
	dbConn     ucdb.Db
	dockerConn uc.Docker
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
	return p
}

// ProcessRequest processes the incoming requests and returns the appropriate
// message response for that request.
func (p PostHook) ProcessRequest(baseAddr string, req string, cont []byte) (m.Response, error) {
	log.Debug("Request: %+v", req)
	return p.postHook(parseRequest(baseAddr, req), cont)
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
// PowerstripPostHookResponse without modifying the request. Useful for debug.
func defaultRequest(cont []byte) (m.Response, error) {
	log.Debug("Posthook")

	var pphreq PowerstripPostHookRequest
	if err := m.DecodeRequest(cont, &pphreq); err != nil {
		return &PowerstripPostHookResponse{}, err
	}

	var clientBody d.Config
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

	containerID := getDockerIDFrom(pphreq.ClientRequest.Request)
	dockerContainer, err := p.dockerConn.InspectContainer(containerID)
	if err != nil {
		return &PowerstripPostHookResponse{}, err
	}

	createConfig := m.NewCreateConfigFromDockerContainer(*dockerContainer)

	users, err := p.dbConn.GetUsers()
	if err != nil {
		// If we can't connect to DB we just sent the response without any
		// modification but we still log it.
		log.Error("Error: %+v", err)
		return defaultRequest(cont)
	}

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
		log.Info("Loaded and merged policy for container %s: %#v", createConfig.ID, runnable)
		if err = runnable.Exec(Type, endPoint, p.dbConn, &createConfig); err != nil {
			return &PowerstripPostHookResponse{}, err
		}
	}

	log.Debug("Response ClientBody Config: %+v", createConfig.Config)
	log.Debug("Response ClientBody HostConfig: %+v", createConfig.HostConfig)

	log.Info("Posthook executed successfully for container %s", createConfig.ID)
	return NewPowerstripPostHookResponse(pphreq.ServerResponse.ContentType,
			pphreq.ServerResponse.Body,
			pphreq.ServerResponse.Code),
		nil
}
