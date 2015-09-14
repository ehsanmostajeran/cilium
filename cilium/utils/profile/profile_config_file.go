package profile

import (
	upsd "github.com/cilium-team/cilium/cilium/utils/profile/subpolicies/docker"
	upsi "github.com/cilium-team/cilium/cilium/utils/profile/subpolicies/intent"
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
	Name         string            `json:"name,omitempty" yaml:"name,omitempty"`
	Owner        string            `json:"owner,omitempty" yaml:"owner,omitempty"`
	Coverage     Coverage          `json:"coverage,omitempty" yaml:"coverage,omitempty"`
	DockerConfig upsd.DockerConfig `json:"docker-config,omitempty" yaml:"docker-config,omitempty"`
	IntentConfig upsi.IntentConfig `json:"intent-config,omitempty" yaml:"intent-config,omitempty"`
}

// FilterPoliciesByUser returns a slice of PolicySource where the Owner value is
// the same as the name of the given user.
func FilterPoliciesByUser(policies []PolicySource, user User) []PolicySource {
	filteredPolicy := make([]PolicySource, 0, len(policies))
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
		p.IntentConfig.Config.NetPolicy.OVSConfig.Rules = new([]string)
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
