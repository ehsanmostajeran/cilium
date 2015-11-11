package profile

import (
	"encoding/json"

	upsd "github.com/cilium-team/cilium/cilium/utils/profile/subpolicies/docker"
	upsi "github.com/cilium-team/cilium/cilium/utils/profile/subpolicies/intent"
	upsk "github.com/cilium-team/cilium/cilium/utils/profile/subpolicies/kubernetes"
)

type ProfileFile struct {
	PolicySource []PolicySource `json:"policy-source,omitempty" yaml:"policy-source,omitempty"`
}

type PolicySource struct {
	Owner    string   `json:"owner,omitempty" yaml:"owner,omitempty"`
	Policies []Policy `json:"policies,omitempty" yaml:"policies,omitempty"`
}

type Policy struct {
	//TODO remove owner redundancy present in this structure
	Name             string                `json:"name,omitempty" yaml:"name,omitempty"`
	Owner            string                `json:"owner,omitempty" yaml:"owner,omitempty"`
	Coverage         Coverage              `json:"coverage,omitempty" yaml:"coverage,omitempty"`
	DockerConfig     upsd.DockerConfig     `json:"docker-config,omitempty" yaml:"docker-config,omitempty"`
	IntentConfig     upsi.IntentConfig     `json:"intent-config,omitempty" yaml:"intent-config,omitempty"`
	KubernetesConfig upsk.KubernetesConfig `json:"kubernetes-config,omitempty" yaml:"kubernetes-config,omitempty"`
}

// Value marshals the receiver Policy into a json string.
func (p Policy) Value() (string, error) {
	if data, err := json.Marshal(p); err != nil {
		return "", err
	} else {
		return string(data), nil
	}
}

// Scan unmarshals the input into the receiver Policy.
func (p *Policy) Scan(input string) error {
	return json.Unmarshal([]byte(input), p)
}

// FilterPoliciesByUser returns a slice of PolicySource where the Owner value is
// the same as the name of the given user.
func FilterPoliciesByUser(policies []PolicySource, user User) []PolicySource {
	filteredPolicy := []PolicySource{}
	for _, policy := range policies {
		if policy.Owner == user.Name {
			filteredPolicy = append(filteredPolicy, policy)
		}
	}
	return filteredPolicy
}

// ReadOVSConfigFiles reads all OVSConfigFiles under IntentConfig and appends
// those rules read under OVSConfigRules.
func (p *Policy) ReadOVSConfigFiles(basePath string) error {
	rules, err := p.IntentConfig.Config.NetPolicy.OVSConfig.ReadOVSConfigFiles(basePath)
	if err != nil {
		return err
	}
	log.Debug("Rules read: %+v\n", rules)
	if p.IntentConfig.Config.NetPolicy.OVSConfig.Rules == nil {
		p.IntentConfig.Config.NetPolicy.OVSConfig.Rules = &[]string{}
	}
	*p.IntentConfig.Config.NetPolicy.OVSConfig.Rules =
		append(*p.IntentConfig.Config.NetPolicy.OVSConfig.Rules, rules...)
	return nil
}

// GetDockerConfigs returns all DockerConfig from the given slice of
// PolicySource.
func GetDockerConfigs(policiesSource []PolicySource) []upsd.DockerConfig {
	dockerConfigs := make([]upsd.DockerConfig, 0, len(policiesSource))
	for _, policySource := range policiesSource {
		for _, policy := range policySource.Policies {
			dockerConfigs = append(dockerConfigs, policy.DockerConfig)
		}
	}
	return dockerConfigs
}

// GetIntentConfigs returns all IntentConfig from the given slice of
// PolicySource.
func GetIntentConfigs(policiesSource []PolicySource) []upsi.IntentConfig {
	intentConfigs := make([]upsi.IntentConfig, 0, len(policiesSource))
	for _, policySource := range policiesSource {
		for _, policy := range policySource.Policies {
			intentConfigs = append(intentConfigs, policy.IntentConfig)
		}
	}
	return intentConfigs
}

// GetKubernetesConfigs returns all KubernetesConfig from the given slice of
// PolicySource.
func GetKubernetesConfigs(policiesSource []PolicySource) []upsk.KubernetesConfig {
	kubernetesConfigs := make([]upsk.KubernetesConfig, 0, len(policiesSource))
	for _, policySource := range policiesSource {
		for _, policy := range policySource.Policies {
			kubernetesConfigs = append(kubernetesConfigs, policy.KubernetesConfig)
		}
	}
	return kubernetesConfigs
}

// FilterPoliciesByKubernetesKind returns a slice of PolicySource where the Kind
// value is the same as the given Kind.
func FilterPoliciesByKubernetesKind(policies []PolicySource, kind string) []PolicySource {
	filteredOwnerPolicy := map[string][]Policy{}
	for _, policy := range policies {
		for _, policyOwner := range policy.Policies {
			if policyOwner.KubernetesConfig.ObjectReference.Kind == kind {
				filteredOwnerPolicy[policy.Owner] = append(filteredOwnerPolicy[policy.Owner], policyOwner)
			}
		}
	}
	filteredPolicy := []PolicySource{}
	for k, v := range filteredOwnerPolicy {
		ps := PolicySource{
			Owner:    k,
			Policies: v,
		}
		filteredPolicy = append(filteredPolicy, ps)
	}
	return filteredPolicy
}
