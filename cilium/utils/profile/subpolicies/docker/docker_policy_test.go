package docker_policy

import (
	"encoding/json"
	"reflect"
	"testing"
)

var (
	wanthcjson = `{"Dns":["1.1.0.252"],"NetworkMode":"none","RestartPolicy":{},"LogConfig":{"Type":"json-file"}}`
	hcjson     = `{
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
        "Dns": [ "1.1.0.252" ],
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
    }`
	wantcjson = `{"Hostname":"web","ExposedPorts":{"5000/tcp":{}},"Env":` +
		`["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",` +
		`"LANG=C.UTF-8","PYTHON_VERSION=2.7.10","PYTHON_PIP_VERSION=7.1.0"],` +
		`"Cmd":["python","app.py"],"Image":"compose_web","WorkingDir":"/code",` +
		`"Entrypoint":null,"Labels":{"com.docker.compose.config-hash":` +
		`"86f0ff063d4b5f5015cd95ff6e0f7688a925d95c8244f966029b64887f84bb06",` +
		`"com.docker.compose.container-number":"1","com.docker.compose.oneoff":` +
		`"False","com.docker.compose.project":"compose","com.docker.compose.service":` +
		`"web","com.docker.compose.version":"1.3.0rc3","com.docker.swarm.id":` +
		`"8935dfa305a17d03b50de019c52c05f136612efc75dcaefdae9c3ba4658b7b01"}}`
	cjson = `{
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
    }
}`
)

func TestHostConfigValue(t *testing.T) {
	var hc HostConfig
	if err := json.Unmarshal([]byte(hcjson), &hc); err != nil {
		t.Fatalf("error while unmarshalling hcjson: %s", err)
	}
	gotHostConfig, err := hc.Value()
	if err != nil {
		t.Errorf("error while executing Value of HostConfig: %s", err)
	}
	if gotHostConfig != wanthcjson {
		t.Errorf("invalid host config gotten:\ngot  %s\nwant %s\n", gotHostConfig, wanthcjson)
	}
}

func TestHostConfigScan(t *testing.T) {
	var hc, hcwant HostConfig
	if err := json.Unmarshal([]byte(hcjson), &hcwant); err != nil {
		t.Fatalf("error while unmarshalling hcjson: %s", err)
	}
	if err := hc.Scan(hcjson); err != nil {
		t.Errorf("error while executing Value of HostConfig: %s", err)
	}
	if !reflect.DeepEqual(hc, hcwant) {
		t.Errorf("invalid host config gotten:\ngot  %s\nwant %s\n", hc, hcwant)
	}
}

func TestHostOverwriteWith(t *testing.T) {
	var hc1, hc2, hcwant HostConfig
	if err := json.Unmarshal([]byte(hcjson), &hcwant); err != nil {
		t.Fatalf("error while unmarshalling hcjson: %s", err)
	}
	if err := json.Unmarshal([]byte(hcjson), &hc1); err != nil {
		t.Fatalf("error while unmarshalling hcjson: %s", err)
	}
	if err := json.Unmarshal([]byte(hcjson), &hc2); err != nil {
		t.Fatalf("error while unmarshalling hcjson: %s", err)
	}
	hc1.Links = []string{"foo"}
	hc2.DNS = []string{"8.8.8.8"}
	hcwant.Links = []string{"foo"}
	hcwant.DNS = []string{"8.8.8.8"}
	if err := hc1.OverwriteWith(hc2); err != nil {
		t.Errorf("error while executing OverwriteWith of HostConfig: %s", err)
	}
	if !reflect.DeepEqual(hc1, hcwant) {
		t.Errorf("invalid host config gotten:\ngot  %s\nwant %s\n", hc1, hcwant)
	}
}

func TestConfigValue(t *testing.T) {
	var c Config
	if err := json.Unmarshal([]byte(cjson), &c); err != nil {
		t.Fatalf("error while unmarshalling cjson: %s", err)
	}
	gotConfig, err := c.Value()
	if err != nil {
		t.Errorf("error while executing Value of Config: %s", err)
	}
	if gotConfig != wantcjson {
		t.Errorf("invalid config gotten:\ngot  %s\nwant %s\n", gotConfig, wantcjson)
	}
}

func TestConfigScan(t *testing.T) {
	var c, cwant Config
	if err := json.Unmarshal([]byte(cjson), &cwant); err != nil {
		t.Fatalf("error while unmarshalling cjson: %s", err)
	}
	if err := c.Scan(cjson); err != nil {
		t.Errorf("error while executing Value of Config: %s", err)
	}
	if !reflect.DeepEqual(c, cwant) {
		t.Errorf("invalid config gotten:\ngot  %s\nwant %s\n", c, cwant)
	}
}

func TestOverwriteWith(t *testing.T) {
	var c1, c2, cwant Config
	if err := json.Unmarshal([]byte(cjson), &cwant); err != nil {
		t.Fatalf("error while unmarshalling cjson: %s", err)
	}
	if err := json.Unmarshal([]byte(cjson), &c1); err != nil {
		t.Fatalf("error while unmarshalling cjson: %s", err)
	}
	if err := json.Unmarshal([]byte(cjson), &c2); err != nil {
		t.Fatalf("error while unmarshalling cjson: %s", err)
	}
	c1.MacAddress = "00:01:02:03:04:05"
	c2.DNS = []string{"8.8.8.8"}
	c2.Labels = nil
	c2.MacAddress = ""
	cwant.MacAddress = "00:01:02:03:04:05"
	cwant.DNS = []string{"8.8.8.8"}
	if err := c1.OverwriteWith(c2); err != nil {
		t.Errorf("error while executing OverwriteWith of Config: %s", err)
	}
	if !reflect.DeepEqual(c1, cwant) {
		t.Errorf("invalid config gotten:\ngot  %s\nwant %s\n", c1, cwant)
	}
}

func TestDockerConfigMergeWithOverwrite(t *testing.T) {
	var hc HostConfig
	if err := json.Unmarshal([]byte(hcjson), &hc); err != nil {
		t.Fatalf("error while unmarshalling hcjson: %s", err)
	}
	var c Config
	if err := json.Unmarshal([]byte(cjson), &c); err != nil {
		t.Fatalf("error while unmarshalling cjson: %s", err)
	}
	dc1 := DockerConfig{
		Priority:   222,
		HostConfig: hc,
		Config:     c,
	}
	dc2 := DockerConfig{
		Priority:   2,
		HostConfig: hc,
		Config:     c,
	}
	dc2.HostConfig.DNS = []string{"1.1.1.1"}
	dcwant := DockerConfig{
		Priority:   2,
		HostConfig: hc,
		Config:     c,
	}
	dcwant.HostConfig.DNS = []string{"1.1.1.1"}
	if err := dc1.MergeWithOverwrite(dc2); err != nil {
		t.Errorf("error while MergeWithOverwrite dc1 with dc2: %s", err)
	}
	if !reflect.DeepEqual(dc1, dcwant) {
		t.Errorf("invalid dc1:\ngot  %+v\nwant %+v", dc1, dcwant)
	}
}

func TestDockerConfigOverwriteWith(t *testing.T) {
	var hc HostConfig
	if err := json.Unmarshal([]byte(hcjson), &hc); err != nil {
		t.Fatalf("error while unmarshalling hcjson: %s", err)
	}
	var c Config
	if err := json.Unmarshal([]byte(cjson), &c); err != nil {
		t.Fatalf("error while unmarshalling cjson: %s", err)
	}
	dc1 := DockerConfig{
		Priority:   222,
		HostConfig: hc,
		Config:     c,
	}
	dc2 := DockerConfig{
		Priority:   2,
		HostConfig: hc,
		Config:     c,
	}
	dc1.Config.MacAddress = "00:01:02:03:04:05"
	dc2.HostConfig.DNS = []string{"8.8.8.8"}
	dc2.Config.Labels = nil
	dc2.Config.MacAddress = ""
	dcwant := DockerConfig{
		Priority:   2,
		HostConfig: hc,
		Config:     c,
	}
	dcwant.Config.MacAddress = "00:01:02:03:04:05"
	dcwant.HostConfig.DNS = []string{"8.8.8.8"}
	if err := dc1.OverwriteWith(dc2); err != nil {
		t.Errorf("error while OverwriteWith dc1 with dc2: %s", err)
	}
	if !reflect.DeepEqual(dc1, dcwant) {
		t.Errorf("invalid dc1:\ngot  %+v\nwant %+v", dc1, dcwant)
	}
}

func TestOrderDockerConfigsByAscendingPriority(t *testing.T) {
	var hc HostConfig
	if err := json.Unmarshal([]byte(hcjson), &hc); err != nil {
		t.Fatalf("error while unmarshalling hcjson: %s", err)
	}
	var c Config
	if err := json.Unmarshal([]byte(cjson), &c); err != nil {
		t.Fatalf("error while unmarshalling cjson: %s", err)
	}
	dcs := []DockerConfig{
		DockerConfig{
			Priority:   222,
			HostConfig: hc,
			Config:     c,
		},
		DockerConfig{
			Priority:   2,
			HostConfig: hc,
			Config:     c,
		},
		DockerConfig{
			Priority:   1,
			HostConfig: hc,
			Config:     c,
		},
		DockerConfig{
			Priority:   199,
			HostConfig: hc,
			Config:     c,
		},
	}
	want := []DockerConfig{
		DockerConfig{
			Priority:   1,
			HostConfig: hc,
			Config:     c,
		},
		DockerConfig{
			Priority:   2,
			HostConfig: hc,
			Config:     c,
		},
		DockerConfig{
			Priority:   199,
			HostConfig: hc,
			Config:     c,
		},
		DockerConfig{
			Priority:   222,
			HostConfig: hc,
			Config:     c,
		},
	}
	OrderDockerConfigsByAscendingPriority(dcs)
	for i := 0; i < len(dcs); i++ {
		if dcs[i].Priority != want[i].Priority {
			t.Errorf("DockerConfigs are blady sorted (Priority):\ngot  %d\nwant %d", dcs[i].Priority, want[i].Priority)
		}
	}
}
