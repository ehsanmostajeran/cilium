package posthook

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"testing"

	m "github.com/cilium-team/cilium/cilium/messages"
	uc "github.com/cilium-team/cilium/cilium/utils/comm"
	up "github.com/cilium-team/cilium/cilium/utils/profile"
	upr "github.com/cilium-team/cilium/cilium/utils/profile/runnables"
	uprd "github.com/cilium-team/cilium/cilium/utils/profile/runnables/docker"

	d "github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

var (
	validType          = `"post-hook"`
	validPPV           = 1
	validRequestHeader = `"/v1.15/containers/create?name=hello-world"`
	validBody          = `{\"Hostname\":\"myhostname\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,` +
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
	validMethod        = `POST`
	validHeaderReq     = `/docker/daemon/cilium-adapter/v1.15/containers/` + validContainerID + `/start?name=hello`
	validContentType   = `"application/json"`
	validServerRequest = `{"Type": ` + validType + `, "PowerstripProtocolVersion": ` +
		strconv.Itoa(validPPV) + `, "ClientRequest": {"Body": "` + validBody + `", ` +
		`"Request": "` + validHeaderReq + `", "Method": "` + validMethod + `"}, ` +
		`"ServerResponse": {"Body": "{\"Id\":\"` + validContainerID + `\",\"Warnings\":null}", ` +
		`"Code": ` + strconv.Itoa(validCode) + `, "ContentType": ` + validContentType + `}}`
	validCode                   = 201
	validContainerID            = `6b27a943823d0f735346861bbce6e24acdaf435edb259748be556300d1c361f3`
	validServerResponseBody     = `{"Id":"` + validContainerID + `","Warnings":null}`
	validRequestHeaderWoutQuot  = strings.Replace(validRequestHeader, `"`, ``, -1)
	validContentTypeWoutEscQuot = strings.Replace(validContentType, `"`, ``, -1)

	invalidRequest       = `"Request": "/nonque/create?name=hello-world"`
	invalidRequestHeader = `/v1.15/foo/bar?name=hello-world`
)

func TestDefaultValidRequest(t *testing.T) {
	defaultPPHR, err := defaultRequest([]byte(validServerRequest))
	if err != nil {
		t.Fatal("error while parsing a valid request on Default post-hook", err)
	}
	pphr, ok := defaultPPHR.(*PowerstripPostHookResponse)
	if !ok {
		t.Fatalf("invalid returned value:\ngot  %s\nwant %s",
			reflect.TypeOf(defaultPPHR),
			"*PowerstripPostHookResponse",
		)
	}
	if pphr.PowerstripProtocolVersion != m.PowerstripProtocolVersion {
		t.Errorf("invalid PowerstripProtocolVersion:\ngot  %d\nwant %d",
			pphr.PowerstripProtocolVersion,
			m.PowerstripProtocolVersion)
	}
	if pphr.ModifiedServerResponse.ContentType != validContentTypeWoutEscQuot {
		t.Errorf("invalid ModifiedServerResponse.ContentType:\ngot  %s\nwant %s",
			pphr.ModifiedServerResponse.ContentType,
			validContentTypeWoutEscQuot)
	}
	if pphr.ModifiedServerResponse.Code != validCode {
		t.Errorf("invalid ModifiedServerResponse.Code:\ngot  %d\nwant %d",
			pphr.ModifiedServerResponse.Code,
			validCode)
	}
	if pphr.ModifiedServerResponse.Body != validServerResponseBody {
		t.Errorf("invalid ModifiedServerResponse.Body:\ngot  %s\nwant %s",
			pphr.ModifiedServerResponse.Body,
			validServerResponseBody)
	}
}

func TestDefaultInvalidRequest(t *testing.T) {
	_, err := defaultRequest([]byte(invalidRequest))
	if err == nil {
		t.Error("should have return an error while parsing an invalid request" +
			"on Default pre-hook")
	}
}

func TestGetDockerIDFrom(t *testing.T) {
	validReq1 := `/docker/daemon/cilium-adapter/v1.15/containers/` + validContainerID + `/start?name=hello`
	validReq2 := `/docker/daemon/cilium-adapter/v1.15/containers/` + validContainerID + `/start?name=` +
		`hello-025ec22c60f02cdaf765829ac74c0ecb83f3553de9a33f994b448482ad5b2002`
	invalidReq1 := `/docker/daemon/cilium-adapter/v1.15/containers/create?name=` +
		`hello-` + validContainerID

	if got := getDockerIDFrom(validReq1); got != validContainerID {
		t.Errorf("invalid Docker ID:\ngot  %s\nwant %s", got, validContainerID)
	}
	if got := getDockerIDFrom(validReq2); got != validContainerID {
		t.Errorf("invalid Docker ID:\ngot  %s\nwant %s", got, validContainerID)
	}
	if got := getDockerIDFrom(invalidReq1); got != "" {
		t.Errorf("invalid Docker ID:\ngot  %s\nwant %s", got, "")
	}
}

func TestPostHook(t *testing.T) {
	fdb := FakeDB{}
	fdb.OnGetUsers = func() ([]up.User, error) {
		return []up.User{
			up.User{ID: 0, Name: "root"},
		}, nil
	}

	var expected d.Container
	err := json.Unmarshal([]byte(jsonContainer), &expected)
	if err != nil {
		t.Fatal(err)
	}
	fakeDC := &FakeDockerClient{message: jsonContainer, status: http.StatusOK}
	dc := uc.Docker{Client: newMockDockerClient(fakeDC)}

	var ph PostHook
	ph.dbConn = fdb
	ph.dockerConn = dc

	upr.Register("docker-config", uprd.DockerRunnable{})

	defaultPPHR, err := ph.postHook("Default", []byte(validServerRequest))
	if err != nil {
		t.Error("error occured while executing postHook", err)
	}
	pphr, ok := defaultPPHR.(*PowerstripPostHookResponse)
	if !ok {
		t.Fatalf("invalid returned value:\ngot  %s\nwant %s",
			reflect.TypeOf(defaultPPHR),
			"*PowerstripPostHookResponse",
		)
	}
	if pphr.PowerstripProtocolVersion != m.PowerstripProtocolVersion {
		t.Errorf("invalid PowerstripProtocolVersion:\ngot  %d\nwant %d",
			pphr.PowerstripProtocolVersion,
			m.PowerstripProtocolVersion)
	}
	if pphr.ModifiedServerResponse.ContentType != validContentTypeWoutEscQuot {
		t.Errorf("invalid ModifiedServerResponse.ContentType:\ngot  %s\nwant %s",
			pphr.ModifiedServerResponse.ContentType,
			validContentTypeWoutEscQuot)
	}
	if pphr.ModifiedServerResponse.Code != validCode {
		t.Errorf("invalid ModifiedServerResponse.Code:\ngot  %d\nwant %d",
			pphr.ModifiedServerResponse.Code,
			validCode)
	}
	if pphr.ModifiedServerResponse.Body != validServerResponseBody {
		t.Errorf("invalid ModifiedServerResponse.Body:\ngot  %s\nwant %s",
			pphr.ModifiedServerResponse.Body,
			validServerResponseBody)
	}
}

func TestPostHookInvalid(t *testing.T) {
	fdb := FakeDB{}
	fdb.OnGetUsers = func() ([]up.User, error) {
		return nil, fmt.Errorf("unable to connect DB")
	}

	var expected d.Container
	err := json.Unmarshal([]byte(jsonContainer), &expected)
	if err != nil {
		t.Fatal(err)
	}
	fakeDC := &FakeDockerClient{message: jsonContainer, status: http.StatusOK}
	dc := uc.Docker{Client: newMockDockerClient(fakeDC)}

	var ph PostHook
	ph.dbConn = fdb
	ph.dockerConn = dc

	upr.Register("docker-config", uprd.DockerRunnable{})

	defaultPPHR, err := ph.postHook("Default", []byte(validServerRequest))
	if err != nil {
		t.Error("error occured while executing postHook", err)
	}
	pphr, ok := defaultPPHR.(*PowerstripPostHookResponse)
	if !ok {
		t.Fatalf("invalid returned value:\ngot  %s\nwant %s",
			reflect.TypeOf(defaultPPHR),
			"*PowerstripPostHookResponse",
		)
	}
	if pphr.PowerstripProtocolVersion != m.PowerstripProtocolVersion {
		t.Errorf("invalid PowerstripProtocolVersion:\ngot  %d\nwant %d",
			pphr.PowerstripProtocolVersion,
			m.PowerstripProtocolVersion)
	}
	if pphr.ModifiedServerResponse.ContentType != validContentTypeWoutEscQuot {
		t.Errorf("invalid ModifiedServerResponse.ContentType:\ngot  %s\nwant %s",
			pphr.ModifiedServerResponse.ContentType,
			validContentTypeWoutEscQuot)
	}
	if pphr.ModifiedServerResponse.Code != validCode {
		t.Errorf("invalid ModifiedServerResponse.Code:\ngot  %d\nwant %d",
			pphr.ModifiedServerResponse.Code,
			validCode)
	}
	if pphr.ModifiedServerResponse.Body != validServerResponseBody {
		t.Errorf("invalid ModifiedServerResponse.Body:\ngot  %s\nwant %s",
			pphr.ModifiedServerResponse.Body,
			validServerResponseBody)
	}
}

func TestParseRequest(t *testing.T) {
	testsBase := []struct {
		baseAddr string
		want     string
	}{
		{`/docker/daemon/cilium-adapter`, upr.DockerDaemonCreate},
		{`/something`, "Default"},
	}
	testsHandlers := []struct {
		baseAddr string
		want     string
	}{
		{`/docker/daemon/cilium-adapter/v1.20/containers/48380b123e1be550f171787473a1f6683b1e3d966b2521b46d01eccfdf0e8b1f/restart?t=10`, "DockerDaemonRestart"},
		{`/docker/daemon/cilium-adapter/v1.20/containers/48380b123e1be550f171787473a1f6683b1e3d966b2521b46d01eccfdf0e8b1f/start?t=10`, "DockerDaemonStart"},
	}
	for _, tt := range testsHandlers {
		if got := parseRequest(tt.baseAddr, ""); got != tt.want {
			t.Errorf("invalid parsed request:\ngot  %s\nwant %s", got, tt.want)
		}
	}
	for _, tt := range testsBase {
		if got := parseRequest(tt.baseAddr, validRequestHeaderWoutQuot); got != tt.want {
			t.Errorf("invalid parsed request:\ngot  %s\nwant %s", got, tt.want)
		}
	}
	for _, tt := range testsBase {
		if got := parseRequest(tt.baseAddr, invalidRequestHeader); got != "Default" {
			t.Errorf("invalid parsed request:\ngot  %s\nwant %s", got, "Default")
		}
	}
}
