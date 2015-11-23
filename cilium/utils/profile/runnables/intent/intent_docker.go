package intent

import (
	"fmt"
	"net"

	m "github.com/cilium-team/cilium/cilium/messages"
	u "github.com/cilium-team/cilium/cilium/utils"
	uc "github.com/cilium-team/cilium/cilium/utils/comm"
	ucdb "github.com/cilium-team/cilium/cilium/utils/comm/db"
	up "github.com/cilium-team/cilium/cilium/utils/profile"
	upsi "github.com/cilium-team/cilium/cilium/utils/profile/subpolicies/intent"

	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/deckarep/golang-set"
	d "github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

func preHookDockerDaemonCreate(conn ucdb.Db, intent *upsi.Intent, containerConfig *m.DockerCreateConfig) error {
	log.Debug("intent %#v", intent)
	log.Debug("container config %+v", containerConfig.Config)
	if len(containerConfig.Labels) == 0 {
		return nil
	}

	//intent.AddArguments
	containerConfig.Cmd = addArgumentsToCmd(intent, containerConfig.Cmd)
	return nil
}

func preHookDockerSwarmCreate(conn ucdb.Db, intent *upsi.Intent, containerConfig *m.DockerCreateConfig) error {
	log.Debug("intent %#v", intent)
	log.Debug("container config %+v", containerConfig.Config)
	if len(containerConfig.Labels) == 0 {
		return nil
	}
	docker, err := uc.NewDockerClient()
	if err != nil {
		return err
	}

	//intent.MaxScale
	if err := maxScaleDocker(docker, intent, containerConfig.Labels); err != nil {
		return err
	}

	//intent.HostnameIs
	log.Debug("containerConfig.Hostname %+v", containerConfig.Hostname)
	containerConfig.Hostname = hostnameIs(intent, containerConfig.Labels, containerConfig.Hostname)
	log.Debug("containerConfig.Hostname %+v", containerConfig.Hostname)

	//intent.RemoveDockerLinks
	if err := removeDockerLinksDocker(conn, intent, containerConfig); err != nil {
		return err
	}

	//intent.RemovePortBindings
	if err := removePortBindingsDocker(conn, intent, containerConfig); err != nil {
		return err
	}

	return nil
}

func postHookDockerDaemonStart(dbConn ucdb.Db, intent *upsi.Intent, containerConfig *m.DockerCreateConfig) error {
	log.Debug("intent: %#v", intent)
	log.Debug("container config %+v", containerConfig.Config)
	if len(containerConfig.Labels) == 0 {
		return nil
	}

	//intent.Netconf
	if err := netConfDocker(dbConn, intent, containerConfig); err != nil {
		return err
	}

	//Update configurations for special containers
	saveCiliumServices(dbConn, containerConfig.Labels)

	log.Info("container successfully configured: %+v", containerConfig.ID)

	return nil
}

func maxScaleDocker(docker uc.Docker, intent *upsi.Intent, labels map[string]string) error {
	svcName := u.LookupServiceName(labels)
	if svcName == "" {
		return nil
	}
	containers, err := docker.ListContainers(d.ListContainersOptions{All: true})
	if err != nil {
		return err
	}
	instancesRunning := 0
	for _, container := range containers {
		dockerContainer, err := docker.InspectContainer(container.ID)
		if err != nil {
			return err
		}
		dockerServiceName := u.LookupServiceName(dockerContainer.Config.Labels)
		if dockerServiceName != "" && svcName == dockerServiceName {
			instancesRunning++
			log.Debug("intent.MaxScale  %+v", intent.MaxScale)
			log.Debug("instancesRunning %+v", instancesRunning)
			if instancesRunning >= *intent.MaxScale {
				log.Warning("Reached maximum scalability for containers with labels: %s", labels)
				return fmt.Errorf("Reached maximum scalability for containers with labels: %s", labels)
			}
		}
	}
	return nil
}

func removeDockerLinksDocker(conn ucdb.Db, intent *upsi.Intent, containerConfig *m.DockerCreateConfig) error {
	if !*intent.RemoveDockerLinks || containerConfig.HostConfig == nil {
		return nil
	}
	cl := up.ContainerLinks{Container: containerConfig.Name, Links: containerConfig.HostConfig.Links}
	if err := conn.PutDockerLinksOfContainerTemp(cl); err != nil {
		return err
	}
	containerConfig.HostConfig.Links = nil
	log.Info("Removed docker links for container %s", containerConfig.Name)
	return nil
}

func removePortBindingsDocker(conn ucdb.Db, intent *upsi.Intent, containerConfig *m.DockerCreateConfig) error {
	if !*intent.RemovePortBindings || containerConfig.HostConfig == nil {
		return nil
	}
	cpb := up.ContainerPortBindings{Container: containerConfig.Name, PortBindings: containerConfig.HostConfig.PortBindings}
	if err := conn.PutDockerPortBindingsOfContainerTemp(cpb); err != nil {
		return err
	}
	containerConfig.HostConfig.PortBindings = nil
	log.Info("Removed PortBindings for container %s", containerConfig.Name)
	return nil
}

func netConfDocker(dbConn ucdb.Db, intent *upsi.Intent, containerConfig *m.DockerCreateConfig) error {
	log.Debug("intent.Netconf %+v", intent.NetConf)

	if *intent.NetConf.CIDR == "" {
		return nil
	}

	ip, ipnet, err := getIPFrom(dbConn, *intent.NetConf.CIDR)
	if err != nil {
		return err
	}

	//Create bridge for this container
	ifname, mac, err := u.CreateBridge(ip, *ipnet, intent.NetConf, containerConfig.State.Pid, containerConfig.ID)
	if err != nil {
		dbConn.DeleteIP(ip)
		log.Error("Fail while setting up networking for container %s: %s", containerConfig.ID, err)
		return err
	}

	//Save this container's endpoint
	if err := saveEndpoint(dbConn, intent, containerConfig.Labels, containerConfig.ID, ifname, []net.IP{ip}, []string{mac}); err != nil {
		dbConn.DeleteIP(ip)
		log.Error("Fail while saving up endpoint for container %s: %s", containerConfig.ID, err)
		return err
	}

	//intent.AddToDNS
	if err := addToDNS(dbConn, intent, containerConfig.Labels, containerConfig.ID, []net.IP{ip}); err != nil {
		dbConn.DeleteEndpoint(containerConfig.ID)
		dbConn.DeleteIP(ip)
		log.Error("Fail while setting up DNS entries for container %s: %s", containerConfig.ID, err)
		return err
	}

	// Deal with links between containers, which only happens if they were
	// previously removed (on pre-hook)
	if err := resolveMultiContainerLinksDocker(dbConn, intent, containerConfig, []net.IP{ip}); err != nil {
		dbConn.DeleteEndpoint(containerConfig.ID)
		dbConn.DeleteIP(ip)
		log.Error("Fail while setting resolving multi container links for container %s: %s", containerConfig.ID, err)
		return err
	}

	//intent.NetPolicy
	if err := forceNetworkRules(intent); err != nil {
		log.Error("Fail creating network rules for %s: %s", containerConfig.ID, err)
	}

	//intent.LoadBalancer
	if err := addToLoadBalancer(dbConn, intent, []net.IP{ip}, containerConfig.ID,
		containerConfig.Name, containerConfig.Labels); err != nil {

		log.Error("Fail while adding container to load balancer %s: %s", containerConfig.ID, err)
	}

	return nil
}

func resolveMultiContainerLinksDocker(dbConn ucdb.Db, intent *upsi.Intent, containerConfig *m.DockerCreateConfig, ips []net.IP) error {
	if !*intent.RemoveDockerLinks {
		return nil
	}

	containerLinks, err := dbConn.GetDockerLinksOfContainerTemp(containerConfig.Name)
	if err != nil {
		return err
	}

	log.Debug("links %+v", containerLinks.Links)
	if len(containerLinks.Links) != 0 {
		return nil
	}

	cl := up.ContainerLinks{Container: containerConfig.ID, Links: containerLinks.Links}
	if err := dbConn.PutDockerLinksOfContainer(cl); err != nil {
		return err
	}

	containerConfig.HostConfig.Links = containerLinks.Links

	containersLinked := mapset.NewSet()
	for _, link := range containerLinks.Links {
		container, _ := uc.SplitLink(link)
		containersLinked.Add(container)
	}

	containersLinkedIPs := []net.IP{}
	for _, container := range containersLinked.ToSlice() {
		if containerstr, ok := container.(string); ok {
			if ips, err := dbConn.GetEndpoint(containerstr); err != nil {
				return err
			} else {
				containersLinkedIPs = append(containersLinkedIPs, ips.IPs...)
			}
		}
	}
	for _, ip := range ips {
		freshRules := createOVSRules(ip, containersLinkedIPs)
		*intent.NetPolicy.OVSConfig.Rules = append(*intent.NetPolicy.OVSConfig.Rules, freshRules...)
	}
	return nil
}
