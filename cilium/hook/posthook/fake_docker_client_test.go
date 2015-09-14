package posthook

import (
	"io/ioutil"
	"net/http"
	"strings"

	uc "github.com/cilium-team/cilium/cilium/utils/comm"

	d "github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

var jsonContainer = `{
    "Id": "6b27a943823d0f735346861bbce6e24acdaf435edb259748be556300d1c361f3",
    "Created": "2015-07-10T00:56:23.570550811Z",
    "Path": "python",
    "Args": [
        "app.py"
    ],
    "State": {
        "Running": false,
        "Paused": false,
        "Restarting": false,
        "OOMKilled": false,
        "Dead": false,
        "Pid": 0,
        "ExitCode": 0,
        "Error": "",
        "StartedAt": "2015-07-10T00:56:23.729628128Z",
        "FinishedAt": "2015-07-10T02:32:16.128286075Z"
    },
    "Image": "63162beb990104dfc2088d8537d7a57399f55d29e63be68b1fdf8033bc181a55",
    "NetworkSettings": {
        "Bridge": "",
        "EndpointID": "",
        "Gateway": "",
        "GlobalIPv6Address": "",
        "GlobalIPv6PrefixLen": 0,
        "HairpinMode": false,
        "IPAddress": "",
        "IPPrefixLen": 0,
        "IPv6Gateway": "",
        "LinkLocalIPv6Address": "",
        "LinkLocalIPv6PrefixLen": 0,
        "MacAddress": "",
        "NetworkID": "",
        "PortMapping": null,
        "Ports": null,
        "SandboxKey": "",
        "SecondaryIPAddresses": null,
        "SecondaryIPv6Addresses": null
    },
    "ResolvConfPath": "/var/lib/docker/containers/6b27a943823d0f735346861bbce6e24acdaf435edb259748be556300d1c361f3/resolv.conf",
    "HostnamePath": "/var/lib/docker/containers/6b27a943823d0f735346861bbce6e24acdaf435edb259748be556300d1c361f3/hostname",
    "HostsPath": "/var/lib/docker/containers/6b27a943823d0f735346861bbce6e24acdaf435edb259748be556300d1c361f3/hosts",
    "LogPath": "/var/lib/docker/containers/6b27a943823d0f735346861bbce6e24acdaf435edb259748be556300d1c361f3/6b27a943823d0f735346861bbce6e24acdaf435edb259748be556300d1c361f3-json.log",
    "Name": "/compose_web_1",
    "RestartCount": 0,
    "Driver": "overlay",
    "ExecDriver": "native-0.2",
    "MountLabel": "system_u:object_r:svirt_sandbox_file_t:s0:c419,c621",
    "ProcessLabel": "system_u:system_r:svirt_lxc_net_t:s0:c419,c621",
    "Volumes": {},
    "VolumesRW": {},
    "AppArmorProfile": "",
    "ExecIDs": null,
    "HostConfig": {
        "Binds": null,
        "ContainerIDFile": "",
        "LxcConf": null,
        "Memory": 0,
        "MemorySwap": 0,
        "CpuShares": 0,
        "CpuPeriod": 0,
        "CpusetCpus": "",
        "CpusetMems": "",
        "CpuQuota": 0,
        "BlkioWeight": 0,
        "OomKillDisable": false,
        "Privileged": false,
        "PortBindings": null,
        "Links": null,
        "PublishAllPorts": false,
        "Dns": [
            "1.1.0.252"
        ],
        "DnsSearch": null,
        "ExtraHosts": null,
        "VolumesFrom": null,
        "Devices": null,
        "NetworkMode": "none",
        "IpcMode": "",
        "PidMode": "",
        "UTSMode": "",
        "CapAdd": null,
        "CapDrop": null,
        "RestartPolicy": {
            "Name": "",
            "MaximumRetryCount": 0
        },
        "SecurityOpt": null,
        "ReadonlyRootfs": false,
        "Ulimits": null,
        "LogConfig": {
            "Type": "json-file",
            "Config": null
        },
        "CgroupParent": ""
    },
    "Config": {
        "Hostname": "web",
        "Domainname": "",
        "User": "",
        "AttachStdin": false,
        "AttachStdout": false,
        "AttachStderr": false,
        "PortSpecs": null,
        "ExposedPorts": {
            "5000/tcp": {}
        },
        "Tty": false,
        "OpenStdin": false,
        "StdinOnce": false,
        "Env": [
            "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
            "LANG=C.UTF-8",
            "PYTHON_VERSION=2.7.10",
            "PYTHON_PIP_VERSION=7.1.0"
        ],
        "Cmd": [
            "python",
            "app.py"
        ],
        "Image": "compose_web",
        "Volumes": null,
        "VolumeDriver": "",
        "WorkingDir": "/code",
        "Entrypoint": null,
        "NetworkDisabled": false,
        "MacAddress": "",
        "OnBuild": null,
        "Labels": {
            "com.docker.compose.config-hash": "86f0ff063d4b5f5015cd95ff6e0f7688a925d95c8244f966029b64887f84bb06",
            "com.docker.compose.container-number": "1",
            "com.docker.compose.oneoff": "False",
            "com.docker.compose.project": "compose",
            "com.docker.compose.service": "web",
            "com.docker.compose.version": "1.3.0rc3",
            "com.docker.swarm.id": "8935dfa305a17d03b50de019c52c05f136612efc75dcaefdae9c3ba4658b7b01"
        },
        "Init": ""
    }
}`

type FakeDocker uc.Docker

func newMockDockerClient(fdc *FakeDockerClient) *d.Client {
	dc, _ := d.NewClient("http://localhost:8080")
	dc.HTTPClient = &http.Client{Transport: fdc}
	dc.SkipServerVersionCheck = true
	return dc
}

type FakeDockerClient struct {
	message  string
	status   int
	header   map[string]string
	requests []*http.Request
}

func (fdc *FakeDockerClient) RoundTrip(r *http.Request) (*http.Response, error) {
	body := strings.NewReader(fdc.message)
	fdc.requests = append(fdc.requests, r)
	res := &http.Response{
		StatusCode: fdc.status,
		Body:       ioutil.NopCloser(body),
		Header:     make(http.Header),
	}
	for k, v := range fdc.header {
		res.Header.Set(k, v)
	}
	return res, nil
}

func (fdc *FakeDockerClient) Reset() {
	fdc.requests = nil
}
