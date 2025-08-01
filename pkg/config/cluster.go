/*
Copyright © 2019-2023 footloose developers
Copyright © 2024-2025 Bright Zheng <bright.zheng@outlook.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package config

import (
	"fmt"
	"os"

	"github.com/brightzheng100/vind/pkg/utils"
	"gopkg.in/yaml.v2"
)

// Config is the top level config object.
type Config struct {
	// Cluster describes cluster-wide configuration.
	Cluster Cluster `json:"cluster"`
	// MachineSets describe the sets of machines we define in this cluster.
	MachineSets []MachineSet `json:"machineSets"`
}

// Cluster is a set of Machines.
type Cluster struct {
	// Name is the cluster name. Defaults to "cluster".
	Name string `json:"name"`
	// PrivateKey is the path to the private SSH key used to login into the cluster
	// machines. Can be expanded to user homedir if ~ is found. Ex. ~/.ssh/id_rsa.
	//
	// This field is optional. If absent, machines are expected to have a public
	// key defined.
	PrivateKey string `json:"privateKey,omitempty"`

	// KnownHosts is the path the SSH known_hosts file used to record host keys.
	// Can be expanded to user homedir if ~ is found. Ex. ~/.ssh/known_hosts.
	//
	// This field is optional. If absent, strict host key checking will be disabled.
	KnownHosts string `json:"knownHosts,omitempty"`
}

// MachineSet are a set of machines following the same specification.
type MachineSet struct {
	// Name is the MachineSet's name. Defaults to "test"
	Name string `json:"name"`
	// Replicas is the number of machines within the MachineSet
	Replicas int `json:"replicas"`
	// Spec is the detailed specifications of the machines within the MachineSet
	Spec Machine `json:"spec"`
}

func NewConfigFromYAML(data []byte) (*Config, error) {
	config := Config{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func NewConfigFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return NewConfigFromYAML(data)
}

// validate checks basic rules for MachineReplicas's fields
func (conf MachineSet) validate() error {
	return conf.Spec.validate()
}

// Validate checks basic rules for Config's fields
func (conf Config) Validate() error {
	valid := true
	for _, machine := range conf.MachineSets {
		err := machine.validate()
		if err != nil {
			valid = false
			utils.Logger.Fatalf(err.Error())
		}
	}
	if !valid {
		return fmt.Errorf("Configuration file non valid")
	}
	return nil
}
