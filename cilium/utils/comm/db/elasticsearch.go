package db

import (
	"encoding/json"
	"errors"
	l "log"
	"net"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"

	uc "github.com/cilium-team/cilium/cilium/utils/comm"
	upl "github.com/cilium-team/cilium/cilium/utils/plugins/loadbalancer"
	up "github.com/cilium-team/cilium/cilium/utils/profile"

	"github.com/cilium-team/cilium/Godeps/_workspace/src/gopkg.in/olivere/elastic.v2"
)

type EConn struct {
	*elastic.Client
}

const (
	elasticDefaultPort = "9200"
	elasticDefaultIP   = "127.0.0.1"
	IndexConfig        = "cilium-configs"
	IndexState         = "cilium-state"
	logNameTimeFormat  = time.RFC3339
)

var (
	ec         EConn
	clientInit sync.Once
	Indexes    = [...]string{IndexConfig, IndexState}
)

func InitElasticDb() error {
	c, err := NewElasticConn()
	if err != nil {
		return err
	}
	defer c.Close()

	for _, index := range Indexes {
		if _, err = c.DeleteIndex(index).Do(); err != nil {
			return err
		}
		if _, err = c.CreateIndex(index).Do(); err != nil {
			return err
		}
	}

	return nil
}

func ElasticFlushConfig() error {
	c, err := NewElasticConn()
	if err != nil {
		return err
	}
	defer c.Close()

	if _, err = c.DeleteIndex(IndexConfig).Do(); err != nil {
		return err
	}
	if _, err = c.CreateIndex(IndexConfig).Do(); err != nil {
		return err
	}

	return nil
}

func NewElasticConn() (EConn, error) {
	log.Debug("")
	port := os.Getenv("ELASTIC_PORT")
	if port == "" {
		port = elasticDefaultPort
	}
	ip := os.Getenv("ELASTIC_IP")
	if ip == "" {
		ip = elasticDefaultIP
	}
	return NewElasticConnTo(ip, port)
}

func NewElasticConnTo(ip, port string) (EConn, error) {
	log.Debug("")
	var outerr error
	clientInit.Do(func() {
		logTimename := time.Now().Format(logNameTimeFormat)
		fo, err := os.Create(os.TempDir() + "/cilium-elastic-out-" + logTimename + ".log")
		if err != nil {
			l.Fatalf("Error while creating a log file: %s", err)
		}
		fe, err := os.Create(os.TempDir() + "/cilium-elastic-error-" + logTimename + ".log")
		if err != nil {
			l.Fatalf("Error while creating a log file: %s", err)
		}
		//		ft, err := os.Create(os.TempDir() + "/cilium-elastic-trace-" + logTimename + ".log")
		//		if err != nil {
		//			l.Fatalf("Error while creating a log file: %s", err)
		//		}
		l.Printf("Trying to connect to ElasticSearch to %s, %s\n", ip, port)

		ec.Client, err = elastic.NewClient(
			elastic.SetURL("http://"+ip+":"+port),
			elastic.SetMaxRetries(10),
			elastic.SetHealthcheckTimeoutStartup(60*time.Second),
			elastic.SetSniff(false),
			elastic.SetErrorLog(l.New(fe, "", l.LstdFlags)),
			elastic.SetInfoLog(l.New(fo, "", l.LstdFlags)),
			//elastic.SetTraceLog(l.New(ft, "", l.LstdFlags)),
		)
		if err == nil {
			l.Printf("Success!\n")
		} else {
			l.Printf("Error %+v\n", err)
		}
		outerr = err
	})
	return ec, outerr
}

func (c EConn) GetName() (string, error) {
	nir, err := c.NodesInfo().NodeId("_local").Do()
	if err != nil {
		return "", err
	}
	for _, val := range nir.Nodes {
		return val.Name, nil
	}
	return "", nil
}

func (c EConn) Close() {
}

func (c EConn) GetUsers() ([]up.User, error) {
	log.Debug("")
	searchResult, err := c.Search().Index(IndexConfig).Type(TNUsers).Do()
	if err != nil {
		return nil, err
	}
	var (
		user  up.User
		users []up.User
	)
	for _, item := range searchResult.Each(reflect.TypeOf(user)) {
		if u, ok := item.(up.User); ok {
			users = append(users, u)
		}
	}
	return users, nil
}

func (c EConn) GetDNSConfig() (uc.DNSClient, error) {
	log.Debug("")
	searchResult, err := c.Search().Index(IndexConfig).Type(TNDNSconfig).Do()
	var dnsConfig uc.DNSClient
	if err != nil {
		return dnsConfig, err
	}
	for _, item := range searchResult.Each(reflect.TypeOf(dnsConfig)) {
		if dnsConfig, ok := item.(uc.DNSClient); ok {
			return dnsConfig, nil
		}
	}
	return dnsConfig, nil
}

func (c EConn) GetHAProxyConfig() (upl.HAProxyClient, error) {
	log.Debug("")
	var hAProxyClient upl.HAProxyClient
	searchResult, err := c.Search().Index(IndexConfig).Type(TNHAProxyconfig).Do()
	if err != nil {
		return upl.HAProxyClient{}, err
	}
	for _, item := range searchResult.Each(reflect.TypeOf(hAProxyClient)) {
		if hAProxyClient, ok := item.(upl.HAProxyClient); ok {
			return hAProxyClient, nil
		}
	}
	return hAProxyClient, nil
}

func (c EConn) GetDockerLinksOfContainerTemp(containerName string) (up.ContainerLinks, error) {
	log.Debug("")
	var linksConfig up.ContainerLinks
	getResult, err := c.Get().Index(IndexState).Type(TNLinksConfigTemp).Id(url.QueryEscape(containerName)).Do()
	if err != nil {
		return linksConfig, err
	}
	if getResult.Found {
		if err := json.Unmarshal(*getResult.Source, &linksConfig); err != nil {
			return linksConfig, err
		}
	}
	return linksConfig, nil
}

func (c EConn) GetDockerLinksOfContainer(containerID string) (up.ContainerLinks, error) {
	log.Debug("")
	var linksConfig up.ContainerLinks
	getResult, err := c.Get().Index(IndexState).Type(TNLinksConfig).Id(url.QueryEscape(containerID)).Do()
	if err != nil {
		return linksConfig, err
	}
	if getResult.Found {
		if err := json.Unmarshal(*getResult.Source, &linksConfig); err != nil {
			return linksConfig, err
		}
	}
	return linksConfig, nil
}

func (c EConn) GetEndpoint(containerID string) (up.Endpoint, error) {
	log.Debug("")
	var ipConfig up.Endpoint
	getResult, err := c.Get().Index(IndexState).Type(TNEndpoint).Id(url.QueryEscape(containerID)).Do()
	if err != nil {
		return ipConfig, err
	}
	if getResult.Found {
		if err := json.Unmarshal(*getResult.Source, &ipConfig); err != nil {
			return ipConfig, err
		}
	}
	return ipConfig, nil
}

func (c EConn) GetDockerPortBindingsOfContainerTemp(containerID string) (up.ContainerPortBindings, error) {
	log.Debug("")
	var portBindings up.ContainerPortBindings
	getResult, err := c.Get().Index(IndexState).Type(TNPortBindingsConfigTemp).Id(url.QueryEscape(containerID)).Do()
	if err != nil {
		return portBindings, err
	}
	if getResult.Found {
		if err := json.Unmarshal(*getResult.Source, &portBindings); err != nil {
			return portBindings, err
		}
	}
	return portBindings, nil
}

func (c EConn) GetDockerPortBindingsOfContainer(containerID string) (up.ContainerPortBindings, error) {
	log.Debug("")
	var portBindings up.ContainerPortBindings
	getResult, err := c.Get().Index(IndexState).Type(TNPortBindingsConfig).Id(url.QueryEscape(containerID)).Do()
	if err != nil {
		return portBindings, err
	}
	if getResult.Found {
		if err := json.Unmarshal(*getResult.Source, &portBindings); err != nil {
			return portBindings, err
		}
	}
	return portBindings, nil
}

func (c EConn) PutUser(userName string) (bool, error) {
	log.Debug("userName: %+v", userName)
	users, err := c.GetUsers()
	if err != nil {
		return false, err
	}
	up.OrderUsersByAscendingID(users)
	userID, isNewUser := up.GetUserID(userName, users)
	if isNewUser {
		usr := up.User{ID: userID, Name: userName}
		if _, err := c.Index().Index(IndexConfig).Type(TNUsers).Refresh(true).Id(url.QueryEscape(strconv.Itoa(userID))).BodyJson(usr).Do(); err != nil {
			return isNewUser, err
		}
	}
	return isNewUser, nil
}

func (c EConn) PutDNSConfig(dnsConfig uc.DNSClient) error {
	log.Debug("")
	_, err := c.Index().Index(IndexConfig).Type(TNDNSconfig).Refresh(true).Id(url.QueryEscape(TNDNSconfig)).BodyJson(dnsConfig).Do()
	if err != nil {
		return err
	}
	return nil
}

func (c EConn) PutHAProxyConfig(haProxyClient upl.HAProxyClient) error {
	log.Debug("")
	_, err := c.Index().Index(IndexConfig).Type(TNHAProxyconfig).Refresh(true).Id(url.QueryEscape(TNHAProxyconfig)).BodyJson(haProxyClient).Do()
	if err != nil {
		return err
	}
	return nil
}

func (c EConn) PutDockerLinksOfContainer(containerLinks up.ContainerLinks) error {
	log.Debug("")
	_, err := c.Index().Index(IndexState).Type(TNLinksConfig).Refresh(true).Id(url.QueryEscape(containerLinks.Container)).BodyJson(containerLinks).Do()
	if err != nil {
		return err
	}
	return nil
}

func (c EConn) PutDockerLinksOfContainerTemp(containerLinks up.ContainerLinks) error {
	log.Debug("")
	_, err := c.Index().Index(IndexState).Type(TNLinksConfigTemp).Refresh(true).Id(url.QueryEscape(containerLinks.Container)).BodyJson(containerLinks).Do()
	if err != nil {
		return err
	}
	return nil
}

func (c EConn) PutDockerPortBindingsOfContainerTemp(portBindings up.ContainerPortBindings) error {
	log.Debug("")
	_, err := c.Index().Index(IndexState).Type(TNPortBindingsConfigTemp).Refresh(true).Id(url.QueryEscape(portBindings.Container)).BodyJson(portBindings).Do()
	if err != nil {
		return err
	}
	return nil
}

func (c EConn) PutDockerPortBindingsOfContainer(portBindings up.ContainerPortBindings) error {
	log.Debug("")
	_, err := c.Index().Index(IndexState).Type(TNPortBindingsConfig).Refresh(true).Id(url.QueryEscape(portBindings.Container)).BodyJson(portBindings).Do()
	if err != nil {
		return err
	}
	return nil
}

func (c EConn) PutIP(ip net.IP) error {
	log.Debug("")
	dbIP := up.IP{IPAddress: up.IPAddress(ip)}
	log.Debug("ipStr %+v", ip.String())
	result, err := c.Index().Index(IndexState).Type(TNIPsinUse).Refresh(true).Id(url.QueryEscape(url.QueryEscape(ip.String()))).BodyJson(dbIP).Do()
	if err != nil {
		return err
	}
	if !result.Created {
		return errors.New("IP already in use")
	}
	return nil
}

func (c EConn) DeleteIP(ip net.IP) error {
	log.Debug("ipStr %+v", ip.String())
	_, err := c.Delete().Index(IndexState).Type(TNIPsinUse).Refresh(true).Id(url.QueryEscape(url.QueryEscape(ip.String()))).Do()
	if err != nil {
		return err
	}
	return nil
}

func (c EConn) PutEndpoint(endpoint up.Endpoint) error {
	log.Debug("Endpoint %+v\n", endpoint)
	_, err := c.Index().Index(IndexState).Type(TNEndpoint).Refresh(true).Id(url.QueryEscape(endpoint.Container)).BodyJson(endpoint).Do()
	if err != nil {
		return err
	}
	return nil
}

func (c EConn) DeleteEndpoint(containerID string) error {
	log.Debug("containerID %+v\n", containerID)
	_, err := c.Delete().Index(IndexState).Type(TNEndpoint).Refresh(true).Id(url.QueryEscape(containerID)).Do()
	if err != nil {
		return err
	}
	return nil
}

func (c EConn) GetPoliciesThatCovers(labels map[string]string) ([]up.PolicySource, error) {
	log.Debug("")
	policiesMap := make(map[string]*up.PolicySource)
	searchResult, err := c.Search().Index(IndexConfig).Type(TNPolicySource).Do()
	if err != nil {
		return nil, err
	}
	if searchResult.Hits != nil {
		for _, hit := range searchResult.Hits.Hits {
			var dbPolicy up.Policy
			if json.Unmarshal(*hit.Source, &dbPolicy) != nil {
				continue
			}
			if dbPolicy.Coverage.Covers(labels) {
				owner := dbPolicy.Owner
				if _, ok := policiesMap[owner]; !ok {
					policiesMap[owner] = &up.PolicySource{Owner: owner}
				}
				policiesMap[owner].Policies = append(policiesMap[owner].Policies, dbPolicy)
			}
		}
	}
	var policies []up.PolicySource
	for _, v := range policiesMap {
		policies = append(policies, *v)
	}
	log.Debug("policies %+v", policies)
	return policies, err
}

func (c EConn) PutPolicy(policies up.PolicySource) error {
	log.Debug("policies %+v\n", policies)
	for _, policy := range policies.Policies {
		policy.Owner = url.QueryEscape(policies.Owner)
		_, err := c.Index().Index(IndexConfig).Type(TNPolicySource).Refresh(true).Id(url.QueryEscape(policy.Name)).BodyJson(policy).Do()
		if err != nil {
			return err
		}
	}
	return nil
}
