package loadbalancer

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/op/go-logging"
)

var log = logging.MustGetLogger("cilium")

const (
	haproxyClientIP         = "127.0.0.1"
	haproxyClientConfigPort = "10001"
	statsEndpoint           = "/v1/stats"
	configEndpoint          = "/v1/config"
	infoEndpoint            = "/v1/info"
	frontendEndpoint        = "/v1/frontend"
	backendEndpoint         = "/v1/backend"
)

type HAProxyClient struct {
	IP   string `json:"ip,omitempty" yaml:"ip,omitempty"`
	Port string `json:"port,omitempty" yaml:"port,omitempty"`
}

// Value marshals the receiver HAProxyClient into a json string.
func (h HAProxyClient) Value() (string, error) {
	if data, err := json.Marshal(h); err != nil {
		return "", err
	} else {
		return string(data), err
	}
}

// Scan unmarshals the input into the receiver HAProxyClient.
func (h *HAProxyClient) Scan(input string) error {
	return json.Unmarshal([]byte(input), h)
}

func NewHAProxyClientTo(ip, port string) (*HAProxyClient, error) {
	haproxyCli := new(HAProxyClient)
	haproxyCli.IP = ip
	haproxyCli.Port = port
	return haproxyCli, nil
}

func NewHAProxyClientToIP(ip string) (*HAProxyClient, error) {
	return NewHAProxyClientTo(ip, haproxyClientConfigPort)
}

func NewHAProxyClient() (*HAProxyClient, error) {
	return NewHAProxyClientTo(haproxyClientIP, haproxyClientConfigPort)
}

func (hac *HAProxyClient) getStats(statsType string) ([]StatsGroup, error) {
	resp, err := http.Get("http://" + hac.IP + ":" + hac.Port + statsEndpoint + statsType)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var statsGroup []StatsGroup
	json.Unmarshal(body, &statsGroup)
	return statsGroup, nil
}

func (hac *HAProxyClient) GetStats() ([]StatsGroup, error) {
	return hac.getStats("")
}

func (hac *HAProxyClient) GetStatsBackend() ([]StatsGroup, error) {
	return hac.getStats("/backend")
}

func (hac *HAProxyClient) GetStatsFrontend() ([]StatsGroup, error) {
	return hac.getStats("/frontend")
}

func (hac *HAProxyClient) GetStatsServer() ([]StatsGroup, error) {
	return hac.getStats("/server")
}

func (hac *HAProxyClient) GetConfig() (Config, error) {
	resp, err := http.Get("http://" + hac.IP + ":" + hac.Port + configEndpoint)
	if err != nil {
		return Config{}, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Config{}, err
	}
	var config Config
	json.Unmarshal(body, &config)
	return config, nil
}

func (hac *HAProxyClient) GetInfo() (Info, error) {
	resp, err := http.Get("http://" + hac.IP + ":" + hac.Port + infoEndpoint)
	if err != nil {
		return Info{}, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Info{}, err
	}
	var info Info
	json.Unmarshal(body, &info)
	return info, nil
}

func (hac *HAProxyClient) GetFrontEndACLs(name string) ([]ACL, error) {
	resp, err := http.Get("http://" + hac.IP + ":" + hac.Port + frontendEndpoint + "/" + name + "/acls")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var acls []ACL
	json.Unmarshal(body, &acls)
	return acls, nil
}

func (hac *HAProxyClient) PostConfig(c Config) error {
	log.Debug("Putting Config: %+v ", c)
	url := "http://" + hac.IP + ":" + hac.Port + configEndpoint
	jsonStr, err := json.Marshal(c)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(string(body))
	}
	return nil
}

func (hac *HAProxyClient) PostServerWeigth(backendName, serverName string, weight int) error {
	url := "http://" + hac.IP + ":" + hac.Port + backendEndpoint + "/" + backendName + "/servers/" + serverName + "/weight/" + strconv.Itoa(weight)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte{0}))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(string(body))
	}
	return nil
}

func (hac *HAProxyClient) PostFrontendACLPttern(frontendName, aclName, pattern string) error {
	url := "http://" + hac.IP + ":" + hac.Port + frontendEndpoint + "/" + frontendName + "/acl/" + aclName + "/" + pattern
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte{0}))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(string(body))
	}
	return nil
}

func (c *Config) DeleteBackend(containerID string) {
	for i, be := range c.Backends {
		for j, bes := range be.BackendServers {
			log.Debug("bes %+v", bes)
			if bes.HasName("docker_intent_bes_" + containerID) {
				cbes := c.Backends[i].BackendServers
				cbes[j], cbes[len(cbes)-1], cbes = cbes[len(cbes)-1], nil, cbes[:len(cbes)-1]
				return
			}
		}
	}
}

func (hac *HAProxyClient) DeleteBackend(containerID string) error {
	c, err := hac.GetConfig()
	if err != nil {
		return err
	}
	log.Debug("Config %+v", c)
	c.DeleteBackend(containerID)
	return hac.PostConfig(c)
}

func (c *Config) UpdateConfig(containerID, svcName, svcIP, hostPort, containerPort, trafficType string) error {
	log.Debug("")

	be := Backend{}
	be.Name = "docker_intent_be_" + svcName + "_" + containerPort + "_" + hostPort
	be.Mode = trafficType

	bes := BackendServer{}
	bes.Name = "docker_intent_bes_" + containerID
	bes.Weight = 100
	bes.CheckInterval = 10
	bes.MaxConn = 1000
	bes.Host = svcIP
	if i, err := strconv.Atoi(containerPort); err != nil {
		return err
	} else {
		bes.Port = i
	}
	if backend, exist := c.HasBackendWithName(be.Name); exist {
		//Backend exist on ha-proxy but we need to verify if the BackendServer exist as well
		if _, exist := backend.HasBackendServerWithName(bes.Name); !exist {
			//BackendServer doesn't exist on the given backend
			backend.BackendServers = append(backend.BackendServers, &bes)
			log.Debug("Backends is 0: %+v", c.Backends)
		}
	} else {
		//If it doesn't exist on ha-proxy we create everything
		be.BackendServers = []*BackendServer{&bes}
		c.Backends = append(c.Backends, &be)
		log.Debug("Backends are: %+v", c.Backends)
	}

	fe := Frontend{}
	fe.Name = "docker_intent_fe_" + svcName + "_" + containerPort + "_" + hostPort
	fe.Mode = trafficType
	fe.BindIp = "0.0.0.0"
	fe.DefaultBackend = be.Name
	if port, err := strconv.Atoi(hostPort); err != nil {
		return err
	} else {
		fe.BindPort = port
	}

	if _, exist := c.HasFrontendWithName(fe.Name); !exist {
		c.Frontends = append(c.Frontends, &fe)
		log.Debug("Frontends are: %+v", c.Frontends)
	}
	log.Debug("Config is %+v", c)
	return nil
}
