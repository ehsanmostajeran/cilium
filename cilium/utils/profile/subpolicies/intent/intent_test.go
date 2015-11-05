package intent

import (
	"encoding/json"
	"reflect"
	"testing"
)

var (
	wantintent = `{"add-arguments":["foo","bar"],"add-to-dns":true,"hostname-is"` +
		`:{"value-of-label":"^com\\.intent\\.logical-name$"},"load-balancer":{"name":` +
		`"web","traffic-type":"http","bind-port":80},"max-scale":4,"net-conf":{"br":` +
		`"lxc-br0","cidr":"1.1.0.0/25","mac":"00:01:02:03:04:05","gw":"1.1.0.126",` +
		`"route":"192.168.50.0/24 via 172.17.42.1","group":3,"bd":5,"namespace":9},` +
		`"net-policy":{"ovs-config":{"ovs-config-files":[` +
		`"operator-ovs-intent-web-service.yml","operator-ovs-intent-dns.yml"]}},` +
		`"remove-docker-links":true,"remove-port-bindings":false,"service-key-is":` +
		`{"label":"^com\\.intent\\.logical-name$"}}`
	wantintentgostr = `Intent.AddArguments: 'foo', 'bar', Intent.AddToDNS: true, ` +
		`Intent.HostNameIs HostNameType.Label: ^com\.intent\.logical-name$, ` +
		`Intent.LoadBalancer LoadBalancer.Name: web, LoadBalancer.TrafficType: ` +
		`http, LoadBalancer.BindPort: 80, Intent.MaxScale: 4, Intent.NetConf: ` +
		`NetConf.Br: lxc-br0, NetConf.CIDR: 1.1.0.0/25, NetConf.MAC: ` +
		`00:01:02:03:04:05, NetConf.Gw: 1.1.0.126, NetConf.Route: ` +
		`192.168.50.0/24 via 172.17.42.1, NetConf.Group: 3, NetConf.BD: 5, ` +
		`NetConf.Namespace: 9, Intent.NetPolicy: NetPolicy.OVSConfig: OVSConfig.ConfigFiles: ` +
		`'operator-ovs-intent-web-service.yml', 'operator-ovs-intent-dns.yml', ` +
		`OVSConfig.Rules: (nil), Intent.RemoveDockerLinks: true, ` +
		`Intent.RemovePortBindings: false, Intent.ServiceKeyIs ServiceKeyType.Label: ` +
		`^com\.intent\.logical-name$`
	intentjson = `{
   "net-conf":{
      "br":"lxc-br0",
      "cidr":"1.1.0.0/25",
      "gw":"1.1.0.126",
      "mac":"00:01:02:03:04:05",
      "route":"192.168.50.0/24 via 172.17.42.1",
      "group":3,
      "bd":5,
      "namespace":9
   },
   "net-policy":{
      "ovs-config":{
         "ovs-config-files":[
            "operator-ovs-intent-web-service.yml",
            "operator-ovs-intent-dns.yml"
         ]
      }
   },
   "max-scale":4,
   "remove-docker-links":true,
   "hostname-is":{
      "value-of-label":"^com\\.intent\\.logical-name$"
   },
   "add-arguments":[
      "foo",
      "bar"
   ],
   "add-to-dns":true,
   "remove-docker-links":true,
   "remove-port-bindings":false,
   "service-key-is":{
      "label":"^com\\.intent\\.logical-name$"
   },
   "load-balancer":{
      "name":"web",
      "traffic-type":"http",
      "bind-port":80
   }
}`
)

func TestIntentGoString(t *testing.T) {
	var i Intent
	if err := json.Unmarshal([]byte(intentjson), &i); err != nil {
		t.Fatalf("error while unmarshalling intentjson: %s", err)
	}
	gotIntentStr := i.GoString()
	if gotIntentStr != wantintentgostr {
		t.Errorf("invalid intent gotten:\ngot  %s\nwant %s\n", gotIntentStr, wantintentgostr)
	}
}

func TestIntentValue(t *testing.T) {
	var i Intent
	if err := json.Unmarshal([]byte(intentjson), &i); err != nil {
		t.Fatalf("error while unmarshalling intentjson: %s", err)
	}
	gotIntent, err := i.Value()
	if err != nil {
		t.Errorf("error while executing Value of Intent: %s", err)
	}
	if gotIntent != wantintent {
		t.Errorf("invalid intent gotten:\ngot  %s\nwant %s\n", gotIntent, wantintent)
	}
}

func TestIntentScan(t *testing.T) {
	var i, iwant Intent
	if err := json.Unmarshal([]byte(intentjson), &iwant); err != nil {
		t.Fatalf("error while unmarshalling intentjson: %s", err)
	}
	if err := i.Scan(intentjson); err != nil {
		t.Errorf("error while executing Value of Intent: %s", err)
	}
	if !reflect.DeepEqual(i, iwant) {
		t.Errorf("invalid intent gotten:\ngot  %s\nwant %s\n", i, iwant)
	}
}

func TestHostNameTypeGoString(t *testing.T) {
	var i Intent
	if err := json.Unmarshal([]byte(intentjson), &i); err != nil {
		t.Fatalf("error while unmarshalling intentjson: %s", err)
	}
	gotHostNameIsStr := i.HostNameIs.GoString()
	wantHostNameIsgostr := `HostNameType.Label: ^com\.intent\.logical-name$`
	if gotHostNameIsStr != wantHostNameIsgostr {
		t.Errorf("invalid HostNameIs gotten:\ngot  %s\nwant %s\n", gotHostNameIsStr, wantHostNameIsgostr)
	}
}

func TestLoadBalancerGoString(t *testing.T) {
	var i Intent
	if err := json.Unmarshal([]byte(intentjson), &i); err != nil {
		t.Fatalf("error while unmarshalling intentjson: %s", err)
	}
	gotLoadBalancerStr := i.LoadBalancer.GoString()
	wantLoadBalancergostr := `LoadBalancer.Name: web, LoadBalancer.TrafficType: http, ` +
		`LoadBalancer.BindPort: 80`
	if gotLoadBalancerStr != wantLoadBalancergostr {
		t.Errorf("invalid LoadBalancer gotten:\ngot  %s\nwant %s\n", gotLoadBalancerStr, wantLoadBalancergostr)
	}
}

func TestNetConfGoString(t *testing.T) {
	var i Intent
	if err := json.Unmarshal([]byte(intentjson), &i); err != nil {
		t.Fatalf("error while unmarshalling intentjson: %s", err)
	}
	gotNetConfStr := i.NetConf.GoString()
	wantNetConfgostr := `NetConf.Br: lxc-br0, NetConf.CIDR: 1.1.0.0/25, ` +
		`NetConf.MAC: 00:01:02:03:04:05, NetConf.Gw: 1.1.0.126, NetConf.Route: ` +
		`192.168.50.0/24 via 172.17.42.1, NetConf.Group: 3, NetConf.BD: 5, NetConf.Namespace: 9`
	if gotNetConfStr != wantNetConfgostr {
		t.Errorf("invalid NetConf gotten:\ngot  %s\nwant %s\n", gotNetConfStr, wantNetConfgostr)
	}
}

func TestNetPolicyGoString(t *testing.T) {
	var i Intent
	if err := json.Unmarshal([]byte(intentjson), &i); err != nil {
		t.Fatalf("error while unmarshalling intentjson: %s", err)
	}
	gotNetPolicyStr := i.NetPolicy.GoString()
	wantNetPolicygostr := `NetPolicy.OVSConfig: OVSConfig.ConfigFiles: ` +
		`'operator-ovs-intent-web-service.yml', 'operator-ovs-intent-dns.yml', OVSConfig.Rules: (nil)`
	if gotNetPolicyStr != wantNetPolicygostr {
		t.Errorf("invalid NetPolicy gotten:\ngot  %s\nwant %s\n", gotNetPolicyStr, wantNetPolicygostr)
	}
}

func TestServiceKeyTypeGoString(t *testing.T) {
	var i Intent
	if err := json.Unmarshal([]byte(intentjson), &i); err != nil {
		t.Fatalf("error while unmarshalling intentjson: %s", err)
	}
	gotServiceKeyIsStr := i.ServiceKeyIs.GoString()
	wantServiceKeyIsgostr := `ServiceKeyType.Label: ^com\.intent\.logical-name$`
	if gotServiceKeyIsStr != wantServiceKeyIsgostr {
		t.Errorf("invalid ServiceKeyIs gotten:\ngot  %s\nwant %s\n", gotServiceKeyIsStr, wantServiceKeyIsgostr)
	}
}

func TestGetHostNameFromLabels(t *testing.T) {
	test := func(hnLbl string, labelsFromUser map[string]string, wanted string) {
		i1 := NewIntent()
		*i1.HostNameIs.Label = hnLbl
		hostname := i1.GetHostNameFromLabels(labelsFromUser)
		if hostname != wanted {
			t.Logf("hnLbl          : %#v\n", hnLbl)
			t.Logf("labelsFromUser : %#v\n", labelsFromUser)
			t.Errorf("got  %#v\nwant %#v", hostname, wanted)
		}
	}
	test(
		`^com\.intent\.virtual-name$`,
		map[string]string{`com.intent.virtual-name`: "web"},
		"web",
	)
	test(
		`^com\.intent\.virtual-name$`,
		map[string]string{`com.intent.virtual-name`: ""},
		"",
	)
	test(
		`^com.intent.v.*$`,
		map[string]string{
			`com.intent.virtual-name`:        "web",
			`com.intent.second-virtual-name`: "redis"},
		"web",
	)
}

func TestgetDefaultOf(t *testing.T) {
	type testStruct struct {
		foo int    `default_value:"1234"`
		bar string `default_value:"something"`
	}
	s := testStruct{foo: 43212, bar: "bar"}
	if got := getDefaultOf(s, "foo"); got != "1234" {
		t.Errorf("invalid default value:\ngot  %s\nwant %s", got, "1234")
	}
	if got := getDefaultOf(s, "bar"); got != "something" {
		t.Errorf("invalid default value:\ngot  %s\nwant %s", got, "something")
	}
}

func TestSetDefaults(t *testing.T) {
	i := Intent{}
	i.SetDefaults()
	if i.AddArguments == nil || len(*i.AddArguments) != 0 {
		t.Errorf("invalid AddArguments:\ngot  %+v\nwant %s", i.AddArguments, "&[]")
	}
	if i.AddToDNS == nil || !*i.AddToDNS {
		t.Errorf("invalid AddToDNS:\ngot  %+v\nwant %t", i.AddToDNS, true)
	}
	if i.HostNameIs.Label == nil || *i.HostNameIs.Label != "" {
		t.Errorf("invalid HostNameIs.Label:\ngot  %+v\nwant %s", i.HostNameIs.Label, "")
	}
	if i.LoadBalancer.BindPort == nil || *i.LoadBalancer.BindPort != 0 {
		t.Errorf("invalid LoadBalancer.BindPort:\ngot  %+v\nwant %d", i.LoadBalancer.BindPort, 0)
	}
	if i.LoadBalancer.Name == nil || *i.LoadBalancer.Name != "ha-proxy" {
		t.Errorf("invalid LoadBalancer.Name:\ngot  %+v\nwant %s", i.LoadBalancer.Name, "ha-proxy")
	}
	if i.LoadBalancer.TrafficType == nil || *i.LoadBalancer.TrafficType != "http" {
		t.Errorf("invalid LoadBalancer.TrafficType:\ngot  %+v\nwant %s", i.LoadBalancer.TrafficType, "http")
	}
	if i.MaxScale == nil || *i.MaxScale != 1 {
		t.Errorf("invalid MaxScale:\ngot  %+v\nwant %d", i.MaxScale, 1)
	}
	if i.NetConf.Br == nil || *i.NetConf.Br != "" {
		t.Errorf("invalid NetConf.Br:\ngot  %+v\nwant %s", i.NetConf.Br, "")
	}
	if i.NetConf.CIDR == nil || *i.NetConf.CIDR != "" {
		t.Errorf("invalid NetConf.CIDR:\ngot  %+v\nwant %s", i.NetConf.CIDR, "")
	}
	if i.NetConf.MAC == nil || *i.NetConf.MAC != "auto" {
		t.Errorf("invalid NetConf.MAC:\ngot  %+v\nwant %s", i.NetConf.MAC, "auto")
	}
	if i.NetConf.Gw == nil || *i.NetConf.Gw != "" {
		t.Errorf("invalid NetConf.Gw:\ngot  %+v\nwant %s", i.NetConf.Gw, "")
	}
	if i.NetConf.Route == nil || *i.NetConf.Route != "" {
		t.Errorf("invalid NetConf.Route:\ngot  %+v\nwant %s", i.NetConf.Route, "")
	}
	if i.NetConf.Group == nil || *i.NetConf.Group != 1 {
		t.Errorf("invalid NetConf.Group:\ngot  %+v\nwant %d", i.NetConf.Group, 1)
	}
	if i.NetConf.BD == nil || *i.NetConf.BD != 1 {
		t.Errorf("invalid NetConf.BD:\ngot  %+v\nwant %d", i.NetConf.BD, 1)
	}
	if i.NetConf.Namespace == nil || *i.NetConf.Namespace != 1 {
		t.Errorf("invalid NetConf.Namespace:\ngot  %+v\nwant %d", i.NetConf.Namespace, 1)
	}
	if i.NetPolicy.OVSConfig.ConfigFiles == nil || len(*i.NetPolicy.OVSConfig.ConfigFiles) != 0 {
		t.Errorf("invalid NetPolicy.OVSConfig.ConfigFiles:\ngot  %+v\nwant %s", i.NetPolicy.OVSConfig.ConfigFiles, "&[]")
	}
	if i.NetPolicy.OVSConfig.Rules == nil || len(*i.NetPolicy.OVSConfig.Rules) != 0 {
		t.Errorf("invalid NetPolicy.OVSConfig.Rules:\ngot  %+v\nwant %s", i.NetPolicy.OVSConfig.Rules, "&[]")
	}
	if i.RemoveDockerLinks == nil || *i.RemoveDockerLinks {
		t.Errorf("invalid RemoveDockerLinks:\ngot  %+v\nwant %t", i.RemoveDockerLinks, false)
	}
	if i.RemovePortBindings == nil || *i.RemovePortBindings {
		t.Errorf("invalid RemovePortBindings:\ngot  %+v\nwant %t", i.RemovePortBindings, false)
	}
	if i.ServiceKeyIs.Label == nil || *i.ServiceKeyIs.Label != "" {
		t.Errorf("invalid ServiceKeyIs.Label:\ngot  %+v\nwant %s", i.ServiceKeyIs.Label, "")
	}
}

func TestMergeWith(t *testing.T) {
	test := func(i1, i2 Intent, wanted bool) {
		if reflect.DeepEqual(i1, i2) != wanted {
			t.Logf("i1: %#v\n", i1)
			t.Logf("i2: %#v\n", i2)
			t.Errorf("got  %#v\nwant %#v", wanted, !wanted)
			panic("Fatal")
		}
	}

	i1 := NewIntent()
	*i1.AddToDNS = true
	*i1.MaxScale = 50
	*i1.HostNameIs.Label = `com\.intent\.label`
	*i1.ServiceKeyIs.Label = `com\.intent\.service-key`
	*i1.NetConf.Br = "lxc-br0"
	*i1.NetConf.CIDR = "1.1.1.2/24"
	*i1.NetConf.Gw = "1.1.1.1/24"
	*i1.RemoveDockerLinks = false
	*i1.NetPolicy.OVSConfig.ConfigFiles = []string{"file1.yml"}
	*i1.NetPolicy.OVSConfig.Rules = []string{"ovs-rule"}

	i2 := NewIntent()
	cpy, err := i1.Value()
	if err != nil {
		t.Errorf("error while running Value on i1: %s", err)
	}
	i2.Scan(cpy)

	test(*i1, *i2, true)

	i2.SetDefaults()
	test(*i1, *i2, false)

	t.Logf("Intent: %#v\n", i2)
	i2.MergeWith(*i1)
	test(*i1, *i2, true)

	i2.SetDefaults()
	test(*i1, *i2, false)

	i3 := NewIntent()
	cpy, err = i1.Value()
	if err != nil {
		t.Errorf("error while running Value on i1: %s", err)
	}
	i3.Scan(cpy)

	t.Logf("Intent: %#v\n", i1)
	i1.SetDefaults()
	test(*i1, *i3, false)

	i1.MergeWith(*i3)
	test(*i1, *i2, false)

	*i1.AddToDNS = false
	*i1.MaxScale = 0
	*i1.HostNameIs.Label = ""
	*i1.ServiceKeyIs.Label = ""
	*i1.NetConf.Br = ""
	*i1.NetConf.CIDR = ""
	*i1.NetConf.Gw = ""
	*i1.RemoveDockerLinks = false
	*i1.NetPolicy.OVSConfig.ConfigFiles = []string{""}
	*i1.NetPolicy.OVSConfig.Rules = []string{""}

	i1.MergeWith(*i3)
	i2.MergeWith(*i1)
	*i3.NetPolicy.OVSConfig.ConfigFiles = append([]string{""}, *i3.NetPolicy.OVSConfig.ConfigFiles...)
	*i3.NetPolicy.OVSConfig.Rules = append([]string{""}, *i3.NetPolicy.OVSConfig.Rules...)
	*i3.MaxScale = 0
	*i3.AddToDNS = false
	test(*i2, *i3, true)

	i1.SetDefaults()
	*i1.HostNameIs.Label = `foo_bar`
	i1.MergeWith(*i3)
	test(*i1, *i3, false)
	*i3.HostNameIs.Label = `foo_bar`
	test(*i1, *i3, true)

	//Default value of MaxScale is 1
	*i2.MaxScale = 1
	i1.MergeWith(*i2)
	test(*i1, *i2, false)

	i2.SetDefaults()
	t.Logf("Intent: %#v\n", i2)
	*i2.MaxScale = 99
	*i1.MaxScale = 1
	//We only merge if self.MaxScale < than other.MaxScale
	i1.MergeWith(*i2)
	test(*i1, *i2, false)

	*i2.MaxScale = 2
	*i1.MaxScale = 99
	i1.MergeWith(*i2)
	test(*i1, *i2, false)
	i2.MergeWithOverwrite(*i3)
	*i2.MaxScale = 2
	test(*i1, *i2, true)

	*i1.NetConf.Gw = ""
	*i1.NetConf.CIDR = "1.1.1.2/24"
	*i1.NetConf.Br = ""
	*i2.NetConf.Gw = "1.1.1.1/24"
	*i2.NetConf.CIDR = ""
	*i2.NetConf.Br = "lxc-br0"
	i1.MergeWith(*i2)
	*i2.NetConf.CIDR = "1.1.1.2/24"
	test(*i1, *i2, true)

	*i1.NetConf.Gw = "AAAAAAAA"
	*i1.NetConf.CIDR = "1.1.1.2/24"
	*i1.NetConf.Br = "BBBBBBB"
	*i2.NetConf.Gw = "1.1.1.1/24"
	*i2.NetConf.CIDR = "0"
	*i2.NetConf.Br = "lxc-br0"
	i1.MergeWith(*i2)
	test(*i1, *i2, false)

	i2.SetDefaults()
	i3.SetDefaults()
	*i2.NetPolicy.OVSConfig.ConfigFiles = []string{"a"}
	*i2.NetPolicy.OVSConfig.Rules = []string{"b"}
	*i3.NetPolicy.OVSConfig.ConfigFiles = []string{"c"}
	*i3.NetPolicy.OVSConfig.Rules = []string{"d"}
	i2.MergeWith(*i3)
	test(*i2, *i3, false)
}

func TestMergeWithOverwrite(t *testing.T) {
	test := func(i1, i2 Intent, wanted bool) {
		if reflect.DeepEqual(i1, i2) != wanted {
			t.Logf("i1: %#v\n", i1)
			t.Logf("i2: %#v\n", i2)
			t.Errorf("got  %#v\nwant %#v", wanted, !wanted)
		}
	}

	i1 := NewIntent()
	*i1.AddToDNS = true
	*i1.MaxScale = 50
	*i1.HostNameIs.Label = `com\.intent\.label`
	*i1.ServiceKeyIs.Label = `com\.intent\.service-key`
	*i1.NetConf.Br = "lxc-br0"
	*i1.NetConf.CIDR = "1.1.1.2/24"
	*i1.NetConf.Gw = "1.1.1.1/24"
	*i1.RemoveDockerLinks = false
	*i1.NetPolicy.OVSConfig.ConfigFiles = []string{"file1.yml"}
	*i1.NetPolicy.OVSConfig.Rules = []string{"ovs-rule"}

	i2 := NewIntent()
	cpy, err := i1.Value()
	if err != nil {
		t.Errorf("error while running Value on i1: %s", err)
	}
	i2.Scan(cpy)
	test(*i1, *i2, true)
	i2.SetDefaults()
	test(*i1, *i2, false)

	i2.MergeWithOverwrite(*i1)
	test(*i1, *i2, true)

	i2.SetDefaults()
	test(*i1, *i2, false)

	i3 := NewIntent()
	cpy, err = i1.Value()
	if err != nil {
		t.Errorf("error while running Value on i1: %s", err)
	}
	i3.Scan(cpy)
	t.Logf("Intent: %#v\n", i1)
	i1.SetDefaults()
	test(*i1, *i3, false)

	i1.MergeWithOverwrite(*i3)
	test(*i1, *i2, false)

	*i1.AddToDNS = false
	*i1.MaxScale = 3
	*i1.HostNameIs.Label = ""
	*i1.ServiceKeyIs.Label = ""
	*i1.NetConf.Gw = ""
	*i1.NetConf.CIDR = ""
	*i1.NetConf.Br = ""
	*i1.RemoveDockerLinks = false
	*i1.NetPolicy.OVSConfig.ConfigFiles = []string{""}
	*i1.NetPolicy.OVSConfig.Rules = []string{""}

	i1.MergeWithOverwrite(*i3)
	i2.MergeWithOverwrite(*i1)
	t.Logf("i1: %#v\n", i1)
	t.Logf("i2: %#v\n", i2)
	t.Logf("i3: %#v\n", i3)
	test(*i2, *i3, false)
	*i3.AddToDNS = false
	test(*i1, *i3, true)
	test(*i2, *i3, true)

	i1.SetDefaults()
	*i1.HostNameIs.Label = `foo_bar`
	i1.MergeWithOverwrite(*i3)
	test(*i1, *i3, true)

	//Default value of MaxScale is 1
	*i2.MaxScale = 1
	t.Logf("i1: %#v\n", i1)
	t.Logf("i2: %#v\n", i2)
	i1.MergeWithOverwrite(*i2)
	t.Logf("i1: %#v\n", i1)
	t.Logf("i2: %#v\n", i2)
	test(*i1, *i2, false)

	*i2.MaxScale = 99
	i1.MergeWithOverwrite(*i2)
	test(*i1, *i2, true)

	*i1.NetConf.Gw = ""
	*i1.NetConf.CIDR = "1.1.1.2/24"
	*i1.NetConf.Br = ""
	*i2.NetConf.Gw = "1.1.1.1/24"
	*i2.NetConf.CIDR = ""
	*i2.NetConf.Br = "lxc-br0"
	i1.MergeWithOverwrite(*i2)
	*i2.NetConf.CIDR = "1.1.1.2/24"
	test(*i1, *i2, true)

	*i1.NetConf.Gw = "AAAAAAAA"
	*i1.NetConf.CIDR = "1.1.1.2/24"
	*i1.NetConf.Br = "BBBBBBB"
	*i2.NetConf.Gw = "1.1.1.1/24"
	*i2.NetConf.CIDR = "0"
	*i2.NetConf.Br = "lxc-br0"
	i1.MergeWithOverwrite(*i2)
	test(*i1, *i2, true)

	i2.SetDefaults()
	i3.SetDefaults()
	*i2.NetPolicy.OVSConfig.ConfigFiles = []string{"a"}
	*i2.NetPolicy.OVSConfig.Rules = []string{"b"}
	*i3.NetPolicy.OVSConfig.ConfigFiles = []string{"c"}
	*i3.NetPolicy.OVSConfig.Rules = []string{"d"}
	i2.MergeWithOverwrite(*i3)
	test(*i2, *i3, true)

	var i4 Intent
	if err := json.Unmarshal([]byte(intentjson), &i4); err != nil {
		t.Fatalf("error while unmarshalling intentjson: %s", err)
	}
	i4.NetConf.BD = nil
	i1 = NewIntent()
	i1.OverwriteWith(i4)
	t.Logf("i1 %#v", i1)
	t.Logf("i4 %#v", i4)
	if *i1.NetConf.BD != 1 {
		t.Errorf("NetConf.BD should be equal to its default value:\ngot  %d\nwant %d", *i1.NetConf.BD, 1)
	}
}

func TestIntentOverwriteWith(t *testing.T) {
	var i1, i2, iwant Intent
	if err := json.Unmarshal([]byte(intentjson), &iwant); err != nil {
		t.Fatalf("error while unmarshalling intentjson: %s", err)
	}
	if err := json.Unmarshal([]byte(intentjson), &i1); err != nil {
		t.Fatalf("error while unmarshalling intentjson: %s", err)
	}
	if err := json.Unmarshal([]byte(intentjson), &i2); err != nil {
		t.Fatalf("error while unmarshalling intentjson: %s", err)
	}
	*i2.MaxScale = 9876
	i2.AddArguments = nil
	*iwant.MaxScale = 9876
	if err := i1.OverwriteWith(i2); err != nil {
		t.Errorf("error while executing OverwriteWith of Intent: %s", err)
	}
	if !reflect.DeepEqual(i1, iwant) {
		t.Errorf("invalid host config gotten:\ngot  %#v\nwant %#v\n", i1, iwant)
	}
}

func TestReadOVSConfigFiles(t *testing.T) {
	ovsConfig := OVSConfig{
		ConfigFiles: &[]string{},
		Rules:       &[]string{},
	}
	ovsConfigWant := OVSConfig{
		ConfigFiles: &[]string{},
		Rules:       &[]string{},
	}
	*ovsConfig.Rules = append(*ovsConfig.Rules,
		`priority=70,ip,nw_dst=1.1.0.128/26,actions=NORMAL`,
		`priority=70,ip,nw_src=1.1.0.128/26,actions=NORMAL`)
	*ovsConfigWant.Rules = append(*ovsConfigWant.Rules,
		`priority=70,ip,nw_dst=1.1.0.128/26,actions=NORMAL`,
		`priority=70,ip,nw_src=1.1.0.128/26,actions=NORMAL`,
		`priority=100,ip,nw_src=1.1.0.252,actions=NORMAL`,
		`priority=100,ip,nw_dst=1.1.0.252,actions=NORMAL`)
	*ovsConfig.ConfigFiles = append(*ovsConfig.ConfigFiles,
		`ovs-rules.yml`)
	*ovsConfigWant.ConfigFiles = append(*ovsConfigWant.ConfigFiles,
		`ovs-rules.yml`)
	rules, err := ovsConfig.ReadOVSConfigFiles(`./config_files_test/`)
	if err != nil {
		t.Errorf("error while reading OVS config files: %s", err)
	}
	*ovsConfig.Rules = append(*ovsConfig.Rules, rules...)
	if !reflect.DeepEqual(ovsConfig, ovsConfigWant) {
		t.Errorf("invalid rules read:\ngot  %#v\nwant %#v", ovsConfig, ovsConfigWant)
	}
}
