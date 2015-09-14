package runnables

import (
	/*
		"io/ioutil"
		"reflect"
		"sort"
	*/
	"testing"

	/*
		"github.com/davecgh/go-spew/spew"
		"gopkg.in/yaml.v2"
	*/)

func TestGetDockerConfig(t *testing.T) {

}

/*
func setupConfigs(files []string, setDefaults bool) ([]DockerConfig, []IntentConfig, error) {
	var (
		data []byte
		err  error
	)

	dockerConfigs := []DockerConfig{}
	intentConfigs := []IntentConfig{}

	for _, file := range files {
		if data, err = ioutil.ReadFile(file); err != nil {
			return nil, nil, err
		}
		pf := *new(ProfileFile)
		pf = ProfileFile{}

		if err = yaml.Unmarshal(data, &pf); err != nil {
			return nil, nil, err
		}

		for _, profile := range pf.PolicySource {
			for _, policy := range profile.Policies {
				log.Debug("policy %+v", policy.Intentconfig.String())
				//policy.Intentconfig.Config.NetPolicy.OVSConfig.ReadOVSConfigFiles(basePath)
				dockerConfigs = append(dockerConfigs, policy.Dockerconfig)
				if setDefaults {
					policy.Intentconfig.Config = *NewIntentFrom(policy.Intentconfig.Config)
				}
				intentConfigs = append(intentConfigs, policy.Intentconfig)
			}
		}
	}
	return dockerConfigs, intentConfigs, nil
}

func TestGetDockerConfig(t *testing.T) {
	testDockerConfig := func(expected, gotten DockerConfig) {
		if !reflect.DeepEqual(expected.Config, gotten.Config) {
			t.Errorf("Coverage: %+v\nExpected: %+v\nGotten  : %+v\n",
				expected.Coverage,
				spew.Sdump(expected.Config),
				spew.Sdump(gotten.Config))
			t.FailNow()
		}
		if !reflect.DeepEqual(expected.HostConfig, gotten.HostConfig) {
			t.Errorf("Coverage: %+v\nExpected: %+v\nGotten  : %+v\n",
				expected.Coverage,
				spew.Sdump(expected.HostConfig),
				spew.Sdump(gotten.HostConfig))
			t.FailNow()
		}
	}
	testIntentConfig := func(expected, gotten IntentConfig) {
		/*log.Debug("Coverage: %+v\nExpected: %+v\nGotten  : %+v\n",
		expected.Coverage,
		spew.Sdump(expected.Config),
		spew.Sdump(gotten.Config))
		* /
		//We have to sort ovs and ovs rules becase deepequal is sort-aware
		if expected.Config.NetPolicy.OVSConfig.ConfigFiles != nil {
			sort.Strings(*expected.Config.NetPolicy.OVSConfig.ConfigFiles)
		}
		if gotten.Config.NetPolicy.OVSConfig.ConfigFiles != nil {
			sort.Strings(*gotten.Config.NetPolicy.OVSConfig.ConfigFiles)
		}
		if expected.Config.NetPolicy.OVSConfig.Rules != nil {
			sort.Strings(*expected.Config.NetPolicy.OVSConfig.Rules)
		}
		if gotten.Config.NetPolicy.OVSConfig.Rules != nil {
			sort.Strings(*gotten.Config.NetPolicy.OVSConfig.Rules)
		}

		if !reflect.DeepEqual(expected.Config, gotten.Config) {
			t.Errorf("Coverage: %+v\nExpected: %+v\nGotten  : %+v\n",
				expected.Coverage,
				spew.Sdump(expected.Config),
				spew.Sdump(gotten.Config))
			t.FailNow()
		}
	}

	var (
		dockerConfigs       []DockerConfig
		intentConfigs       []IntentConfig
		expectDockerConfigs []DockerConfig
		expectIntentConfigs []IntentConfig
		err                 error
	)

	users := []User{
		User{Id: 0, Name: "governator"},
		User{Id: 101, Name: "operator1"},
		User{Id: 102, Name: "operator2"},
	}

	files := []string{
		"config_files/gov1.yml",
		"config_files/gov2.yml",
		"config_files/op1-services.yml",
		"config_files/op2-redis.yml",
		"config_files/op2-web.yml",
	}

	expectingFiles := []string{
		"config_files/expect1.yml",
		"config_files/expect2.yml",
	}

	if dockerConfigs, intentConfigs, err = setupConfigs(files, false); err != nil {
		t.Error(err)
		t.FailNow()
	}

	for _, expectingFile := range expectingFiles {
		if expectDockerConfigs, expectIntentConfigs, err = setupConfigs([]string{expectingFile}, true); err != nil {
			t.Error(err)
			t.FailNow()
		}
		if len(expectDockerConfigs) != 1 || len(expectIntentConfigs) != 1 {
			t.Error("There should be only 1 expected configuration")
			t.FailNow()
		}
		if dockerConfig, isValid := GetDockerConfig(users, dockerConfigs, expectDockerConfigs[0].Coverage.Labels); !isValid {
			t.Error("Docker configuration should be valid")
			t.FailNow()
		} else {
			testDockerConfig(expectDockerConfigs[0], dockerConfig)
		}
		if intentConfig, isValid := GetIntentConfig(users, intentConfigs, expectIntentConfigs[0].Coverage.Labels); !isValid {
			t.Error("Intent configuration should be valid")
			t.FailNow()
		} else {
			testIntentConfig(expectIntentConfigs[0], *intentConfig)
		}
	}
}
*/
