package comm

import (
	"fmt"
	"os"
	"strings"

	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/cilium-team/go-logging"
	d "github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	dsamalba "github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/samalba/dockerclient"
)

var log = logging.MustGetLogger("cilium")

const (
	defaultEndpoint = "unix:///var/run/docker.sock"
)

type Docker struct {
	*d.Client
}

func NewDockerClient() (cli Docker, err error) {
	endpoint := os.Getenv("DOCKER_HOST")
	if endpoint == "" {
		endpoint = defaultEndpoint
	}
	path := os.Getenv("DOCKER_CERT_PATH")
	if path != "" {
		ca := fmt.Sprintf("%s/ca.pem", path)
		cert := fmt.Sprintf("%s/cert.pem", path)
		key := fmt.Sprintf("%s/key.pem", path)
		cli.Client, err = d.NewTLSClient(endpoint, cert, key, ca)
	} else {
		cli.Client, err = d.NewClient(endpoint)
	}
	return
}

func NewDockerClientSamalba() (cli *dsamalba.DockerClient, err error) {
	endpoint := os.Getenv("DOCKER_HOST")
	if endpoint == "" {
		endpoint = defaultEndpoint
	}
	path := os.Getenv("DOCKER_CERT_PATH")
	if path != "" {
		//		ca := fmt.Sprintf("%s/ca.pem", path)
		//		cert := fmt.Sprintf("%s/cert.pem", path)
		//		key := fmt.Sprintf("%s/key.pem", path)
		log.Warning("DOCKER_CERT_PATH not available yet.")
		cli, err = dsamalba.NewDockerClient(endpoint, nil)
	} else {
		cli, err = dsamalba.NewDockerClient(endpoint, nil)
	}
	return
}

func SplitLink(link string) (container, alias string) {
	split := strings.Split(link, ":")
	switch len(split) {
	case 2:
		return "/" + split[0], split[1]
	case 1:
		return "/" + split[0], ""
	default:
		return "/" + link, ""
	}
}
