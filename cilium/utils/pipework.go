package utils

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	ucdb "github.com/cilium-team/cilium/cilium/utils/comm/db"
	upsi "github.com/cilium-team/cilium/cilium/utils/profile/subpolicies/intent"

	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/op/go-logging"
)

var log = logging.MustGetLogger("cilium")

// CreateBridge creates an OVS bridge on the host with the help of the pipework
// utility. Pipework full path should be set under 'PIPEWORK' environment
// variable. Returns the interface name used inside the container with the ID
// containerID and the MAC address of the bridge attached to it.
func CreateBridge(ip net.IP, ipnet net.IPNet, netConf upsi.NetConf, containerPID int, containerID string) (string, string, error) {
	log.Debug("")
	ones, _ := ipnet.Mask.Size()
	br := "lxc-br0"
	//Run magical script IFNAME=$($PIPEWORK $BRNAME $NAME ${PREFIX}@$GW $MAC)
	pipeworkCmd := fmt.Sprintf("%s --quiet %s %d %s %s/%d", os.Getenv("PIPEWORK"), br, containerPID, containerID, ip, ones)
	if *netConf.Gw != "" {
		pipeworkCmd += "@" + *netConf.Gw
	}
	pipeworkCmd += fmt.Sprintf(" %d %d %d", *netConf.Group, *netConf.BD, *netConf.Namespace)
	if *netConf.MAC != "" {
		pipeworkCmd += " " + *netConf.MAC
	} else {
		pipeworkCmd += " auto"
	}
	if *netConf.Route != "" {
		pipeworkCmd += " '" + *netConf.Route + "'"
	}
	out, err := execShCommand(pipeworkCmd)
	if err != nil {
		return "", "", err
	}
	ret := strings.Split(string(out), " ")
	return ret[0], ret[1], nil
}

// AddEndpoint adds a local endpoint for the remote container with the given
// container ID value.
func AddEndpoint(dbConn ucdb.Db, containerID string) error {
	log.Debug("Adding remote endpoint %s, local node %s", containerID, os.Getenv("HOST_IP"))
	attempts := 1
	for attempts <= 10 {
		log.Debug("Attempt %d for container %v...", attempts, containerID)
		endpoint, err := dbConn.GetEndpoint(containerID)
		if err != nil || endpoint.Container == "" {
			log.Debug("Could not find entry for %s", containerID)
			const delay = 1 * time.Second
			time.Sleep(delay)
			attempts++
			continue
		}

		log.Debug("Found endpoint %+v", endpoint)

		if endpoint.Node != os.Getenv("HOST_IP") {
			for i := 0; i < len(endpoint.IPs); i++ {
				ip := endpoint.IPs[i].String()
				mac := endpoint.MACs[i]
				addEndpointCmd := fmt.Sprintf(
					"%s %s %d %d %d %s %s %s",
					os.Getenv("ADD_ENDPOINT"),
					endpoint.Node,
					endpoint.Group,
					endpoint.BD,
					endpoint.Namespace,
					ip,
					mac,
					containerID)
				if _, err := execShCommand(addEndpointCmd); err != nil {
					log.Debug("%+v", err)
					return err
				}
			}
		}
		return nil
	}
	return fmt.Errorf("Remote endpoint not found for container: '%v'", containerID)
}

// RemoveLocalEndpoint removes the local endpoint for the remote container with
// the given container ID value.
func RemoveLocalEndpoint(dbConn ucdb.Db, containerID string) error {
	log.Debug("")
	attempts := 1
	removeEndpointCmd := fmt.Sprintf("%s %s", os.Getenv("REMOVE_ENDPOINT"), containerID)
	for attempts <= 10 {
		log.Debug("Attempt %d for container %v...", attempts, containerID)
		endpoint, err := dbConn.GetEndpoint(containerID)
		if err != nil || endpoint.Container == "" {
			log.Debug("Could not find entry for %s", containerID)
			const delay = 1 * time.Second
			time.Sleep(delay)
			attempts++
			continue
		}
		log.Debug("Found endpoint: %+v", endpoint)
		if endpoint.Node == os.Getenv("HOST_IP") {
			removeEndpointCmd += " " + endpoint.Interface
		}
		break
	}
	if _, err := execShCommand(removeEndpointCmd); err != nil {
		log.Debug("Error: %+v", err)
		return err
	}
	return nil
}

// RemoveEndpoint removes endpoint for the container with the given container ID
// value.
func RemoveEndpoint(containerID string) error {
	log.Debug("")
	RemoveEndpointCmd := fmt.Sprintf("%s %s", os.Getenv("REMOVE_ENDPOINT"), containerID)
	if _, err := execShCommand(RemoveEndpointCmd); err != nil {
		log.Debug("Error: %+v", err)
		return err
	}
	return nil
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
