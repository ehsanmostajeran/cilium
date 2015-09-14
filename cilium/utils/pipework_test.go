package utils

import (
	"fmt"
	"net"
	"os"
	"testing"

	up "github.com/cilium-team/cilium/cilium/utils/profile"
	upsi "github.com/cilium-team/cilium/cilium/utils/profile/subpolicies/intent"
)

func TestCreateBridge(t *testing.T) {
	cidr := "10.11.12.13/24"
	br := "lxc-br0"
	mac := "00:01:02:03:04:05"
	gw := "199.231.41.15/24"
	ones := 24
	route := "192.168.50.0/24 via 172.17.42.1"
	group := 4
	bd := 19
	namespace := 20
	pipework := "/bin/pipework"
	ipaddr, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		t.Fatalf("Error parsing IP address: %s", err)
	}
	netConf := upsi.NetConf{
		Br:        &br,
		CIDR:      &cidr,
		MAC:       &mac,
		Gw:        &gw,
		Route:     &route,
		Group:     &group,
		BD:        &bd,
		Namespace: &namespace,
	}
	containerPID := 1999
	containerID := `6b27a943823d0f735346861bbce6e24acdaf435edb259748be556300d1c361f3`
	os.Setenv("PIPEWORK", pipework)
	execShCommand = func(strCmd string) ([]byte, error) {
		strWant := fmt.Sprintf(
			"%s --quiet %s %d %s %s/%d@%s %d %d %d %s '%s'",
			pipework, br, containerPID, containerID, ipaddr, ones, gw, group,
			bd, namespace, mac, route,
		)
		if strCmd != strWant {
			return nil, fmt.Errorf("invalid command:\ngot  %s\nwant %s", strCmd, strWant)
		}
		return []byte("foo " + mac), nil
	}
	ifname, gotmac, err := CreateBridge(ipaddr, *ipnet, netConf, containerPID, containerID)
	if err != nil {
		t.Errorf("error while creating a bridge: %s", err)
	}
	if ifname != "foo" {
		t.Errorf("invalid ifname:\ngot  %s\nwant %s", ifname, "foo")
	}
	if gotmac != mac {
		t.Errorf("invalid mac:\ngot  %s\nwant %s", gotmac, mac)
	}

	//empty mac
	execShCommand = func(strCmd string) ([]byte, error) {
		strWant := fmt.Sprintf(
			"%s --quiet %s %d %s %s/%d@%s %d %d %d auto '%s'",
			pipework, br, containerPID, containerID, ipaddr, ones, gw, group,
			bd, namespace, route,
		)
		if strCmd != strWant {
			return nil, fmt.Errorf("invalid command:\ngot  %s\nwant %s", strCmd, strWant)
		}
		return []byte("foo " + mac), nil
	}
	empty := ""
	netConf.MAC = &empty
	ifname, gotmac, err = CreateBridge(ipaddr, *ipnet, netConf, containerPID, containerID)
	if err != nil {
		t.Errorf("error while creating a bridge: %s", err)
	}
	if ifname != "foo" {
		t.Errorf("invalid ifname:\ngot  %s\nwant %s", ifname, "foo")
	}
	if gotmac != mac {
		t.Errorf("invalid mac:\ngot  %s\nwant %s", gotmac, mac)
	}

	//empty route
	execShCommand = func(strCmd string) ([]byte, error) {
		strWant := fmt.Sprintf(
			"%s --quiet %s %d %s %s/%d@%s %d %d %d auto",
			pipework, br, containerPID, containerID, ipaddr, ones, gw, group,
			bd, namespace,
		)
		if strCmd != strWant {
			return nil, fmt.Errorf("invalid command:\ngot  %s\nwant %s", strCmd, strWant)
		}
		return []byte("foo " + mac), nil
	}
	netConf.Route = &empty
	ifname, gotmac, err = CreateBridge(ipaddr, *ipnet, netConf, containerPID, containerID)
	if err != nil {
		t.Errorf("error while creating a bridge: %s", err)
	}
	if ifname != "foo" {
		t.Errorf("invalid ifname:\ngot  %s\nwant %s", ifname, "foo")
	}
	if gotmac != mac {
		t.Errorf("invalid mac:\ngot  %s\nwant %s", gotmac, mac)
	}
}

func TestAddEndpoint(t *testing.T) {
	containerID := `6b27a943823d0f735346861bbce6e24acdaf435edb259748be556300d1c361f3`
	ips := []net.IP{net.IP{10, 10, 10, 20}, net.IP{10, 10, 10, 30}}
	macs := up.MACs{"00:01:02:03:04:05", "06:07:08:09:0A:0B"}
	endpoint := up.Endpoint{Container: containerID,
		IPs:       ips,
		MACs:      macs,
		Node:      "10.10.10.20",
		Interface: "lxc-br0",
		Group:     4,
		BD:        5,
		Namespace: 6,
		Service:   "web",
	}
	fdb := FakeDB{}
	fdb.OnGetEndpoint = func(cID string) (up.Endpoint, error) {
		if cID != containerID {
			fmt.Errorf("invalid container ID\ngot  %s\nwant %s", cID, containerID)
		}
		return endpoint, nil
	}

	execShCommand = func(strCmd string) ([]byte, error) {
		strWant1 := fmt.Sprintf(
			"%s %s %d %d %d %s %s %s",
			"/opt/backend/add-endpoint.sh",
			endpoint.Node,
			endpoint.Group,
			endpoint.BD,
			endpoint.Namespace,
			ips[0],
			macs[0],
			containerID)
		strWant2 := fmt.Sprintf(
			"%s %s %d %d %d %s %s %s",
			"/opt/backend/add-endpoint.sh",
			endpoint.Node,
			endpoint.Group,
			endpoint.BD,
			endpoint.Namespace,
			ips[1],
			macs[1],
			containerID)
		if strCmd != strWant1 {
			if strCmd != strWant2 {
				return nil, fmt.Errorf("invalid command:\ngot  %s\nwant %s\n or\nwant %s", strCmd, strWant1, strWant2)
			}
		}
		return nil, nil
	}
	os.Setenv("HOST_IP", "10.10.10.20")
	os.Setenv("ADD_ENDPOINT", "/opt/backend/add-endpoint.sh")
	if err := AddEndpoint(fdb, containerID); err != nil {
		t.Errorf("Error while executing AddEndpoint: %s", err)
	}
}

func TestRemoveLocalEndpoint(t *testing.T) {
	containerID := `6b27a943823d0f735346861bbce6e24acdaf435edb259748be556300d1c361f3`
	ips := []net.IP{net.IP{10, 10, 10, 20}, net.IP{10, 10, 10, 30}}
	macs := up.MACs{"00:01:02:03:04:05", "06:07:08:09:0A:0B"}
	endpoint := up.Endpoint{Container: containerID,
		IPs:       ips,
		MACs:      macs,
		Node:      "10.10.10.20",
		Interface: "lxc-br0",
		Group:     4,
		BD:        5,
		Namespace: 6,
		Service:   "web",
	}
	fdb := FakeDB{}
	fdb.OnGetEndpoint = func(cID string) (up.Endpoint, error) {
		if cID != containerID {
			fmt.Errorf("invalid container ID\ngot  %s\nwant %s", cID, containerID)
		}
		return endpoint, nil
	}

	execShCommand = func(strCmd string) ([]byte, error) {
		strWant := fmt.Sprintf(
			"%s %s %s",
			"/opt/backend/remove-endpoint.sh",
			containerID,
			endpoint.Interface,
		)
		if strCmd != strWant {
			return nil, fmt.Errorf("invalid command:\ngot  %s\nwant %s", strCmd, strWant)
		}
		return nil, nil
	}
	os.Setenv("HOST_IP", "10.10.10.20")
	os.Setenv("REMOVE_ENDPOINT", "/opt/backend/remove-endpoint.sh")
	if err := RemoveLocalEndpoint(fdb, containerID); err != nil {
		t.Errorf("Error while executing RemoveLocalEndpoint: %s", err)
	}
}

func TestRemoveEndpoint(t *testing.T) {
	containerID := `6b27a943823d0f735346861bbce6e24acdaf435edb259748be556300d1c361f3`

	execShCommand = func(strCmd string) ([]byte, error) {
		strWant := fmt.Sprintf(
			"%s %s",
			"/opt/backend/remove-endpoint.sh",
			containerID,
		)
		if strCmd != strWant {
			return nil, fmt.Errorf("invalid command:\ngot  %s\nwant %s", strCmd, strWant)
		}
		return nil, nil
	}

	os.Setenv("REMOVE_ENDPOINT", "/opt/backend/remove-endpoint.sh")
	if err := RemoveEndpoint(containerID); err != nil {
		t.Errorf("Error while executing RemoveEndpoint: %s", err)
	}
}
