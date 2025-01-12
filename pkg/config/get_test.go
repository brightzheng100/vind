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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetValueFromConfig(t *testing.T) {
	config := Config{
		Cluster: Cluster{Name: "clustername", PrivateKey: "privatekey"},
		MachineSets: []MachineSet{
			MachineSet{
				Name:     "mySet",
				Replicas: 3,
				Spec: Machine{
					Image:      "myImage",
					Name:       "myName",
					Privileged: true,
				},
			},
		},
	}

	tests := []struct {
		name           string
		stringPath     string
		config         Config
		expectedOutput interface{}
	}{
		{
			"simple path select string",
			"cluster.name",
			Config{
				Cluster:     Cluster{Name: "clustername", PrivateKey: "privatekey"},
				MachineSets: []MachineSet{MachineSet{Name: "mySet", Replicas: 3, Spec: Machine{}}},
			},
			"clustername",
		},
		{
			"array path select global",
			"machines[0].spec",
			config,
			Machine{
				Image:      "myImage",
				Name:       "myName",
				Privileged: true,
			},
		},
		{
			"array path select bool",
			"machines[0].spec.Privileged",
			config,
			true,
		},
	}

	for _, utest := range tests {
		t.Run(utest.name, func(t *testing.T) {
			res, err := GetValueFromConfig(utest.stringPath, utest.config)
			assert.Nil(t, err)
			assert.Equal(t, utest.expectedOutput, res)
		})
	}
}
