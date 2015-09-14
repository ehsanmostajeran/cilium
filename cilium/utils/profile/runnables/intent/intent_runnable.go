package intent

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	m "github.com/cilium-team/cilium/cilium/messages"
	u "github.com/cilium-team/cilium/cilium/utils"
	uc "github.com/cilium-team/cilium/cilium/utils/comm"
	ucdb "github.com/cilium-team/cilium/cilium/utils/comm/db"
	upl "github.com/cilium-team/cilium/cilium/utils/plugins/loadbalancer"
	up "github.com/cilium-team/cilium/cilium/utils/profile"
	upr "github.com/cilium-team/cilium/cilium/utils/profile/runnables"
	upsi "github.com/cilium-team/cilium/cilium/utils/profile/subpolicies/intent"

	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/cilium-team/go-logging"
	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/deckarep/golang-set"
	d "github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

const Name = "intent-runnable"

var (
	log          = logging.MustGetLogger("cilium")
	hookHandlers = map[string]func(ucdb.Db, *upsi.Intent, *m.CreateConfig) error{
		"pre-hookDaemonCreate":   preHookDaemonCreateExecIntentConfig,
		"pre-hookSwarmCreate":    preHookExecIntentConfig,
		"post-hookDaemonStart":   postHookExecIntentConfig,
		"post-hookDaemonRestart": postHookExecIntentConfig,
	}
	postHookHandlers = map[string]func(ucdb.Db, upsi.Intent, *m.CreateConfig, *d.Container) error{}
)

type IntentRunnable struct {
	intent *upsi.Intent
}

func (ir IntentRunnable) Exec(hookType, reqType string, db ucdb.Db, cc *m.CreateConfig) error {
	if f, ok := hookHandlers[hookType+reqType]; ok {
		f(db, ir.intent, cc)
	}
	return nil
}

func (ir IntentRunnable) GetRunnableFrom(users []up.User, policies []up.PolicySource) upr.PolicyRunnable {
	log.Debug("users %+v", users)
	isDefaultIntentConfig := true
	lastUserIntentCfgCovered := upsi.NewIntentConfig()
	usersIntentCfg := upsi.NewIntentConfig()
	usersIntentCfg.Config = upsi.Intent{}
	up.OrderUsersByDescendingID(users)
	for _, user := range users {
		log.Debug("user %+v", user)
		userPolicies := up.FilterPoliciesByUser(policies, user)
		intentConfigs := up.GetIntentConfigs(userPolicies)
		upsi.OrderIntentConfigsByAscendingPriority(intentConfigs)
		for i, iConfig := range intentConfigs {
			intentConfigs[i] = iConfig.DeepCopy()
			log.Debug("Filtered userIntentConfigs: %+v", iConfig)
		}
		//TODO: If we want we can make that every rule from a specific user
		//would be overwritten, and not merged, by a prioritary order
		userIntentCfg := upsi.NewIntentConfig()
		//TODO: If we can use what is below as it is we don't
		//need to use "NewIntentConfig()" above
		userIntentCfg.Config = upsi.Intent{}
		for _, iConfig := range intentConfigs {
			log.Debug("It's covered")
			if isDefaultIntentConfig {
				log.Debug("Isn't the default")
				isDefaultIntentConfig = false
				userIntentCfg.MergeWithOverwrite(iConfig)
			} else {
				userIntentCfg.MergeWith(iConfig)
			}
			lastUserIntentCfgCovered.Config = userIntentCfg.Config
			log.Debug("current userIntentCfg.Config: %#v", userIntentCfg.Config)
		}
		log.Debug("New userIntentCfg %#v", userIntentCfg.Config)
		log.Debug("New usersIntentCfg %#v", usersIntentCfg.Config)
		usersIntentCfg.OverwriteWith(*userIntentCfg)
		log.Debug("current usersIntentCfg.Config: %#v", usersIntentCfg.Config)
	}
	/*
		We still have to create a new "default" IntentConfig
		This way we will make sure that every configuration that
		was not configured by users will have a default value set.
	*/
	finalIntentCfg := upsi.NewIntentConfig()
	finalIntentCfg.MergeWithOverwrite(*usersIntentCfg)
	log.Info("Final intent loaded: %#v", finalIntentCfg.Config)
	return IntentRunnable{intent: &finalIntentCfg.Config}
}

func preHookDaemonCreateExecIntentConfig(conn ucdb.Db, intent *upsi.Intent, containerConfig *m.CreateConfig) error {
	log.Debug("intent %#v", intent)
	log.Debug("container config %+v", containerConfig.Config)
	if len(containerConfig.Labels) == 0 {
		return nil
	}
	//intent.AddArguments
	{
		if intent.AddArguments != nil && len(*intent.AddArguments) > 0 {
			//Replacing special keywords
			containerConfig.Cmd = make([]string, 0, len(*intent.AddArguments))
			for _, v := range *intent.AddArguments {
				str := v
				if strings.Contains(v, "$public-ip") {
					str = strings.Replace(v, "$public-ip", os.Getenv("HOST_IP"), -1)
				}
				containerConfig.Cmd = append(containerConfig.Cmd, str)
			}
			log.Debug("intent %#v", intent)
			log.Debug("containerConfig.Cmd, %+v", containerConfig.Cmd)
		}
	}

	return nil
}

func preHookExecIntentConfig(conn ucdb.Db, intent *upsi.Intent, containerConfig *m.CreateConfig) error {
	log.Debug("intent %#v", intent)
	log.Debug("container config %+v", containerConfig.Config)
	if len(containerConfig.Labels) == 0 {
		return nil
	}
	var (
		docker uc.Docker
		err    error
	)
	if docker, err = uc.NewDockerClient(); err != nil {
		return err
	}

	//intent.MaxScale
	{
		if svcName := u.LookupServiceName(containerConfig.Labels); svcName != "" {
			instancesRunning := 0
			if containers, err := docker.ListContainers(d.ListContainersOptions{All: true}); err != nil {
				return err
			} else {
				for _, container := range containers {
					if dockerContainer, err := docker.InspectContainer(container.ID); err != nil {
						return err
					} else {
						if dockerServiceName := u.LookupServiceName(dockerContainer.Config.Labels); dockerServiceName != "" {
							if svcName == dockerServiceName {
								instancesRunning++
								log.Debug("intent.MaxScale  %+v", intent.MaxScale)
								log.Debug("instancesRunning %+v", instancesRunning)
								if instancesRunning >= *intent.MaxScale {
									log.Info("Reached maximum scalability for containers with labels: %s", dockerContainer.Config.Labels)
									return errors.New(fmt.Sprintf("Reached maximum scalability for containers with labels: %s", dockerContainer.Config.Labels))
								}
							}
						}
					}
				}
			}
		}
	}
	//intent.HostnameIs
	{
		log.Debug("containerConfig.Hostname %+v", containerConfig.Hostname)
		log.Debug("intent.HostnameIs %+v", intent.HostNameIs)
		if newHostName := intent.GetHostNameFromLabels(containerConfig.Labels); newHostName != "" {
			containerConfig.Hostname = newHostName
			log.Info("New hostname %s for container %s", newHostName, containerConfig.Name)
		}
		log.Debug("containerConfig.Hostname %+v", containerConfig.Hostname)
	}
	//intent.RemoveDockerLinks
	{
		if *intent.RemoveDockerLinks && containerConfig.HostConfig != nil {
			if err = conn.PutDockerLinksOfContainerTemp(up.ContainerLinks{Container: containerConfig.Name, Links: containerConfig.HostConfig.Links}); err != nil {
				return err
			}
			log.Info("Removed docker links for container %s", containerConfig.Name)
			containerConfig.HostConfig.Links = nil
		}
	}
	//intent.RemovePortBindings
	{
		if *intent.RemovePortBindings && containerConfig.HostConfig != nil {
			if err = conn.PutDockerPortBindingsOfContainerTemp(up.ContainerPortBindings{Container: containerConfig.Name, PortBindings: containerConfig.HostConfig.PortBindings}); err != nil {
				return err
			}
			log.Info("Removed PortBindings for container %s", containerConfig.Name)
			containerConfig.HostConfig.PortBindings = nil
		}
	}

	return nil
}

func postHookExecIntentConfig(dbConn ucdb.Db, intent *upsi.Intent, containerConfig *m.CreateConfig) error {
	log.Debug("intent: %#v", intent)
	log.Debug("container config %+v", containerConfig.Config)
	if len(containerConfig.Labels) == 0 {
		return nil
	}

	//intent.Netconf
	{
		log.Debug("intent.Netconf %+v", intent.NetConf)
		if *intent.NetConf.CIDR != "" {
			if ip, ipnet, err := GetIPFrom(dbConn, *intent.NetConf.CIDR); err != nil {
				return err
			} else {
				//Create bridge for this container
				ifname, mac, err := u.CreateBridge(ip, *ipnet, intent.NetConf, containerConfig.State.Pid, containerConfig.ID)
				if err != nil {
					dbConn.DeleteIP(ip)
					log.Error("Fail while setting up networking for container %s: %s", containerConfig.ID, err)
					return err
				}

				//intent.AttachToDNS
				if *intent.AddToDNS {
					hostname := intent.GetHostNameFromLabels(containerConfig.Labels)
					domains := []string{hostname}
					if containerConfig.ID != "" {
						//This prevents dns to return "description": "Domain name parsing failed label empty or too long"
						//containerDomains.Domains = append(containerDomains.Domains, dockerServerResponse.ID)
						domains = append(domains, containerConfig.ID[0:12])
					}
					ipsString := []string{ip.String()}
					if dnsClient, err := dbConn.GetDNSConfig(); err != nil {
						dbConn.DeleteIP(ip)
						return err
					} else {
						if err := dnsClient.SendToDNS(domains, ipsString); err != nil {
							dbConn.DeleteIP(ip)
							return err
						}
					}
				}
				var (
					group     int
					bd        int
					namespace int
				)
				if intent.NetConf.Group != nil {
					group = *intent.NetConf.Group
				}
				if intent.NetConf.BD != nil {
					bd = *intent.NetConf.BD
				}
				if intent.NetConf.Namespace != nil {
					namespace = *intent.NetConf.Namespace
				}
				endpoint := up.Endpoint{
					Container: containerConfig.ID,
					IPs:       []net.IP{ip},
					MACs:      []string{mac},
					Node:      os.Getenv("HOST_IP"),
					Interface: ifname,
					Group:     group,
					BD:        bd,
					Namespace: namespace,
				}

				if svcName := u.LookupServiceName(containerConfig.Labels); svcName != "" {
					endpoint.Service = svcName
				}

				if err := dbConn.PutEndpoint(endpoint); err != nil {
					log.Error("Fail while inserting endpoint %s", err)
					dbConn.DeleteIP(ip)
					return err
				}

				if *intent.RemoveDockerLinks {
					//Deal with links between containers
					var (
						containerLinks      up.ContainerLinks
						containersLinked    mapset.Set
						containersLinkedIPs []net.IP
						err                 error
					)
					if containerLinks, err = dbConn.GetDockerLinksOfContainerTemp(containerConfig.Name); err != nil {
						dbConn.DeleteEndpoint(containerConfig.ID)
						return err
					}
					log.Debug("links %+v", containerLinks.Links)
					if len(containerLinks.Links) != 0 {
						if err = dbConn.PutDockerLinksOfContainer(up.ContainerLinks{Container: containerConfig.ID, Links: containerLinks.Links}); err != nil {
							dbConn.DeleteEndpoint(containerConfig.ID)
							dbConn.DeleteIP(ip)
							return err
						}
						containerConfig.HostConfig.Links = containerLinks.Links
						containersLinked = mapset.NewSet()
						for _, link := range containerLinks.Links {
							container, _ := uc.SplitLink(link)
							containersLinked.Add(container)
						}
						for _, container := range containersLinked.ToSlice() {
							if containerstr, ok := container.(string); ok {
								if ips, err := dbConn.GetEndpoint(containerstr); err != nil {
									dbConn.DeleteEndpoint(containerConfig.ID)
									dbConn.DeleteIP(ip)
									return err
								} else {
									containersLinkedIPs = append(containersLinkedIPs, ips.IPs...)
								}
							}
						}
						freshRules := CreateOVSRules(ip, containersLinkedIPs)
						*intent.NetPolicy.OVSConfig.Rules = append(*intent.NetPolicy.OVSConfig.Rules, freshRules...)
					}
				}

				//intent.NetPolicy
				ForceNetworkRules(intent)

				//HA-Proxy
				if *intent.MaxScale > 1 {
					if svcName := u.LookupServiceName(containerConfig.Labels); svcName != "" {
						switch *intent.LoadBalancer.Name {
						case "ha-proxy":
							var (
								haproxyCli  upl.HAProxyClient
								localConfig upl.Config
								err         error
							)
							if haproxyCli, err = dbConn.GetHAProxyConfig(); err != nil {
								return err
							}
							if localConfig, err = haproxyCli.GetConfig(); err != nil {
								return err
							}
							if *intent.LoadBalancer.BindPort == 0 {
								log.Debug("Config is %d", intent.LoadBalancer.BindPort)
								if containerPortBindings, err := dbConn.GetDockerPortBindingsOfContainerTemp(containerConfig.Name); err != nil {
									for container, hosts := range containerPortBindings.PortBindings {
										for _, h := range hosts {
											localConfig.UpdateConfig(containerConfig.ID, svcName, ip.String(), h.HostPort, container.Port(), *intent.LoadBalancer.TrafficType)
										}
									}
									if err = dbConn.PutDockerPortBindingsOfContainer(containerPortBindings); err != nil {
										return err
									}
								}
							} else {
								strport := strconv.Itoa(*intent.LoadBalancer.BindPort)
								localConfig.UpdateConfig(containerConfig.ID, svcName, ip.String(), strport, strport, *intent.LoadBalancer.TrafficType)
								log.Debug("Config is %+v", localConfig)
							}
							log.Debug("Config is %+v", localConfig)
							if err = haproxyCli.PostConfig(localConfig); err != nil {
								return err
							}
						}
					}
				}
			}
		}
	}

	//Update configurations for special containers
	{
		for k, v := range containerConfig.Config.Labels {
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

	return nil
}
