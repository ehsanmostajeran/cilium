package intent

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/cilium-team/mergo"
	"github.com/cilium-team/cilium/Godeps/_workspace/src/github.com/cilium-team/yaml"
)

type Intent struct {
	AddArguments       *[]string      `json:"add-arguments,omitempty" yaml:"add-arguments,omitempty"`
	AddToDNS           *bool          `json:"add-to-dns,omitempty" yaml:"add-to-dns,omitempty" default_value:"true"`
	HostNameIs         HostNameType   `json:"hostname-is" yaml:"hostname-is"`
	LoadBalancer       LoadBalancer   `json:"load-balancer" yaml:"load-balancer"`
	MaxScale           *int           `json:"max-scale,omitempty" yaml:"max-scale,omitempty" default_value:"1"`
	NetConf            NetConf        `json:"net-conf" yaml:"net-conf"`
	NetPolicy          NetPolicy      `json:"net-policy" yaml:"net-policy"`
	RemoveDockerLinks  *bool          `json:"remove-docker-links,omitempty" yaml:"remove-docker-links,omitempty" default_value:"false"`
	RemovePortBindings *bool          `json:"remove-port-bindings,omitempty" yaml:"remove-port-bindings,omitempty" default_value:"false"`
	ServiceKeyIs       ServiceKeyType `json:"service-key-is" yaml:"service-key-is"`
}

// GoString is the implementation of the GoStringer interface so we can easily
// print the intent fields.
func (i Intent) GoString() string {
	var retStr string
	if i.AddArguments != nil {
		retStr += fmt.Sprintf("Intent.AddArguments: '%s', ", strings.Join(*i.AddArguments, "', '"))
	} else {
		retStr += "Intent.AddArguments: (nil), "
	}
	if i.AddToDNS != nil {
		retStr += fmt.Sprintf("Intent.AddToDNS: %t, ", *i.AddToDNS)
	} else {
		retStr += "Intent.AddToDNS: (nil), "
	}
	retStr += fmt.Sprintf("Intent.HostNameIs %#v, ", i.HostNameIs)
	retStr += fmt.Sprintf("Intent.LoadBalancer %#v, ", i.LoadBalancer)
	if i.MaxScale != nil {
		retStr += fmt.Sprintf("Intent.MaxScale: %d, ", *i.MaxScale)
	} else {
		retStr += "Intent.MaxScale: (nil), "
	}
	retStr += fmt.Sprintf("Intent.NetConf: %#v, ", i.NetConf)
	retStr += fmt.Sprintf("Intent.NetPolicy: %#v, ", i.NetPolicy)
	if i.RemoveDockerLinks != nil {
		retStr += fmt.Sprintf("Intent.RemoveDockerLinks: %t, ", *i.RemoveDockerLinks)
	} else {
		retStr += "Intent.RemoveDockerLinks: (nil), "
	}
	if i.RemovePortBindings != nil {
		retStr += fmt.Sprintf("Intent.RemovePortBindings: %t, ", *i.RemovePortBindings)
	} else {
		retStr += "Intent.RemovePortBindings: (nil), "
	}
	retStr += fmt.Sprintf("Intent.ServiceKeyIs %#v", i.ServiceKeyIs)
	return retStr
}

// Value marshals the receiver Intent into a json string.
func (i Intent) Value() (string, error) {
	if data, err := json.Marshal(i); err != nil {
		return "", err
	} else {
		return string(data), err
	}
}

// Scan unmarshals the input into the receiver Intent.
func (i *Intent) Scan(input string) error {
	return json.Unmarshal([]byte(input), i)
}

type HostNameType struct {
	Label *string `json:"value-of-label,omitempty" yaml:"value-of-label,omitempty" default_value:""`
}

// GoString is the implementation of the GoStringer interface so we can easily
// print the HostNameType fields.
func (hnt HostNameType) GoString() string {
	var retStr string
	if hnt.Label != nil {
		retStr += fmt.Sprintf("HostNameType.Label: %s", *hnt.Label)
	} else {
		retStr += "HostNameType.Label: (nil)"
	}
	return retStr
}

type LoadBalancer struct {
	Name        *string `json:"name,omitempty" yaml:"name,omitempty" default_value:"ha-proxy"`
	TrafficType *string `json:"traffic-type,omitempty" yaml:"traffic-type,omitempty" default_value:"http"`
	BindPort    *int    `json:"bind-port,omitempty" yaml:"bind-port,omitempty" default_value:"0"`
}

// GoString is the implementation of the GoStringer interface so we can easily
// print the LoadBalancer fields.
func (lb LoadBalancer) GoString() string {
	var retStr string
	if lb.Name != nil {
		retStr += fmt.Sprintf("LoadBalancer.Name: %s, ", *lb.Name)
	} else {
		retStr += "LoadBalancer.Name: (nil), "
	}
	if lb.TrafficType != nil {
		retStr += fmt.Sprintf("LoadBalancer.TrafficType: %s, ", *lb.TrafficType)
	} else {
		retStr += "LoadBalancer.TrafficType: (nil), "
	}
	if lb.BindPort != nil {
		retStr += fmt.Sprintf("LoadBalancer.BindPort: %d", *lb.BindPort)
	} else {
		retStr += "LoadBalancer.BindPort: (nil)"
	}
	return retStr
}

type NetConf struct {
	Br        *string `json:"br,omitempty" yaml:"br,omitempty" default_value:""`
	CIDR      *string `json:"cidr,omitempty" yaml:"cidr,omitempty" default_value:""`
	MAC       *string `json:"mac,omitempty" yaml:"mac,omitempty" default_value:"auto"`
	Gw        *string `json:"gw,omitempty" yaml:"gw,omitempty" default_value:""`
	Route     *string `json:"route,omitempty" yaml:"route,omitempty" default_value:""`
	Group     *int    `json:"group,omitempty" yaml:"group,omitempty" default_value:"1"`
	BD        *int    `json:"bd,omitempty" yaml:"bd,omitempty" default_value:"1"`
	Namespace *int    `json:"namespace,omitempty" yaml:"namespace,omitempty" default_value:"1"`
}

// GoString is the implementation of the GoStringer interface so we can easily
// print the NetConf fields.
func (nc NetConf) GoString() string {
	var retStr string
	if nc.Br != nil {
		retStr += fmt.Sprintf("NetConf.Br: %s, ", *nc.Br)
	} else {
		retStr += "NetConf.Br: (nil), "
	}
	if nc.CIDR != nil {
		retStr += fmt.Sprintf("NetConf.CIDR: %s, ", *nc.CIDR)
	} else {
		retStr += "NetConf.CIDR: (nil), "
	}
	if nc.MAC != nil {
		retStr += fmt.Sprintf("NetConf.MAC: %s, ", *nc.MAC)
	} else {
		retStr += "NetConf.MAC: (nil), "
	}
	if nc.Gw != nil {
		retStr += fmt.Sprintf("NetConf.Gw: %s, ", *nc.Gw)
	} else {
		retStr += "NetConf.Gw: (nil), "
	}
	if nc.Route != nil {
		retStr += fmt.Sprintf("NetConf.Route: %s, ", *nc.Route)
	} else {
		retStr += "NetConf.Route: (nil), "
	}
	if nc.Group != nil {
		retStr += fmt.Sprintf("NetConf.Group: %d, ", *nc.Group)
	} else {
		retStr += "NetConf.Group: (nil), "
	}
	if nc.BD != nil {
		retStr += fmt.Sprintf("NetConf.BD: %d, ", *nc.BD)
	} else {
		retStr += "NetConf.BD: (nil), "
	}
	if nc.Namespace != nil {
		retStr += fmt.Sprintf("NetConf.Namespace: %d", *nc.Namespace)
	} else {
		retStr += "NetConf.Namespace: (nil)"
	}
	return retStr
}

type NetPolicy struct {
	OVSConfig OVSConfig `json:"ovs-config" yaml:"ovs-config"`
}

// GoString is the implementation of the GoStringer interface so we can easily
// print the NetPolicy fields.
func (np NetPolicy) GoString() string {
	return fmt.Sprintf("NetPolicy.OVSConfig: %#v", np.OVSConfig)
}

type OVSConfig struct {
	ConfigFiles *[]string `json:"ovs-config-files,omitempty" yaml:"ovs-config-files,omitempty"`
	Rules       *[]string `json:"ovs-rules,omitempty" yaml:"ovs-rules,omitempty"`
}

// GoString is the implementation of the GoStringer interface so we can easily
// print the OVSConfig fields.
func (oc OVSConfig) GoString() string {
	var retStr string
	if oc.ConfigFiles != nil {
		retStr += fmt.Sprintf("OVSConfig.ConfigFiles: '%s', ", strings.Join(*oc.ConfigFiles, "', '"))
	} else {
		retStr += "OVSConfig.ConfigFiles: (nil), "
	}
	if oc.Rules != nil {
		retStr += fmt.Sprintf("OVSConfig.Rules: '%s'", strings.Join(*oc.Rules, "', '"))
	} else {
		retStr += "OVSConfig.Rules: (nil)"
	}
	return retStr
}

type ServiceKeyType struct {
	Label *string `json:"label,omitempty" yaml:"label,omitempty" default_value:""`
}

// GoString is the implementation of the GoStringer interface so we can easily
// print the ServiceKeyType fields.
func (sky ServiceKeyType) GoString() string {
	if sky.Label != nil {
		return "ServiceKeyType.Label: " + *sky.Label
	}
	return "ServiceKeyType.Label: (nil)"
}

func NewOVSConfig() *OVSConfig {
	return &OVSConfig{
		ConfigFiles: &[]string{},
		Rules:       &[]string{},
	}
}

// GetHostNameFromLabels returns the HostName set on the given map of labels
// where one of the labels is matched by the regex expression value under
// ValueOfLabel.
func (i *Intent) GetHostNameFromLabels(labels map[string]string) string {
	log.Debug("")
	if *i.HostNameIs.Label != "" {
		for labelKey, labelValue := range labels {
			if match, _ := regexp.MatchString(*i.HostNameIs.Label, labelKey); match {
				return labelValue
			}
		}
	}
	return ""
}

// getDefaultOf returns the field's value of Default's tag.
// For example:
//  foo int `default_value:"1234"`
//  bar string `default_value:"something"`
// will return "1234" for 'foo' field and "something" for 'bar' field.
func getDefaultOf(structure interface{}, field string) string {
	if val, ok := reflect.ValueOf(structure).Type().FieldByName(field); ok {
		return val.Tag.Get(mergo.TagName)
	}
	return ""
}

// SetDefaults sets all receiver's fields into the values set under
// "default_value" tag.
func (i *Intent) SetDefaults() {
	i.AddArguments = &[]string{}
	i.AddToDNS = new(bool)
	if addToDNS, err := strconv.ParseBool(getDefaultOf(*i, "AddToDNS")); err == nil {
		*i.AddToDNS = addToDNS
	}
	i.HostNameIs.Label = new(string)
	*i.HostNameIs.Label = getDefaultOf(i.HostNameIs, "Label")
	i.LoadBalancer.Name = new(string)
	*i.LoadBalancer.Name = getDefaultOf(i.LoadBalancer, "Name")
	i.LoadBalancer.BindPort = new(int)
	if bindPort, err := strconv.ParseInt(getDefaultOf(*i, "BindPort"), 10, 32); err == nil {
		*i.LoadBalancer.BindPort = int(bindPort)
	}
	i.LoadBalancer.TrafficType = new(string)
	*i.LoadBalancer.TrafficType = getDefaultOf(i.LoadBalancer, "TrafficType")
	i.MaxScale = new(int)
	if maxScale, err := strconv.ParseInt(getDefaultOf(*i, "MaxScale"), 10, 32); err == nil {
		*i.MaxScale = int(maxScale)
	}
	i.NetConf.Br = new(string)
	*i.NetConf.Br = getDefaultOf(i.NetConf, "Br")
	i.NetConf.CIDR = new(string)
	*i.NetConf.CIDR = getDefaultOf(i.NetConf, "CIDR")
	i.NetConf.Gw = new(string)
	*i.NetConf.Gw = getDefaultOf(i.NetConf, "Gw")
	i.NetConf.MAC = new(string)
	*i.NetConf.MAC = getDefaultOf(i.NetConf, "MAC")
	i.NetConf.Route = new(string)
	*i.NetConf.Route = getDefaultOf(i.NetConf, "Route")
	i.NetConf.Group = new(int)
	if group, err := strconv.ParseInt(getDefaultOf(i.NetConf, "Group"), 10, 32); err == nil {
		*i.NetConf.Group = int(group)
	}
	i.NetConf.BD = new(int)
	if bd, err := strconv.ParseInt(getDefaultOf(i.NetConf, "BD"), 10, 32); err == nil {
		*i.NetConf.BD = int(bd)
	}
	i.NetConf.Namespace = new(int)
	if namespace, err := strconv.ParseInt(getDefaultOf(i.NetConf, "Namespace"), 10, 32); err == nil {
		*i.NetConf.Namespace = int(namespace)
	}
	i.NetPolicy.OVSConfig.ConfigFiles = &[]string{}
	i.NetPolicy.OVSConfig.Rules = &[]string{}
	i.RemoveDockerLinks = new(bool)
	if removeDockerLinks, err := strconv.ParseBool(getDefaultOf(*i, "RemoveDockerLinks")); err == nil {
		*i.RemoveDockerLinks = removeDockerLinks
	}
	i.RemovePortBindings = new(bool)
	if removePortBindings, err := strconv.ParseBool(getDefaultOf(*i, "RemovePortBindings")); err == nil {
		*i.RemovePortBindings = removePortBindings
	}
	i.ServiceKeyIs.Label = new(string)
	*i.ServiceKeyIs.Label = getDefaultOf(i.ServiceKeyIs, "Label")
}

// MergeWith merges receiver's values with the `other` Intent's values. `other`
// overwrites the receiver's values if its values are equal to default's.
// Special cases:
// MaxScale - will overwrite only if it's lower than receveiver's value and
// higher than default's.
// NetPolicy.OVSConfig.ConfigFiles and NetPolicy.OVSConfig.Rules will be merged
// into the receveiver's ConfigFiles and Rules respectively.
func (i *Intent) MergeWith(other Intent) error {
	if err := mergo.Merge(i, other); err != nil {
		return err
	}
	// We want MaxScale to be lower which means, for example, dev has priority
	// over gov.
	if maxScale, err := strconv.ParseInt(getDefaultOf(*i, "MaxScale"), 10, 32); err == nil && other.MaxScale != nil {
		if *i.MaxScale > *other.MaxScale && *other.MaxScale > int(maxScale) {
			*i.MaxScale = *other.MaxScale
		}
	}
	// We will merge policies from both intents (only if mergo hasn't merge it
	// itself)
	if other.NetPolicy.OVSConfig.ConfigFiles != nil && !reflect.DeepEqual(i.NetPolicy.OVSConfig.ConfigFiles, other.NetPolicy.OVSConfig.ConfigFiles) {
		if i.NetPolicy.OVSConfig.ConfigFiles == nil {
			i.NetPolicy.OVSConfig.ConfigFiles = &[]string{}
		}
		*i.NetPolicy.OVSConfig.ConfigFiles = append(*i.NetPolicy.OVSConfig.ConfigFiles, *other.NetPolicy.OVSConfig.ConfigFiles...)
	}
	if other.NetPolicy.OVSConfig.Rules != nil && !reflect.DeepEqual(i.NetPolicy.OVSConfig.Rules, other.NetPolicy.OVSConfig.Rules) {
		if i.NetPolicy.OVSConfig.Rules == nil {
			i.NetPolicy.OVSConfig.Rules = &[]string{}
		}
		*i.NetPolicy.OVSConfig.Rules = append(*i.NetPolicy.OVSConfig.Rules, *other.NetPolicy.OVSConfig.Rules...)
	}
	return nil
}

// MergeWithOverwrite merges receiver's values with the `other` Intent's values.
// `other` overwrites the receiver's values if the other's values are different
// than default's.
func (i *Intent) MergeWithOverwrite(other Intent) error {
	return mergo.MergeWithOverwrite(i, other)
}

// OverwriteWith overwrites values with the ones from `other` Intent if those
// ones aren't nil.
// Special cases:
// NetPolicy.OVSConfig.ConfigFiles - `others` values will be merged into the
// receiver's ConfigFiles.
// NetPolicy.OVSConfig.Rules - `others` values will be merged into the
// receiver's Rules.
func (i *Intent) OverwriteWith(other Intent) error {
	// Since the merge method that we use using json doesn't merge config and
	// rules like we want, we must save them and then add those rules again.
	oldOVSConfig := NewOVSConfig()
	if i.NetPolicy.OVSConfig.ConfigFiles != nil {
		*oldOVSConfig.ConfigFiles = append(*oldOVSConfig.ConfigFiles, *i.NetPolicy.OVSConfig.ConfigFiles...)
	}
	if i.NetPolicy.OVSConfig.Rules != nil {
		*oldOVSConfig.Rules = append(*oldOVSConfig.Rules, *i.NetPolicy.OVSConfig.Rules...)
	}

	strOther, err := other.Value()
	if err != nil {
		return err
	}
	if err := i.Scan(strOther); err != nil {
		return err
	}

	if other.NetPolicy.OVSConfig.ConfigFiles != nil {
		if i.NetPolicy.OVSConfig.ConfigFiles == nil {
			i.NetPolicy.OVSConfig.ConfigFiles = &[]string{}
		}
		*i.NetPolicy.OVSConfig.ConfigFiles = append(*other.NetPolicy.OVSConfig.ConfigFiles, *oldOVSConfig.ConfigFiles...)
		*i.NetPolicy.OVSConfig.ConfigFiles = removeDuplicates(*i.NetPolicy.OVSConfig.ConfigFiles)
	}
	if other.NetPolicy.OVSConfig.Rules != nil {
		if i.NetPolicy.OVSConfig.Rules == nil {
			i.NetPolicy.OVSConfig.Rules = &[]string{}
		}
		*i.NetPolicy.OVSConfig.Rules = append(*other.NetPolicy.OVSConfig.Rules, *oldOVSConfig.Rules...)
		*i.NetPolicy.OVSConfig.Rules = removeDuplicates(*i.NetPolicy.OVSConfig.Rules)
	}
	return nil
}

// removeDuplicates removes the duplicates of a slice of strings.
func removeDuplicates(s []string) []string {
	result := []string{}
	seen := map[string]bool{}
	for _, val := range s {
		if _, ok := seen[val]; !ok {
			result = append(result, val)
			seen[val] = true
		}
	}
	return result
}

func NewIntent() *Intent {
	i := Intent{}
	i.SetDefaults()
	return &i
}

// ReadOVSConfigFiles reads all OVSConfigFiles and returns the rules read from
// them.
func (i *OVSConfig) ReadOVSConfigFiles(baseDir string) ([]string, error) {
	configs := []string{}
	if i.ConfigFiles != nil {
		for _, file := range *i.ConfigFiles {
			if file, err := os.Stat(filepath.Join(baseDir, file)); err != nil {
				return nil, err
			} else if err := readOVSConfigFromFile(baseDir, file, &[]os.FileInfo{}, &configs); err != nil {
				return nil, err
			}
		}
	}
	return configs, nil
}

// readOVSConfigFromFile is an helper function to ReadOVSConfigFiles so it can
// read all rules from a file if it wasn't read before.
func readOVSConfigFromFile(baseDir string, file os.FileInfo, filesSeen *[]os.FileInfo, rules *[]string) error {
	for _, fileSeen := range *filesSeen {
		if os.SameFile(file, fileSeen) {
			return nil
		}
	}

	data, err := ioutil.ReadFile(filepath.Join(baseDir, file.Name()))
	if err != nil {
		return err
	}
	ovsConfig := NewOVSConfig()

	if err = yaml.Unmarshal(data, &ovsConfig); err != nil {
		return err
	}

	for _, configFile := range *ovsConfig.ConfigFiles {
		if fileStat, err := os.Stat(filepath.Join(baseDir, configFile)); err != nil {
			return err
		} else {
			if err := readOVSConfigFromFile(baseDir, fileStat, filesSeen, rules); err != nil {
				return err
			}
		}
	}
	*rules = append(*rules, *ovsConfig.Rules...)
	*filesSeen = append(*filesSeen, file)

	return nil
}
