package messages

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	d "github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

var (
	validContainer = d.Container{
		Name:       "hello world",
		Config:     &validConfig,
		HostConfig: &validHostConfig,
		ID:         validContainerID,
		State: d.State{
			Pid: 1245,
		},
	}
	validConfig = d.Config{
		DNS: []string{"1.2.3.4"},
	}
	validHostConfig = d.HostConfig{
		Memory: 123456,
	}
	validContainerID = `6b27a943823d0f735346861bbce6e24acdaf435edb259748be556300d1c361f3`
	validSimpleBody  = `{\"Hostname\":\"myhostname\",\"ExposedPorts\":{\"53/udp\":{},` +
		`\"80/tcp\":{}},\"Cmd\":null,\"Image\":\"fooandbar\",\"Entrypoint\":null,` +
		`\"Labels\":{\"codocker.swarid\":\"123456\"},` +
		`\"HostConfig\":{\"PortBindings\":{\"53/udp\":[{\"HostPort\":\"53\"}],` +
		`\"80/tcp\":[{\"HostPort\":\"80\"}]},\"Dns\":[\"8.8.8.8\",\"8.8.4.4\"],` +
		`\"NetworkMode\":\"bridge\",\"RestartPolicy\":{\"Name\":\"no\"},\"LogConfig\":{}}}`
	validSimpleRequest = `{"Type": ` + validType + `, "PowerstripProtocolVersion": ` +
		strconv.Itoa(validPPV) + `, "ClientRequest": {"Body": "` + validSimpleBody + `", ` +
		`"Request": ` + validRequestHeader + `, "Method": "` + validMethod + `"}}`
	validSimpleConfigWoutEscQuot = `{"Hostname":"myhostname","ExposedPorts":{"53/udp":{},` +
		`"80/tcp":{}},"Cmd":null,"Image":"fooandbar","Entrypoint":null,` +
		`"Labels":{"codocker.swarid":"123456"}}`
	validSimpleBodyWoutEscQuot = strings.Replace(validSimpleBody, `\"`, `"`, -1)
)

func TestNewDockerCreateConfigFromDockerContainer(t *testing.T) {
	cc := NewDockerCreateConfigFromDockerContainer(validContainer)
	if cc.Name != validContainer.Name {
		t.Errorf("invalid CreateConfig:\ngot  %s\nwant %s", cc.Name, validContainer.Name)
	}
	if !reflect.DeepEqual(*cc.Config, validConfig) {
		t.Errorf("invalid CreateConfig:\ngot  %s\nwant %s", *cc.Config, validContainer.Config)
	}
	if !reflect.DeepEqual(*cc.HostConfig, validHostConfig) {
		t.Errorf("invalid CreateConfig:\ngot  %s\nwant %s", *cc.HostConfig, validContainer.HostConfig)
	}
	if !reflect.DeepEqual(cc.ID, validContainerID) {
		t.Errorf("invalid CreateConfig:\ngot  %s\nwant %s", cc.ID, validContainerID)
	}
	if cc.State.Pid != 1245 {
		t.Errorf("invalid CreateConfig:\ngot  %s\nwant %s", cc.State.Pid, 1245)
	}
}

func TestMergeWith(t *testing.T) {
	c1 := NewDockerCreateConfigFromDockerContainer(validContainer)
	c2 := NewDockerCreateConfigFromDockerContainer(validContainer)
	ccwant := NewDockerCreateConfigFromDockerContainer(validContainer)
	c2.Image = "foo"
	ccwant.Image = "foo"
	c1.MergeWith(c2)
	if !reflect.DeepEqual(c1, ccwant) {
		t.Errorf("invalid CreateConfig:\ngot  %s\nwant %s", c1, ccwant)
	}
	c1.Image = "bar"
	ccwant.Image = "bar"
	c1.MergeWith(c2)
	if !reflect.DeepEqual(c1, ccwant) {
		t.Errorf("invalid CreateConfig:\ngot  %s\nwant %s", c1, ccwant)
	}
}

func TestMergeWithOverwrite(t *testing.T) {
	c1 := NewDockerCreateConfigFromDockerContainer(validContainer)
	c2 := NewDockerCreateConfigFromDockerContainer(validContainer)
	ccwant := NewDockerCreateConfigFromDockerContainer(validContainer)
	c2.Image = "foo"
	ccwant.Image = "foo"
	c1.MergeWithOverwrite(c2)
	if !reflect.DeepEqual(c1, ccwant) {
		t.Errorf("invalid CreateConfig:\ngot  %s\nwant %s", c1, ccwant)
	}
	c1.Image = "bar"
	if reflect.DeepEqual(c1, ccwant) {
		t.Errorf("invalid CreateConfig:\ngot  %s\nwant %s", ccwant, c1)
	}
	c1.MergeWithOverwrite(c2)
	if !reflect.DeepEqual(c1, ccwant) {
		t.Errorf("invalid CreateConfig:\ngot  %s\nwant %s", c1, ccwant)
	}
}

func TestUnmarshalCreateClientBody(t *testing.T) {
	var powerStripReq PowerstripRequest
	err := DecodeRequest([]byte(validRequest), &powerStripReq)
	if err != nil {
		t.Fatal("invalid request message:", err)
	}
	var cc DockerCreateConfig
	err = powerStripReq.UnmarshalDockerCreateClientBody(&cc)
	if err != nil {
		t.Fatal("invalid request:", err)
	}
	if cc.Name != "/hello-world" {
		t.Errorf("invalid unmarshalling:\ngot  %s\nwant %s", cc.Name, "/hello-world")
	}
	if cc.Hostname != "myhostname" {
		t.Errorf("invalid unmarshalling:\ngot  %s\nwant %s", cc.Hostname, "myhostname")
	}
}

func TestUnmarshalClientBody(t *testing.T) {
	var powerStripReq PowerstripRequest
	err := DecodeRequest([]byte(validRequest), &powerStripReq)
	if err != nil {
		t.Fatal("invalid request message:", err)
	}
	var dc d.Config
	err = powerStripReq.UnmarshalClientBody(&dc)
	if err != nil {
		t.Fatal("invalid request:", err)
	}
	if dc.Hostname != "myhostname" {
		t.Errorf("invalid unmarshalling:\ngot  %s\nwant %s", dc.Hostname, "myhostname")
	}
}

func TestCreateConfigMarshal2JSONStr(t *testing.T) {
	var powerStripReq PowerstripRequest
	err := DecodeRequest([]byte(validSimpleRequest), &powerStripReq)
	if err != nil {
		t.Fatal("invalid request message:", err)
	}
	var cc DockerCreateConfig
	err = powerStripReq.UnmarshalDockerCreateClientBody(&cc)
	if err != nil {
		t.Fatal("invalid request:", err)
	}
	str, err := cc.Marshal2JSONStr()
	if err != nil {
		t.Fatal("invalid CreateConfig:", err)
	}
	if str != validSimpleBodyWoutEscQuot {
		t.Errorf("invalid marshaling:\ngot  %s\nwant %s", str, validSimpleBodyWoutEscQuot)
	}
}
