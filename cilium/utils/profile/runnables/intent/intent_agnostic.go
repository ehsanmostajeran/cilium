package intent

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	u "github.com/cilium-team/cilium/cilium/utils"
	uc "github.com/cilium-team/cilium/cilium/utils/comm"
	ucdb "github.com/cilium-team/cilium/cilium/utils/comm/db"
	upl "github.com/cilium-team/cilium/cilium/utils/plugins/loadbalancer"
	up "github.com/cilium-team/cilium/cilium/utils/profile"
	upsi "github.com/cilium-team/cilium/cilium/utils/profile/subpolicies/intent"
)

func addArgumentsToCmd(intent *upsi.Intent, cmd []string) []string {
	if intent.AddArguments == nil || len(*intent.AddArguments) == 0 {
		return cmd
	}
	//Replacing with special keywords
	for _, v := range *intent.AddArguments {
		str := v
		if strings.Contains(v, "$public-ip") {
			str = strings.Replace(v, "$public-ip", os.Getenv("HOST_IP"), -1)
		}
		cmd = append(cmd, str)
	}
	log.Debug("intent %#v", intent)
	log.Debug("cmd, %+v", cmd)
	return cmd
}

func hostnameIs(intent *upsi.Intent, labels map[string]string, oldHostName string) string {
	log.Debug("intent.HostnameIs %+v", intent.HostNameIs)
	if newHostName := intent.GetHostNameFromLabels(labels); newHostName != "" {
		return newHostName
	}
	return oldHostName
}

func saveEndpoint(dbConn ucdb.Db, intent *upsi.Intent, labels map[string]string,
	containerID string, ifname string, ips []net.IP, macs []string) error {

	endpoint := up.Endpoint{}
	endpoint.Container = containerID
	endpoint.IPs = ips
	endpoint.MACs = macs
	endpoint.Node = os.Getenv("HOST_IP")
	endpoint.Interface = ifname
	if intent.NetConf.Group != nil {
		endpoint.Group = *intent.NetConf.Group
	}
	if intent.NetConf.BD != nil {
		endpoint.BD = *intent.NetConf.BD
	}
	if intent.NetConf.Namespace != nil {
		endpoint.Namespace = *intent.NetConf.Namespace
	}
	if svcName := u.LookupServiceName(labels); svcName != "" {
		endpoint.Service = svcName
	}

	return dbConn.PutEndpoint(endpoint)
}

func addToDNS(dbConn ucdb.Db, intent *upsi.Intent, labels map[string]string, containerID string, ips []net.IP) error {
	if !*intent.AddToDNS {
		return nil
	}

	hostname := intent.GetHostNameFromLabels(labels)
	domains := []string{hostname}
	if containerID != "" {
		//This prevents dns to return "description": "Domain name parsing failed label empty or too long"
		//containerDomains.Domains = append(containerDomains.Domains, dockerServerResponse.ID)
		domains = append(domains, containerID[0:12])
	}
	dnsClient, err := dbConn.GetDNSConfig()
	if err != nil {
		return err
	}

	ipsString := []string{}
	for _, ip := range ips {
		ipsString = append(ipsString, ip.String())
	}

	if err := dnsClient.SendToDNS(domains, ipsString); err != nil {
		return err
	}
	return nil
}

func addToLoadBalancer(dbConn ucdb.Db, intent *upsi.Intent,
	ips []net.IP, contID, contName string, labels map[string]string) error {

	// We won't load balance services with the max scale less or equal than 1.
	if *intent.MaxScale <= 1 {
		return nil
	}
	svcName := u.LookupServiceName(labels)
	if svcName == "" {
		return nil
	}
	switch *intent.LoadBalancer.Name {
	case "ha-proxy":
		haproxyCli, err := dbConn.GetHAProxyConfig()
		if err != nil {
			return err
		}
		localConfig, err := haproxyCli.GetConfig()
		if err != nil {
			return err
		}
		if *intent.LoadBalancer.BindPort == 0 {
			log.Debug("Config is %d", intent.LoadBalancer.BindPort)
			containerPortBindings, err := dbConn.GetDockerPortBindingsOfContainerTemp(contName)
			if err != nil {
				return err
			}
			for container, hosts := range containerPortBindings.PortBindings {
				for _, h := range hosts {
					for _, ip := range ips {
						localConfig.UpdateConfig(contID, svcName, ip.String(), h.HostPort, container.Port(), *intent.LoadBalancer.TrafficType)
					}
				}
			}
			if err = dbConn.PutDockerPortBindingsOfContainer(containerPortBindings); err != nil {
				return err
			}
		} else {
			containerPort := strconv.Itoa(*intent.LoadBalancer.BindPort)
			for _, ip := range ips {
				localConfig.UpdateConfig(contID, svcName, ip.String(), containerPort, containerPort, *intent.LoadBalancer.TrafficType)
			}
		}
		log.Debug("Config is %+v", localConfig)
		if err = haproxyCli.PostConfig(localConfig); err != nil {
			return err
		}
	default:
		return fmt.Errorf("LoadBalancer '%s' unknown", *intent.LoadBalancer.Name)
	}
	return nil
}

func saveCiliumServices(dbConn ucdb.Db, labels map[string]string) {
	for k, v := range labels {
		if k == "com.intent.service" {
			switch v {
			case "svc_dns":
				dnsClient := uc.NewDNSClientToIP(os.Getenv("HOST_IP"))
				dbConn.PutDNSConfig(dnsClient)
			case "svc_loadbalancer":
				haproxyClient, _ := upl.NewHAProxyClientToIP(os.Getenv("HOST_IP"))
				dbConn.PutHAProxyConfig(*haproxyClient)
			}
		}
	}
}
