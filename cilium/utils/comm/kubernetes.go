package comm

import (
	"fmt"
	"os"

	k8sClient "github.com/cilium-team/cilium/Godeps/_workspace/src/k8s.io/kubernetes/pkg/client/unversioned"
)

const (
	k8sDefaultEndpoint = "http://127.0.0.1:8080"
)

type Kubernetes struct {
	*k8sClient.Client
}

func NewKubernetesClient() (Kubernetes, error) {
	kCli := Kubernetes{}
	endpoint := os.Getenv("KUBERNETES_SERVER")
	if endpoint == "" {
		endpoint = k8sDefaultEndpoint
	}
	config := k8sClient.Config{
		Host: endpoint,
	}
	cli, err := k8sClient.New(&config)
	if err != nil {
		return kCli, fmt.Errorf("can't connect to Kubernetes API:", err)
	}
	kCli.Client = cli
	return kCli, nil
}
