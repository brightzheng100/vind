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
package cluster

import (
	"testing"

	"github.com/docker/docker/api/types/network"
	"github.com/stretchr/testify/assert"
)

func TestNewRuntimeNetworks(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		networks := map[string]*network.EndpointSettings{}
		networks["mynetwork"] = &network.EndpointSettings{
			Gateway:     "172.17.0.1",
			IPAddress:   "172.17.0.4",
			IPPrefixLen: 16,
		}
		res := NewRuntimeNetworks(networks)

		expectedRuntimeNetworks := []*RuntimeNetwork{
			&RuntimeNetwork{Name: "mynetwork", Gateway: "172.17.0.1", IP: "172.17.0.4", Mask: "255.255.0.0"}}
		assert.Equal(t, expectedRuntimeNetworks, res)
	})
}
