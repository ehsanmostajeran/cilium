package kube2compose

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	m "github.com/cilium-team/cilium/cilium/messages"
	uc "github.com/cilium-team/cilium/cilium/utils/comm"
	ucdb "github.com/cilium-team/cilium/cilium/utils/comm/db"
	up "github.com/cilium-team/cilium/cilium/utils/profile"
	upr "github.com/cilium-team/cilium/cilium/utils/profile/runnables"

	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/cilium-team/go-logging"
	d "github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/gorilla/mux"
	k8s "github.com/cilium-team/cilium/Godeps/_workspace/src/k8s.io/kubernetes/pkg/api"
	//  k8sfields "github.com/cilium-team/cilium/Godeps/_workspace/src/k8s.io/kubernetes/pkg/fields"
	k8slabels "github.com/cilium-team/cilium/Godeps/_workspace/src/k8s.io/kubernetes/pkg/labels"
)

const (
	Name = "kube2compose-runnable"

	Kube2ComposeGetImage        = "Kube2ComposeGetImage"
	Kube2ComposeGetImages       = "Kube2ComposeGetImages"
	Kube2ComposeCreateContainer = "Kube2ComposeCreateContainer"
	Kube2ComposeCreateImage     = "Kube2ComposeCreateImage"
	Kube2ComposeCreateRC        = "Kube2ComposeCreateRC"
	Kube2ComposeGetContainers   = "Kube2ComposeGetContainers"
)

var (
	log = logging.MustGetLogger("cilium")

	preHookHandlers = map[string]string{
		`/kube2compose/master/cilium-adapter/.*/images/.*/json(\?.*)?`: Kube2ComposeGetImage,
		`/kube2compose/master/cilium-adapter/.*/images/json(\?.*)?`:    Kube2ComposeGetImages,
		`/kube2compose/master/cilium-adapter/.*/images/create(\?.*)?`:  Kube2ComposeCreateImage,
	}
	postHookHandlers = map[string]string{
		`/kube2compose/master/cilium-adapter/.*/containers/create(\?.*)?`:                        Kube2ComposeCreateContainer,
		`/kube2compose/master/cilium-adapter/api/v1/namespaces/.*/pods(\?.*)?`:                   Kube2ComposeGetContainers,
		`/kube2compose/master/cilium-adapter/api/v1/namespaces/.*/replicationcontrollers(\?.*)?`: Kube2ComposeCreateRC,
	}

	kube2composeHookHanlders = map[string]func(cr *m.ClientRequest, sr *m.ServerResponse, kor *m.KubernetesObjRef) error{
		upr.PreHook + Kube2ComposeGetImage:         preHookGetImage,
		upr.PreHook + Kube2ComposeGetImages:        preHookGetImages,
		upr.PreHook + Kube2ComposeCreateImage:      preHookCreateImage,
		upr.PostHook + Kube2ComposeCreateContainer: postHookCreateContainer,
		upr.PostHook + Kube2ComposeCreateRC:        postHookCreateRC,
		upr.PostHook + Kube2ComposeGetContainers:   postHookGetContainers,
	}
	dc, _ = uc.NewDockerClient()
	kc, _ = uc.NewKubernetesClient()
)

type Kube2ComposeRunnable struct {
}

func (k2cr Kube2ComposeRunnable) GetHandlers(typ string) map[string]string {
	switch typ {
	case upr.PreHook:
		return preHookHandlers
	case upr.PostHook:
		return postHookHandlers
	default:
		return nil
	}
}

func (k2cr Kube2ComposeRunnable) GetRunnableFrom(users []up.User, policies []up.PolicySource) upr.PolicyRunnable {
	return Kube2ComposeRunnable{}
}

func (k2cr Kube2ComposeRunnable) DockerExec(hookType, reqType string, db ucdb.Db, cc *m.DockerCreateConfig) error {
	return nil
}

func (k2cr Kube2ComposeRunnable) KubernetesExec(hookType, reqType string, db ucdb.Db, cc *m.KubernetesObjRef) error {
	return nil
}

func (k2cr Kube2ComposeRunnable) Compose2KubeExec(hookType, reqType string, cr *m.ClientRequest, sr *m.ServerResponse, kor *m.KubernetesObjRef) error {
	return nil
}

func (k2cr Kube2ComposeRunnable) Kube2ComposeExec(hookType, reqType string, cr *m.ClientRequest, sr *m.ServerResponse, kor *m.KubernetesObjRef) error {
	if f, ok := kube2composeHookHanlders[hookType+reqType]; ok {
		return f(cr, sr, kor)
	}
	return nil
}

func convertPodToDockerAPIContainer(pod k8s.Pod) []d.APIContainers {
	dockerContainers := []d.APIContainers{}
	for iContainer, container := range pod.Spec.Containers {
		log.Debug("container: %+v", container)
		dockerPorts := []d.APIPort{}
		for _, port := range container.Ports {
			dockerPort := d.APIPort{
				IP:          port.HostIP,
				PrivatePort: int64(port.ContainerPort),
				PublicPort:  int64(port.HostPort),
				Type:        string(port.Protocol),
			}
			dockerPorts = append(dockerPorts, dockerPort)
		}
		dContainer := d.APIContainers{
			ID:      strings.Replace(pod.Status.ContainerStatuses[iContainer].ContainerID, "docker://", "", -1),
			Names:   []string{"/" + container.Name},
			Image:   container.Image,
			Command: strings.Join(container.Command, " "),
			Created: pod.ObjectMeta.CreationTimestamp.Unix(),
			Status:  string(pod.Status.Phase),
			Ports:   dockerPorts,
			Labels:  pod.ObjectMeta.Labels,
		}
		dockerContainers = append(dockerContainers, dContainer)
	}
	return dockerContainers
}

func postHookGetContainers(cr *m.ClientRequest, sr *m.ServerResponse, kor *m.KubernetesObjRef) error {
	log.Debug("")
	var podList k8s.PodList
	if err := sr.ConvertTo(&podList); err != nil {
		return err
	}
	log.Debug("podList: %+v", podList)

	dockerContainers := []d.APIContainers{}
	for _, pod := range podList.Items {
		log.Debug("pod: %+v", pod)
		convContainers := convertPodToDockerAPIContainer(pod)
		dockerContainers = append(dockerContainers, convContainers...)
	}

	log.Debug("dockerContainers: %+v", dockerContainers)
	bytes, err := json.Marshal(dockerContainers)
	if err != nil {
		return err
	}
	sr.Body = string(bytes)

	log.Debug("sr.Body: %+v", sr.Body)
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

func changeToDockerEndpoint(cr *m.ClientRequest) error {
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

func preHookGetImage(cr *m.ClientRequest, sr *m.ServerResponse, kor *m.KubernetesObjRef) error {
	log.Debug("")
	return changeToDockerEndpoint(cr)
}

func preHookGetImages(cr *m.ClientRequest, sr *m.ServerResponse, kor *m.KubernetesObjRef) error {
	log.Debug("")
	return changeToDockerEndpoint(cr)
}

func preHookCreateImage(cr *m.ClientRequest, sr *m.ServerResponse, kor *m.KubernetesObjRef) error {
	log.Debug("")
	return changeToDockerEndpoint(cr)
}

func convertPodToDockerContainer(pod k8s.Pod) []d.Container {
	dockerContainers := []d.Container{}
	log.Debug("pod: %+v", pod)
	for iContainer, container := range pod.Spec.Containers {
		log.Debug("container: %+v", container)
		dockerPorts := map[d.Port]struct{}{}
		for _, port := range container.Ports {
			dockerPort := d.Port(string(port.ContainerPort) + "/" + string(port.Protocol))
			dockerPorts[dockerPort] = struct{}{}
		}
		envs := []string{}
		for _, envvar := range container.Env {
			env := envvar.Name + "=" + envvar.Value
			envs = append(envs, env)
		}
		config := d.Config{
			Hostname:     pod.ObjectMeta.Name,
			AttachStdin:  container.Stdin,
			ExposedPorts: dockerPorts,
			Tty:          container.TTY,
			Env:          envs,
			Image:        container.Image,
			Cmd:          container.Command,
			Labels:       pod.ObjectMeta.Labels,
		}
		state := d.State{
			Running:   pod.Status.Phase == "Running",
			StartedAt: pod.ObjectMeta.CreationTimestamp.Time,
		}
		log.Debug("cs %+v::: %+v::: %+v", pod.Status.ContainerStatuses, len(pod.Status.ContainerStatuses), iContainer)
		dContainer := d.Container{
			ID:      strings.Replace(pod.Status.ContainerStatuses[iContainer].ContainerID, "docker://", "", -1),
			Created: pod.ObjectMeta.CreationTimestamp.Time,
			State:   state,
			Image:   strings.Replace(pod.Status.ContainerStatuses[iContainer].ImageID, "docker://", "", -1),
			Config:  &config,
			Args:    container.Command,
		}
		dockerContainers = append(dockerContainers, dContainer)
	}
	return dockerContainers
}

type CreateResp struct {
	ID       string  `json:"Id"`
	Warnings *string `json:"Warnings,omitempty"`
}

func dContainer2dCreateResp(c d.Container) CreateResp {
	return CreateResp{
		ID:       c.ID,
		Warnings: nil,
	}
}

func postHookCreateContainer(cr *m.ClientRequest, sr *m.ServerResponse, kor *m.KubernetesObjRef) error {
	log.Debug("")
	if sr.Code != http.StatusCreated {
		return nil
	}
	var pod k8s.Pod
	if err := sr.ConvertTo(&pod); err != nil {
		return err
	}
	lblSelect := k8slabels.SelectorFromSet(k8slabels.Set(pod.Labels))
	// The pods might not being created so we try again.
	var dockerContainers []d.Container
	log.Info("Trying to get pods for labels %+v", lblSelect)
	wait := 1 * time.Second
	retries := 10
	for {
		log.Info("Attempt %d...", 11-retries)
		podList, err := kc.Pods(pod.Namespace).List(lblSelect, nil)
		if err != nil {
			log.Error("Error while getting pods: %+v", err)
			return err
		}
		if len(podList.Items) != 0 &&
			len(podList.Items[0].Status.ContainerStatuses) != 0 {
			for _, pod := range podList.Items {
				k8sContainers := convertPodToDockerContainer(pod)
				dockerContainers = append(dockerContainers, k8sContainers...)
			}
			break
		}
		if retries < 0 {
			log.Error("0 pods gotten")
			return nil
		}
		time.Sleep(wait)
		wait += wait
		retries--
	}
	if len(dockerContainers) != 1 {
		log.Warning("Expecting the number of containers would be 1, it was %d.", len(dockerContainers))
	} else if len(dockerContainers) != 0 {
		dockerCreateResp := dContainer2dCreateResp(dockerContainers[0])
		byt, err := json.Marshal(dockerCreateResp)
		if err != nil {
			log.Error("Error while marshalling %+v", dockerCreateResp)
			return err
		}
		sr.Body = string(byt)
	}

	return nil
}

func postHookCreateRC(cr *m.ClientRequest, sr *m.ServerResponse, kor *m.KubernetesObjRef) error {
	log.Debug("")
	// It doesn't hurt doing this because CreateContainer will only need labels
	// from a RC or Pod
	return postHookCreateContainer(cr, sr, kor)
}
