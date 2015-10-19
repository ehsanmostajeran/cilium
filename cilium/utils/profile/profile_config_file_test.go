package profile

import (
	//	"io/ioutil"
	"reflect"
	//	"sort"
	"testing"

	upsd "github.com/cilium-team/cilium/cilium/utils/profile/subpolicies/docker"
	upsi "github.com/cilium-team/cilium/cilium/utils/profile/subpolicies/intent"
	upsk "github.com/cilium-team/cilium/cilium/utils/profile/subpolicies/kubernetes"
)

func TestFilterPoliciesByUser(t *testing.T) {
	u1 := User{Name: "root", ID: 0}
	rootPolicy := PolicySource{
		Owner: "root",
		Policies: []Policy{
			Policy{
				Name:         "something",
				Owner:        "root",
				Coverage:     Coverage{Labels: map[string]string{"com.compose.dev": "foo"}},
				DockerConfig: upsd.DockerConfig{},
				IntentConfig: upsi.IntentConfig{},
			},
		},
	}
	usr1Policy := PolicySource{
		Owner: "usr1",
		Policies: []Policy{
			Policy{
				Name:         "something2",
				Owner:        "usr2",
				Coverage:     Coverage{Labels: map[string]string{"com.compose.dev": "foo"}},
				DockerConfig: upsd.DockerConfig{},
				IntentConfig: upsi.IntentConfig{},
			},
		},
	}
	filteredPlcy := FilterPoliciesByUser([]PolicySource{rootPolicy, usr1Policy}, u1)
	if len(filteredPlcy) != 1 {
		t.Errorf("invalid number of filtered policies:\ngot  %d\nwant %d", filteredPlcy, 1)
	} else {
		if !reflect.DeepEqual(filteredPlcy[0], rootPolicy) {
			t.Errorf("invalid filtered policies:\ngot  %+v\nwant %+v", filteredPlcy[0], rootPolicy)
		}
	}
}

func TestReadOVSConfigFiles(t *testing.T) {
	rootPolicy := PolicySource{
		Owner: "root",
		Policies: []Policy{
			Policy{
				Name:         "something",
				Owner:        "root",
				Coverage:     Coverage{Labels: map[string]string{"com.compose.dev": "foo"}},
				DockerConfig: upsd.DockerConfig{},
				IntentConfig: upsi.IntentConfig{
					Config: upsi.Intent{
						NetPolicy: upsi.NetPolicy{
							OVSConfig: upsi.OVSConfig{
								ConfigFiles: new([]string),
							},
						},
					},
				},
			},
		},
	}
	rootPolicyWant := PolicySource{
		Owner: "root",
		Policies: []Policy{
			Policy{
				Name:         "something",
				Owner:        "root",
				Coverage:     Coverage{Labels: map[string]string{"com.compose.dev": "foo"}},
				DockerConfig: upsd.DockerConfig{},
				IntentConfig: upsi.IntentConfig{
					Config: upsi.Intent{
						NetPolicy: upsi.NetPolicy{
							OVSConfig: upsi.OVSConfig{
								ConfigFiles: new([]string),
								Rules:       new([]string),
							},
						},
					},
				},
			},
		},
	}
	usr1Policy := PolicySource{
		Owner: "usr1",
		Policies: []Policy{
			Policy{
				Name:         "something2",
				Owner:        "usr2",
				Coverage:     Coverage{Labels: map[string]string{"com.compose.dev": "foo"}},
				DockerConfig: upsd.DockerConfig{},
				IntentConfig: upsi.IntentConfig{
					Config: upsi.Intent{
						NetPolicy: upsi.NetPolicy{
							OVSConfig: upsi.OVSConfig{
								ConfigFiles: new([]string),
								Rules:       new([]string),
							},
						},
					},
				},
			},
		},
	}
	usr1PolicyWant := PolicySource{
		Owner: "usr1",
		Policies: []Policy{
			Policy{
				Name:         "something2",
				Owner:        "usr2",
				Coverage:     Coverage{Labels: map[string]string{"com.compose.dev": "foo"}},
				DockerConfig: upsd.DockerConfig{},
				IntentConfig: upsi.IntentConfig{
					Config: upsi.Intent{
						NetPolicy: upsi.NetPolicy{
							OVSConfig: upsi.OVSConfig{
								ConfigFiles: new([]string),
								Rules:       new([]string),
							},
						},
					},
				},
			},
		},
	}
	*rootPolicy.Policies[0].IntentConfig.Config.NetPolicy.OVSConfig.ConfigFiles =
		append(*rootPolicy.Policies[0].IntentConfig.Config.NetPolicy.OVSConfig.ConfigFiles,
			`ovs-rules.yml`)
	*rootPolicyWant.Policies[0].IntentConfig.Config.NetPolicy.OVSConfig.Rules =
		append(*rootPolicyWant.Policies[0].IntentConfig.Config.NetPolicy.OVSConfig.Rules,
			`priority=100,ip,nw_src=1.1.0.252,actions=NORMAL`,
			`priority=100,ip,nw_dst=1.1.0.252,actions=NORMAL`,
		)
	*rootPolicyWant.Policies[0].IntentConfig.Config.NetPolicy.OVSConfig.ConfigFiles =
		append(*rootPolicyWant.Policies[0].IntentConfig.Config.NetPolicy.OVSConfig.ConfigFiles,
			`ovs-rules.yml`)
	*usr1Policy.Policies[0].IntentConfig.Config.NetPolicy.OVSConfig.Rules =
		append(*usr1Policy.Policies[0].IntentConfig.Config.NetPolicy.OVSConfig.Rules,
			`priority=70,ip,nw_dst=1.1.0.128/26,actions=NORMAL`,
			`priority=70,ip,nw_src=1.1.0.128/26,actions=NORMAL`)
	*usr1Policy.Policies[0].IntentConfig.Config.NetPolicy.OVSConfig.ConfigFiles =
		append(*usr1Policy.Policies[0].IntentConfig.Config.NetPolicy.OVSConfig.ConfigFiles,
			`ovs-rules.yml`)
	*usr1PolicyWant.Policies[0].IntentConfig.Config.NetPolicy.OVSConfig.Rules =
		append(*usr1PolicyWant.Policies[0].IntentConfig.Config.NetPolicy.OVSConfig.Rules,
			`priority=70,ip,nw_dst=1.1.0.128/26,actions=NORMAL`,
			`priority=70,ip,nw_src=1.1.0.128/26,actions=NORMAL`,
			`priority=100,ip,nw_src=1.1.0.252,actions=NORMAL`,
			`priority=100,ip,nw_dst=1.1.0.252,actions=NORMAL`,
		)
	*usr1PolicyWant.Policies[0].IntentConfig.Config.NetPolicy.OVSConfig.ConfigFiles =
		append(*usr1PolicyWant.Policies[0].IntentConfig.Config.NetPolicy.OVSConfig.ConfigFiles,
			`ovs-rules.yml`)
	basePath := `./subpolicies/intent/config_files_test/`
	if err := rootPolicy.Policies[0].ReadOVSConfigFiles(basePath); err != nil {
		t.Errorf("error while reading OVS config files: %s", err)
	}
	if !reflect.DeepEqual(rootPolicy, rootPolicyWant) {
		t.Errorf("invalid rules read:\ngot  %#v\nwant %#v",
			rootPolicy.Policies[0].IntentConfig.Config.NetPolicy.OVSConfig,
			rootPolicyWant.Policies[0].IntentConfig.Config.NetPolicy.OVSConfig)
	}
	if err := usr1Policy.Policies[0].ReadOVSConfigFiles(basePath); err != nil {
		t.Errorf("error while reading OVS config files: %s", err)
	}
	if !reflect.DeepEqual(usr1Policy, usr1PolicyWant) {
		t.Errorf("invalid rules read:\ngot  %#v\nwant %#v",
			usr1Policy.Policies[0].IntentConfig.Config.NetPolicy.OVSConfig,
			usr1PolicyWant.Policies[0].IntentConfig.Config.NetPolicy.OVSConfig)
	}
}

func TestGetDockerConfigs(t *testing.T) {
	dockerCfg1 := upsd.DockerConfig{Priority: 9999}
	dockerCfg2 := upsd.DockerConfig{Priority: 1456}
	rootPolicy := PolicySource{
		Owner: "root",
		Policies: []Policy{
			Policy{
				Name:         "something",
				Owner:        "root",
				Coverage:     Coverage{Labels: map[string]string{"com.compose.dev": "foo"}},
				DockerConfig: dockerCfg1,
				IntentConfig: upsi.IntentConfig{},
			},
		},
	}
	usr1Policy := PolicySource{
		Owner: "usr1",
		Policies: []Policy{
			Policy{
				Name:         "something2",
				Owner:        "usr2",
				Coverage:     Coverage{Labels: map[string]string{"com.compose.dev": "foo"}},
				DockerConfig: dockerCfg2,
				IntentConfig: upsi.IntentConfig{},
			},
		},
	}
	filteredDC := GetDockerConfigs([]PolicySource{rootPolicy, usr1Policy})
	if len(filteredDC) != 2 {
		t.Errorf("invalid number of filtered policies:\ngot  %d\nwant %d", filteredDC, 2)
	} else {
		if filteredDC[0].Priority != dockerCfg1.Priority {
			t.Errorf("invalid filtered policies:\ngot  %+v\nwant %+v", filteredDC[0], dockerCfg1)
		}
		if filteredDC[1].Priority != dockerCfg2.Priority {
			t.Errorf("invalid filtered policies:\ngot  %+v\nwant %+v", filteredDC[1], dockerCfg2)
		}
	}
}

func TestGetIntentConfigs(t *testing.T) {
	intentCfg1 := upsi.IntentConfig{Priority: 9999}
	intentCfg2 := upsi.IntentConfig{Priority: 1456}
	rootPolicy := PolicySource{
		Owner: "root",
		Policies: []Policy{
			Policy{
				Name:         "something",
				Owner:        "root",
				Coverage:     Coverage{Labels: map[string]string{"com.compose.dev": "foo"}},
				DockerConfig: upsd.DockerConfig{},
				IntentConfig: intentCfg1,
			},
		},
	}
	usr1Policy := PolicySource{
		Owner: "usr1",
		Policies: []Policy{
			Policy{
				Name:         "something2",
				Owner:        "usr2",
				Coverage:     Coverage{Labels: map[string]string{"com.compose.dev": "foo"}},
				DockerConfig: upsd.DockerConfig{},
				IntentConfig: intentCfg2,
			},
		},
	}
	filteredIC := GetIntentConfigs([]PolicySource{rootPolicy, usr1Policy})
	if len(filteredIC) != 2 {
		t.Errorf("invalid number of filtered policies:\ngot  %d\nwant %d", filteredIC, 2)
	} else {
		if filteredIC[0].Priority != intentCfg1.Priority {
			t.Errorf("invalid filtered policies:\ngot  %+v\nwant %+v", filteredIC[0], intentCfg1)
		}
		if filteredIC[1].Priority != intentCfg2.Priority {
			t.Errorf("invalid filtered policies:\ngot  %+v\nwant %+v", filteredIC[1], intentCfg2)
		}
	}
}

func TestGetKubernetesConfigs(t *testing.T) {
	kubernetesCfg1 := upsk.KubernetesConfig{Priority: 9999}
	kubernetesCfg2 := upsk.KubernetesConfig{Priority: 1456}
	rootPolicy := PolicySource{
		Owner: "root",
		Policies: []Policy{
			Policy{
				Name:             "something",
				Owner:            "root",
				Coverage:         Coverage{Labels: map[string]string{"com.compose.dev": "foo"}},
				DockerConfig:     upsd.DockerConfig{},
				IntentConfig:     upsi.IntentConfig{},
				KubernetesConfig: kubernetesCfg1,
			},
		},
	}
	usr1Policy := PolicySource{
		Owner: "usr1",
		Policies: []Policy{
			Policy{
				Name:             "something2",
				Owner:            "usr2",
				Coverage:         Coverage{Labels: map[string]string{"com.compose.dev": "foo"}},
				DockerConfig:     upsd.DockerConfig{},
				IntentConfig:     upsi.IntentConfig{},
				KubernetesConfig: kubernetesCfg2,
			},
		},
	}
	filteredKC := GetKubernetesConfigs([]PolicySource{rootPolicy, usr1Policy})
	if len(filteredKC) != 2 {
		t.Errorf("invalid number of filtered policies:\ngot  %d\nwant %d", filteredKC, 2)
	} else {
		if filteredKC[0].Priority != kubernetesCfg1.Priority {
			t.Errorf("invalid filtered policies:\ngot  %+v\nwant %+v", filteredKC[0], kubernetesCfg1)
		}
		if filteredKC[1].Priority != kubernetesCfg2.Priority {
			t.Errorf("invalid filtered policies:\ngot  %+v\nwant %+v", filteredKC[1], kubernetesCfg2)
		}
	}
}

func TestFilterPoliciesByKubernetesKind(t *testing.T) {
	or1 := upsk.ObjectReference{Kind: "fooKind"}
	rootPolicy := PolicySource{
		Owner: "root",
		Policies: []Policy{
			Policy{
				Name:             "something",
				Owner:            "root",
				Coverage:         Coverage{Labels: map[string]string{"com.compose.dev": "foo"}},
				DockerConfig:     upsd.DockerConfig{},
				IntentConfig:     upsi.IntentConfig{},
				KubernetesConfig: upsk.KubernetesConfig{ObjectReference: or1},
			},
		},
	}
	or2 := upsk.ObjectReference{Kind: "barKind"}
	usr1Policy := PolicySource{
		Owner: "usr1",
		Policies: []Policy{
			Policy{
				Name:             "something2",
				Owner:            "usr2",
				Coverage:         Coverage{Labels: map[string]string{"com.compose.dev": "foo"}},
				DockerConfig:     upsd.DockerConfig{},
				KubernetesConfig: upsk.KubernetesConfig{ObjectReference: or2},
			},
		},
	}
	filteredPlcy := FilterPoliciesByKubernetesKind([]PolicySource{rootPolicy, usr1Policy}, "fooKind")
	if len(filteredPlcy) != 1 {
		t.Errorf("invalid number of filtered policies:\ngot  %d\nwant %d", filteredPlcy, 1)
	} else {
		if !reflect.DeepEqual(filteredPlcy[0], rootPolicy) {
			t.Errorf("invalid filtered policies:\ngot  %+v\nwant %+v", filteredPlcy[0], rootPolicy)
		}
	}
}
