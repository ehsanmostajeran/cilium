package compose2kube

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	m "github.com/cilium-team/cilium/cilium/messages"
	uc "github.com/cilium-team/cilium/cilium/utils/comm"
	ucdb "github.com/cilium-team/cilium/cilium/utils/comm/db"
	up "github.com/cilium-team/cilium/cilium/utils/profile"
	upr "github.com/cilium-team/cilium/cilium/utils/profile/runnables"

	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/cilium-team/go-logging"
	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/docker/docker/pkg/parsers/filters"
	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/gorilla/mux"
	k8s "github.com/cilium-team/cilium/Godeps/_workspace/src/k8s.io/kubernetes/pkg/api"
	"github.com/cilium-team/cilium/Godeps/_workspace/src/k8s.io/kubernetes/pkg/api/resource"
	"github.com/cilium-team/cilium/Godeps/_workspace/src/k8s.io/kubernetes/pkg/api/unversioned"
)

const (
	Name = "compose2kube-runnable"

	Compose2KubeCreateContainer = "Compose2KubeCreateContainer"
	Compose2KubeGetContainers   = "Compose2KubeGetContainers"
	Compose2KubeGetContainer    = "Compose2KubeGetContainer"
)

var (
	log = logging.MustGetLogger("cilium")

	preHookHandlers = map[string]string{
		`/compose2kube/master/cilium-adapter/.*/containers/create(\?.*)?`:  Compose2KubeCreateContainer,
		`/compose2kube/master/cilium-adapter/.*/containers/json(\?.*)?`:    Compose2KubeGetContainers,
		`/compose2kube/master/cilium-adapter/.*/containers/.*/json(\?.*)?`: Compose2KubeGetContainer,
	}
	postHookHandlers = map[string]string{}

	compose2kubeHookHanlders = map[string]func(cr *m.ClientRequest, sr *m.ServerResponse, kor *m.KubernetesObjRef) error{
		upr.PreHook + Compose2KubeCreateContainer: preHookCreateContainer,
		upr.PreHook + Compose2KubeGetContainers:   preHookGetContainers,
		upr.PreHook + Compose2KubeGetContainer:    preHookGetContainer,
	}
	dc, _ = uc.NewDockerClient()
)

type Compose2KubeRunnable struct {
}

func (c2kr Compose2KubeRunnable) GetHandlers(typ string) map[string]string {
	switch typ {
	case upr.PreHook:
		return preHookHandlers
	case upr.PostHook:
		return postHookHandlers
	default:
		return nil
	}
}

func (c2kr Compose2KubeRunnable) GetRunnableFrom(users []up.User, policies []up.PolicySource) upr.PolicyRunnable {
	return Compose2KubeRunnable{}
}

func (c2kr Compose2KubeRunnable) DockerExec(hookType, reqType string, db ucdb.Db, cc *m.DockerCreateConfig) error {
	return nil
}

func (c2kr Compose2KubeRunnable) KubernetesExec(hookType, reqType string, db ucdb.Db, cc *m.KubernetesObjRef) error {
	return nil
}

func (c2kr Compose2KubeRunnable) Compose2KubeExec(hookType, reqType string, cr *m.ClientRequest, sr *m.ServerResponse, kor *m.KubernetesObjRef) error {
	if f, ok := compose2kubeHookHanlders[hookType+reqType]; ok {
		return f(cr, sr, kor)
	}
	return nil
}

func (c2kr Compose2KubeRunnable) Kube2ComposeExec(hookType, reqType string, cr *m.ClientRequest, sr *m.ServerResponse, kor *m.KubernetesObjRef) error {
	return nil
}

//
type dummyResp struct {
	io.Writer
	h    int
	Vars map[string]string
}

func newDummyResp() http.ResponseWriter {
	return &dummyResp{Writer: &bytes.Buffer{}}
}

func (w dummyResp) Header() http.Header               { return make(http.Header) }
func (w dummyResp) WriteHeader(h int)                 { w.h = h }
func (w dummyResp) Write(b []byte) (int, error)       { return len(b), nil }
func (w *dummyResp) WriteVars(vars map[string]string) { w.Vars = vars }

func getVars(pattern, reqStr string) (map[string]string, error) {
	dummyRequest, err := http.NewRequest("GET", "http://localhost"+reqStr, nil)
	if err != nil {
		return nil, err
	}
	myHandler := func(w http.ResponseWriter, r *http.Request) {
		d := w.(*dummyResp)
		d.WriteVars(mux.Vars(r))
	}
	r := mux.NewRouter()
	r.HandleFunc(pattern, myHandler).Methods("GET")
	res := newDummyResp()
	r.ServeHTTP(res, dummyRequest)
	resDummyResp := res.(*dummyResp)
	return resDummyResp.Vars, nil
}

//

func preHookGetContainer(cr *m.ClientRequest, sr *m.ServerResponse, kor *m.KubernetesObjRef) error {
	log.Debug("")
	dCEndpoint := dc.Client.Endpoint()
	protocol := strings.Split(dCEndpoint, ":")[0]
	vars, err := getVars("/{ip}:{port}/", strings.Replace(dCEndpoint, protocol+":/", "", -1))
	if err != nil {
		return err
	}
	cr.ServerIP = vars["ip"]
	port, err := strconv.ParseInt(vars["port"], 10, 32)
	if err != nil {
		port = 80
	}
	cr.ServerPort = int(port)
	return nil
}

func preHookGetContainers(cr *m.ClientRequest, sr *m.ServerResponse, kor *m.KubernetesObjRef) error {
	log.Debug("")
	u, err := url.Parse(cr.Request)
	if err != nil {
		return err
	}

	kubernetesRequest := "/api/v1/namespaces/default/pods"
	v := u.Query()
	if f, ok := v["filters"]; ok {
		if len(f) != 0 {
			if a, err := filters.FromParam(f[0]); err == nil {
				if labels, ok := a["label"]; ok {
					kubernetesRequest += "?labelSelector="
					kubernetesRequest += strings.Join(labels, ",")
				}
			}
		}
	}
	cr.Request = kubernetesRequest
	return nil
}

func replicationController(name string, pod *k8s.Pod) *k8s.ReplicationController {
	return &k8s.ReplicationController{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		ObjectMeta: k8s.ObjectMeta{
			Name:   name,
			Labels: pod.ObjectMeta.Labels,
		},
		Spec: k8s.ReplicationControllerSpec{
			Replicas: 1,
			Selector: pod.ObjectMeta.Labels,
			Template: &k8s.PodTemplateSpec{
				ObjectMeta: k8s.ObjectMeta{
					Labels: pod.ObjectMeta.Labels,
				},
				Spec: pod.Spec,
			},
		},
	}
}

func convertDockerContainerToRC(dockerContainer m.DockerCreateConfig) (string, []byte) {
	labels := map[string]string{}
	for k, v := range dockerContainer.Labels {
		if len(v) > 63 {
			labels[k] = v[:63]
		} else {
			labels[k] = v
		}
	}
	name := dockerContainer.Name
	if len(name) == 0 {
		name = labels["com.docker.compose.project"] + "-" + labels["com.docker.compose.service"]
	}
	pod := &k8s.Pod{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: k8s.ObjectMeta{
			Name:   dockerContainer.Name,
			Labels: labels,
		},
		Spec: k8s.PodSpec{
			Containers: []k8s.Container{
				{
					Name:  name,
					Image: dockerContainer.Image,
					Args:  dockerContainer.Cmd,
					Resources: k8s.ResourceRequirements{
						Limits: k8s.ResourceList{},
					},
				},
			},
		},
	}
	if dockerContainer.HostConfig.CPUShares != 0 {
		pod.Spec.Containers[0].Resources.Limits[k8s.ResourceCPU] = *resource.NewQuantity(dockerContainer.HostConfig.CPUShares, "decimalSI")
	}

	if dockerContainer.HostConfig.Memory != 0 {
		pod.Spec.Containers[0].Resources.Limits[k8s.ResourceMemory] = *resource.NewQuantity(dockerContainer.HostConfig.Memory, "decimalSI")
	}
	// Configure the environment variables
	var environment []k8s.EnvVar
	for _, envs := range dockerContainer.Config.Env {
		value := strings.Split(envs, "=")
		environment = append(environment, k8s.EnvVar{Name: value[0], Value: value[1]})
	}

	pod.Spec.Containers[0].Env = environment

	// Configure the container ports.
	var ports []k8s.ContainerPort
	for _, port := range dockerContainer.Config.PortSpecs {
		portNumber, err := strconv.Atoi(port)
		if err != nil {
			log.Error("Invalid container port %s for service %s", port, dockerContainer.Name)
		}
		ports = append(ports, k8s.ContainerPort{ContainerPort: portNumber})
	}

	pod.Spec.Containers[0].Ports = ports

	// Configure the container restart policy.
	var (
		rc      *k8s.ReplicationController
		objType string
		data    []byte
		err     error
	)
	switch dockerContainer.HostConfig.RestartPolicy.Name {
	case "", "always":
		objType = "replicationcontrollers"
		rc = replicationController(name, pod)
		pod.Spec.RestartPolicy = k8s.RestartPolicyAlways
		data, err = json.MarshalIndent(rc, "", "  ")
	case "no", "false":
		objType = "pod"
		pod.Spec.RestartPolicy = k8s.RestartPolicyNever
		data, err = json.MarshalIndent(pod, "", "  ")
	case "on-failure":
		objType = "replicationcontrollers"
		rc = replicationController(name, pod)
		pod.Spec.RestartPolicy = k8s.RestartPolicyOnFailure
		data, err = json.MarshalIndent(rc, "", "  ")
	default:
		log.Error("Unknown restart policy %s for service %s", dockerContainer.HostConfig.RestartPolicy.Name, dockerContainer.Name)
	}

	if err != nil {
		log.Error("Failed to marshal the replication controller: %v", err)
	}
	return objType, data
}

func preHookCreateContainer(cr *m.ClientRequest, sr *m.ServerResponse, kor *m.KubernetesObjRef) error {
	log.Debug("")

	var dockerContainer m.DockerCreateConfig
	if err := json.Unmarshal([]byte(cr.Body), &dockerContainer); err != nil {
		return err
	}
	objType, k8sConfig := convertDockerContainerToRC(dockerContainer)
	cr.Body = string(k8sConfig)
	cr.Request = "/api/v1/namespaces/default/" + objType

	return nil
}
