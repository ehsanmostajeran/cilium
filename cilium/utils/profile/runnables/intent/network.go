package intent

import (
	"fmt"
	"io/ioutil"
	"net"
	"os/exec"

	ucdb "github.com/cilium-team/cilium/cilium/utils/comm/db"
	upsi "github.com/cilium-team/cilium/cilium/utils/profile/subpolicies/intent"
)

func forceNetworkRules(intent *upsi.Intent) error {
	log.Debug("intent %#v\n", intent)
	//Install OVS Rules
	return forceOVSRules(*intent.NetConf.Br, intent.NetPolicy.OVSConfig)
}

func forceOVSRules(bridge string, ovsConfig upsi.OVSConfig) error {
	log.Debug("bridge %+v ovsConfig %+v\n", bridge, ovsConfig)
	if bridge == "" {
		return nil
	}
	for _, rule := range *ovsConfig.Rules {
		//ovs-ofctl add-flow $BRIDGE $RULE
		ovsAddFlow := fmt.Sprintf("ovs-ofctl add-flow %s %s", bridge, rule)
		if _, err := execShCommand(ovsAddFlow); err != nil {
			log.Error("Error while adding OVS rule: ", err)
		}
	}
	return nil
}

func createOVSRules(originIP net.IP, destinationIPs []net.IP) []string {
	log.Debug("originIP %+v destinationIPs %+v", originIP, destinationIPs)
	rulesCreated := []string{}
	for _, destinationIP := range destinationIPs {
		origToDest := fmt.Sprintf("priority=100,ip,nw_src=%s,nw_dst=%s,actions=NORMAL",
			originIP.String(), destinationIP.String())
		destToOrig := fmt.Sprintf("priority=100,ip,nw_src=%s,nw_dst=%s,actions=NORMAL",
			destinationIP.String(), originIP.String())
		rulesCreated = append(rulesCreated, origToDest, destToOrig)
	}
	return rulesCreated
}

// getIPFrom gets a valid and unused IP address and IPNet from the given ipStr.
// If the ipStr is a IP address, returnsTODO finish
func getIPFrom(dbConn ucdb.Db, ipStr string) (net.IP, *net.IPNet, error) {
	log.Debug("")
	incIP := func(ip net.IP) {
		for j := len(ip) - 1; j >= 0; j-- {
			ip[j]++
			if ip[j] > 0 {
				break
			}
		}
	}
	isNetworkAddr := func(ip net.IP, ipnet net.IPNet) bool {
		return ip.Equal(ipnet.IP)
	}
	ip, ipnet, err := net.ParseCIDR(ipStr)
	if err != nil {
		return nil, nil, err
	}
	log.Debug("Checking if IP %+v is a network address or an unique IP", ip)
	if isNetworkAddr(ip, *ipnet) {
		log.Debug("Is a network addr")
		incIP(ip)
		log.Debug("Incremented ip %+v", ip)
		for ipnet.Contains(ip) {
			if err = dbConn.PutIP(ip); err != nil {
				log.Debug("Error %s", err)
				incIP(ip)
			} else {
				return ip, ipnet, nil
			}
		}
		return nil, nil, fmt.Errorf("reached maximum IPs used")
	} else {
		log.Debug("Is an unique IP")
		if err = dbConn.PutIP(ip); err == nil {
			return ip, ipnet, nil
		} else {
			return nil, nil, err
		}
	}
}

// This way it's easier to mock this func on tests.
var execShCommand = execShCmd

// execShCmd executes the given strCmd on the with the help of /bin/sh.
func execShCmd(strCmd string) ([]byte, error) {
	log.Debug("Executing %+v", strCmd)
	cmd := exec.Command("/bin/sh", "-c", strCmd)

	stdoutpipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Error("Error stdout: %s. for command: %s", err, strCmd)
		return nil, err
	}
	stderrpipe, err := cmd.StderrPipe()
	if err != nil {
		log.Error("Error stderr: %s. for command: %s", err, strCmd)
		return nil, err
	}
	err = cmd.Start()
	if err != nil {
		log.Error("Error: %s. for command: %s", err, strCmd)
		return nil, err
	}
	stdout, errstderr := ioutil.ReadAll(stdoutpipe)
	stderr, errstdout := ioutil.ReadAll(stderrpipe)

	cmderr := cmd.Wait()

	if errstderr != nil {
		log.Debug("Stdout err: %v", errstderr)
	}
	if errstdout != nil {
		log.Debug("Stderr err: %v", errstdout)
	}
	log.Debug("Stdout is: '%s'\n", stdout)
	log.Debug("Stderr is: '%s'\n", stderr)
	if cmderr != nil {
		log.Error("cmderr: %v, %v", cmderr, string(stderr))
	}
	return stdout, cmderr
}
