package intent

import (
	"encoding/json"
	"reflect"
	"testing"
)

var wantintentconfiggostr = `IntentConfig.Priority: 500, IntentConfig.Config ` +
	`Intent.AddArguments: 'foo', 'bar', Intent.AddToDNS: true, Intent.HostNameIs ` +
	`HostNameType.Label: ^com\.intent\.logical-name$, Intent.LoadBalancer ` +
	`LoadBalancer.Name: web, LoadBalancer.TrafficType: http, LoadBalancer.BindPort: ` +
	`80, Intent.MaxScale: 4, Intent.NetConf: NetConf.Br: lxc-br0, NetConf.CIDR: ` +
	`1.1.0.0/25, NetConf.MAC: 00:01:02:03:04:05, NetConf.Gw: 1.1.0.126, ` +
	`NetConf.Route: 192.168.50.0/24 via 172.17.42.1, NetConf.Group: 3, NetConf.BD: 5, ` +
	`NetConf.Namespace: 9, Intent.NetPolicy: NetPolicy.OVSConfig: OVSConfig.ConfigFiles: ` +
	`'operator-ovs-intent-web-service.yml', 'operator-ovs-intent-dns.yml', ` +
	`OVSConfig.Rules: (nil), Intent.RemoveDockerLinks: true, Intent.RemovePortBindings: ` +
	`false, Intent.ServiceKeyIs ServiceKeyType.Label: ^com\.intent\.logical-name$`

func TestIntentConfigGoString(t *testing.T) {
	var i Intent
	if err := json.Unmarshal([]byte(intentjson), &i); err != nil {
		t.Fatalf("error while unmarshalling intentjson: %s", err)
	}
	ic := IntentConfig{Config: i, Priority: 500}
	gotIntentConfigStr := ic.GoString()
	if gotIntentConfigStr != wantintentconfiggostr {
		t.Errorf("invalid intent config gotten:\ngot  %s\nwant %s\n", gotIntentConfigStr, wantintentconfiggostr)
	}
}

func TestIntentConfigDeepCopy(t *testing.T) {
	var i Intent
	if err := json.Unmarshal([]byte(intentjson), &i); err != nil {
		t.Fatalf("error while unmarshalling intentjson: %s", err)
	}
	ic := IntentConfig{Config: i, Priority: 500}
	igot := ic.DeepCopy()
	if !reflect.DeepEqual(igot, ic) {
		t.Errorf("both intent config should have the same values:\ngot  %#v\n want%#v", igot, ic)
	}
	if &igot.Priority == &ic.Priority {
		t.Errorf("priority should not have the same pointers")
	}
	if &igot.Config.AddToDNS == &ic.Config.AddToDNS {
		t.Errorf("config.AddToDNS should not have the same pointers")
	}
}

func TestOrderIntentConfigsByAscendingPriority(t *testing.T) {
	var i Intent
	if err := json.Unmarshal([]byte(intentjson), &i); err != nil {
		t.Fatalf("error while unmarshalling intentjson: %s", err)
	}
	ics := []IntentConfig{
		IntentConfig{
			Priority: 223,
			Config:   i,
		},
		IntentConfig{
			Priority: 2,
			Config:   i,
		},
		IntentConfig{
			Priority: 1,
			Config:   i,
		},
		IntentConfig{
			Priority: 199,
			Config:   i,
		},
	}
	want := []IntentConfig{
		IntentConfig{
			Priority: 1,
			Config:   i,
		},
		IntentConfig{
			Priority: 2,
			Config:   i,
		},
		IntentConfig{
			Priority: 199,
			Config:   i,
		},
		IntentConfig{
			Priority: 223,
			Config:   i,
		},
	}
	OrderIntentConfigsByAscendingPriority(ics)
	for i := 0; i < len(ics); i++ {
		if ics[i].Priority != want[i].Priority {
			t.Errorf("IntentConfigs are blady sorted (Priority):\ngot  %d\nwant %d", ics[i].Priority, want[i].Priority)
		}
	}
}
