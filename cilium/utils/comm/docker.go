package comm

import (
	"os"
	"strings"
	"time"

	d "github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/docker/engine-api/client"
	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/op/go-logging"
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
	//	path := os.Getenv("DOCKER_CERT_PATH")
	//	if path != "" {
	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
	cli.Client, err = d.NewClient(endpoint, "v1.21", nil, defaultHeaders)
	//	}
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

func WaitForDockerReady(dClient Docker, timeout int) error {
	for {
		_, err := dClient.ServerVersion()
		if err == nil || timeout < 0 {
			return err
		}
		timeout--
		time.Sleep(1 * time.Second)
	}
	return nil
}
