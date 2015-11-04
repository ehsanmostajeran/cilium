package prehook

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	m "github.com/cilium-team/cilium/cilium/messages"
	up "github.com/cilium-team/cilium/cilium/utils/profile"
	upr "github.com/cilium-team/cilium/cilium/utils/profile/runnables"
	uprd "github.com/cilium-team/cilium/cilium/utils/profile/runnables/docker"
	uprk "github.com/cilium-team/cilium/cilium/utils/profile/runnables/kubernetes"
	upsd "github.com/cilium-team/cilium/cilium/utils/profile/subpolicies/docker"
)

var (
	validType                    = `"pre-hook"`
	validPPV                     = 1
	validDockerRequestHeader     = `"/v1.15/containers/create?name=hello-world"`
	validKubernetesRequestHeader = `"/v1/pods"`
	validBody                    = `{\"Hostname\":\"myhostname\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,` +
		`\"AttachStderr\":false,\"ExposedPorts\":{\"53/udp\":{},\"80/tcp\":{}},\"Tty\":false,\"OpenStdin\":false,` +
		`\"StdinOnce\":false,\"Env\":null,\"Cmd\":null,\"Image\":\"fooandbar\",` +
		`\"Volumes\":null,\"VolumeDriver\":\"\",\"WorkingDir\":\"\",\"Entrypoint\":null,\"NetworkDisabled\":false,` +
		`\"MacAddress\":\"\",\"OnBuild\":null,\"Labels\":{\"com.docker.swarm.id\":\"123456\"},\"Memory\":0,\"MemorySwap\":0,` +
		`\"CpuShares\":0,\"Cpuset\":\"\",\"PortSpecs\":null,\"HostConfig\":{\"Binds\":null,\"ContainerIDFile\":\"\",` +
		`\"LxcConf\":null,\"Memory\":0,\"MemorySwap\":0,\"CpuShares\":0,\"CpuPeriod\":0,\"CpusetCpus\":\"\",\"CpusetMems\":\"\",` +
		`\"CpuQuota\":0,\"BlkioWeight\":0,\"OomKillDisable\":false,\"Privileged\":false,` +
		`\"PortBindings\":{\"53/udp\":[{\"HostIp\":\"\",\"HostPort\":\"53\"}],\"80/tcp\":[{\"HostIp\":\"\",\"HostPort\":\"80\"}]},` +
		`\"Links\":null,\"PublishAllPorts\":false,\"Dns\":[\"8.8.8.8\",\"8.8.4.4\"],\"DnsSearch\":null,\"ExtraHosts\":null,` +
		`\"VolumesFrom\":null,\"Devices\":null,\"NetworkMode\":\"bridge\",\"IpcMode\":\"\",\"PidMode\":\"\",\"UTSMode\":\"\",` +
		`\"CapAdd\":null,\"CapDrop\":null,\"RestartPolicy\":{\"Name\":\"no\",\"MaximumRetryCount\":0},\"SecurityOpt\":null,` +
		`\"ReadonlyRootfs\":false,\"Ulimits\":null,\"LogConfig\":{\"type\":\"\",\"config\":null},\"CgroupParent\":\"\"}}`
	validMethod  = `POST`
	validRequest = `{"Type": ` + validType + `, "PowerstripProtocolVersion": ` +
		strconv.Itoa(validPPV) + `, "ClientRequest": {"Body": "` + validBody + `", ` +
		`"Request": ` + validDockerRequestHeader + `, "Method": "` + validMethod + `"}}`
	validWantBodyWoutEscQuot = `{"Hostname":"myhostname","ExposedPorts":{"53/udp":{},"80/tcp":{}},` +
		`"Cmd":null,"Image":"fooandbar","Entrypoint":null,"Labels":{"com.docker.swarm.id":"123456"},` +
		`"HostConfig":{"PortBindings":{"53/udp":[{"HostPort":"53"}],"80/tcp":[{"HostPort":"80"}]},` +
		`"Dns":["1.2.3.4"],"NetworkMode":"bridge","RestartPolicy":{"Name":"no"},"LogConfig":{}}}`
	validBodyWoutEscQuot                 = strings.Replace(validBody, `\"`, `"`, -1)
	validDockerRequestHeaderWoutQuot     = strings.Replace(validDockerRequestHeader, `"`, ``, -1)
	validKubernetesRequestHeaderWoutQuot = strings.Replace(validKubernetesRequestHeader, `"`, ``, -1)

	invalidRequest             = `"Request": "/nonque/create?name=hello-world"`
	invalidDockerRequestHeader = `/v1.15/foo/bar?name=hello-world`
)

func TestDefaultValidRequest(t *testing.T) {
	defaultPPHR, err := defaultRequest([]byte(validRequest))
	if err != nil {
		t.Fatal("error while parsing a valid request on Default pre-hook", err)
	}
	pphr, ok := defaultPPHR.(*PowerstripPreHookResponse)
	if !ok {
		t.Fatalf("invalid returned value:\ngot  %s\nwant %s",
			reflect.TypeOf(defaultPPHR),
			"*PowerstripPreHookResponse",
		)
	}
	if pphr.PowerstripProtocolVersion != m.PowerstripProtocolVersion {
		t.Errorf("invalid PowerstripProtocolVersion:\ngot  %d\nwant %d",
			pphr.PowerstripProtocolVersion,
			m.PowerstripProtocolVersion)
	}
	if pphr.ModifiedClientRequest.Request != validDockerRequestHeaderWoutQuot {
		t.Errorf("invalid ModifiedClientRequest.Request:\ngot  %s\nwant %s",
			pphr.ModifiedClientRequest.Request,
			validDockerRequestHeaderWoutQuot)
	}
	if pphr.ModifiedClientRequest.Method != validMethod {
		t.Errorf("invalid ModifiedClientRequest.Method:\ngot  %s\nwant %s",
			pphr.ModifiedClientRequest.Method,
			validMethod)
	}
	if pphr.ModifiedClientRequest.Body != validBodyWoutEscQuot {
		t.Errorf("invalid ModifiedClientRequest.Body:\ngot  %s\nwant %s",
			pphr.ModifiedClientRequest.Body,
			validBodyWoutEscQuot)
	}
}

func TestDefaultInvalidRequest(t *testing.T) {
	_, err := defaultRequest([]byte(invalidRequest))
	if err == nil {
		t.Error("should have return an error while parsing an invalid request" +
			"on Default pre-hook")
	}
}

func TestPreHook(t *testing.T) {
	f := FakeDB{}
	f.OnGetUsers = func() ([]up.User, error) {
		return []up.User{
			up.User{ID: 0, Name: "root"},
		}, nil
	}

	f.OnGetPoliciesThatCovers = func(labels map[string]string) ([]up.PolicySource, error) {
		own1 := "root"
		own1Pol := []up.Policy{
			up.Policy{
				Name:  "something",
				Owner: own1,
				Coverage: up.Coverage{
					Labels: map[string]string{"com.docker.swarm.id": "123456"},
				},
				DockerConfig: upsd.DockerConfig{
					HostConfig: upsd.HostConfig{
						DNS: []string{"1.2.3.4"},
					},
				},
			},
		}

		return []up.PolicySource{
			up.PolicySource{Owner: own1, Policies: own1Pol},
		}, nil
	}

	var ph PreHook
	ph.dbConn = f

	upr.Register(uprd.Name, uprd.DockerRunnable{})

	defaultPPHR, err := ph.preHook(uprd.DockerSwarmCreate, []byte(validRequest))
	if err != nil {
		t.Error("error occured while executing preHook", err)
	}
	pphr, ok := defaultPPHR.(*PowerstripPreHookResponse)
	if !ok {
		t.Fatalf("invalid returned value:\ngot  %s\nwant %s",
			reflect.TypeOf(defaultPPHR),
			"*PowerstripPreHookResponse",
		)
	}
	if pphr.PowerstripProtocolVersion != m.PowerstripProtocolVersion {
		t.Errorf("invalid PowerstripProtocolVersion:\ngot  %d\nwant %d",
			pphr.PowerstripProtocolVersion,
			m.PowerstripProtocolVersion)
	}
	if pphr.ModifiedClientRequest.Request != validDockerRequestHeaderWoutQuot {
		t.Errorf("invalid ModifiedClientRequest.Request:\ngot  %s\nwant %s",
			pphr.ModifiedClientRequest.Request,
			validDockerRequestHeaderWoutQuot)
	}
	if pphr.ModifiedClientRequest.Method != validMethod {
		t.Errorf("invalid ModifiedClientRequest.Method:\ngot  %s\nwant %s",
			pphr.ModifiedClientRequest.Method,
			validMethod)
	}
	if pphr.ModifiedClientRequest.Body != validWantBodyWoutEscQuot {
		t.Errorf("invalid ModifiedClientRequest.Body:\ngot  %s\nwant %s",
			pphr.ModifiedClientRequest.Body,
			validWantBodyWoutEscQuot)
	}
}

func TestPreHookInvalid(t *testing.T) {
	f := FakeDB{}
	f.OnGetUsers = func() ([]up.User, error) {
		return nil, fmt.Errorf("unable to connect DB")
	}

	var ph PreHook
	ph.dbConn = f

	upr.Register(uprd.Name, uprd.DockerRunnable{})

	defaultPPHR, err := ph.preHook("Default", []byte(validRequest))
	if err != nil {
		t.Error("error occured while executing preHook", err)
	}
	pphr, ok := defaultPPHR.(*PowerstripPreHookResponse)
	if !ok {
		t.Fatalf("invalid returned value:\ngot  %s\nwant %s",
			reflect.TypeOf(defaultPPHR),
			"*PowerstripPreHookResponse",
		)
	}
	if pphr.PowerstripProtocolVersion != m.PowerstripProtocolVersion {
		t.Errorf("invalid PowerstripProtocolVersion:\ngot  %d\nwant %d",
			pphr.PowerstripProtocolVersion,
			m.PowerstripProtocolVersion)
	}
	if pphr.ModifiedClientRequest.Request != validDockerRequestHeaderWoutQuot {
		t.Errorf("invalid ModifiedClientRequest.Request:\ngot  %s\nwant %s",
			pphr.ModifiedClientRequest.Request,
			validDockerRequestHeaderWoutQuot)
	}
	if pphr.ModifiedClientRequest.Method != validMethod {
		t.Errorf("invalid ModifiedClientRequest.Method:\ngot  %s\nwant %s",
			pphr.ModifiedClientRequest.Method,
			validMethod)
	}
	if pphr.ModifiedClientRequest.Body != validBodyWoutEscQuot {
		t.Errorf("invalid ModifiedClientRequest.Body:\ngot  %s\nwant %s",
			pphr.ModifiedClientRequest.Body,
			validBodyWoutEscQuot)
	}
}

func TestParseRequest(t *testing.T) {
	dockerTests := []struct {
		baseAddr string
		want     string
	}{
		{`/docker/swarm/cilium-adapter`, uprd.DockerSwarmCreate},
		{`/docker/daemon/cilium-adapter`, uprd.DockerDaemonCreate},
		{`/something`, "Default"},
	}

	upr.Register(uprd.Name, uprd.DockerRunnable{})
	upr.Register(uprk.Name, uprk.KubernetesRunnable{})
	p := PreHook{
		handlers: map[string]string{},
	}
	for _, runnable := range upr.GetRunnables() {
		runHandlers := runnable.GetHandlers(Type)
		for k, v := range runHandlers {
			p.handlers[k] = v
		}
	}

	for _, tt := range dockerTests {
		if got := p.parseRequest(tt.baseAddr, validDockerRequestHeaderWoutQuot); got != tt.want {
			t.Errorf("invalid parsed request:\ngot  %s\nwant %s", got, tt.want)
		}
	}
	for _, tt := range dockerTests {
		if got := p.parseRequest(tt.baseAddr, invalidDockerRequestHeader); got != "Default" {
			t.Errorf("invalid parsed request:\ngot  %s\nwant %s", got, "Default")
		}
	}

	kubernetesTests := []struct {
		baseAddr string
		want     string
	}{
		{`/kubernetes/master/cilium-adapter/api/v1/namespaces`, uprk.KubernetesMasterCreate},
	}
	for _, tt := range kubernetesTests {
		if got := p.parseRequest(tt.baseAddr, validKubernetesRequestHeaderWoutQuot); got != tt.want {
			t.Errorf("invalid parsed request:\ngot  %s\nwant %s", got, tt.want)
		}
	}
	for _, tt := range kubernetesTests {
		if got := p.parseRequest(tt.baseAddr, invalidDockerRequestHeader); got != "Default" {
			t.Errorf("invalid parsed request:\ngot  %s\nwant %s", got, "Default")
		}
	}
}
