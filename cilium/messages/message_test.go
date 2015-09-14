package messages

import (
	"strconv"
	"strings"
	"testing"
)

var (
	validType          = `"pre-hook"`
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
	validMethod  = `POST`
	validRequest = `{"Type": ` + validType + `, "PowerstripProtocolVersion": ` +
		strconv.Itoa(validPPV) + `, "ClientRequest": {"Body": "` + validBody + `", ` +
		`"Request": ` + validRequestHeader + `, "Method": "` + validMethod + `"}}`
	validBodyWoutEscQuot       = strings.Replace(validBody, `\"`, `"`, -1)
	validRequestHeaderWoutQuot = strings.Replace(validRequestHeader, `"`, ``, -1)
	validTypeWoutQuot          = strings.Replace(validType, `"`, ``, -1)

	invalidRequest = `"Request": "/nonque/create?name=hello-world"`
)

func TestDecodeRequestValidRequest(t *testing.T) {
	var powerStripReq PowerstripRequest
	err := DecodeRequest([]byte(validRequest), &powerStripReq)
	if err != nil {
		t.Fatal("invalid request message:", err)
	}
	if powerStripReq.PowerstripProtocolVersion != validPPV {
		t.Errorf("invalid PowerstripProtocolVersion: \ngot  %d \nwant %d",
			powerStripReq.PowerstripProtocolVersion,
			PowerstripProtocolVersion)
	}
	if powerStripReq.ClientRequest.Body != validBodyWoutEscQuot {
		t.Errorf("invalid ClientRequest.Body: \ngot  %s \nwant %s",
			powerStripReq.ClientRequest.Body,
			validBodyWoutEscQuot)
	}
	if powerStripReq.ClientRequest.Method != validMethod {
		t.Errorf("invalid ClientRequest.Method: \ngot  %s \nwant %s",
			powerStripReq.ClientRequest.Method,
			validMethod)
	}
	if powerStripReq.ClientRequest.Request != validRequestHeaderWoutQuot {
		t.Errorf("invalid ClientRequest.Request: \ngot  %s \nwant %s",
			powerStripReq.ClientRequest.Request,
			validRequestHeaderWoutQuot)
	}
	if powerStripReq.Type != validTypeWoutQuot {
		t.Errorf("invalid Type: \ngot  %s \nwant %s",
			powerStripReq.Type,
			validTypeWoutQuot)
	}
}

func TestDecodeRequestInvalidRequest(t *testing.T) {
	var powerStripReq PowerstripRequest
	err := DecodeRequest([]byte(invalidRequest), powerStripReq)
	if err == nil {
		t.Fatal("should return an error code of invalid message")
	}
}
