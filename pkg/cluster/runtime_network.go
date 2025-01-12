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
	"net"

	"github.com/docker/docker/api/types/network"
)

const (
	ipv4Length = 32
)

// RuntimeNetwork contains information about the network
type RuntimeNetwork struct {
	// Name of the network
	Name string `json:"name,omitempty"`
	// IP of the container
	IP string `json:"ip,omitempty"`
	// Mask of the network
	Mask string `json:"mask,omitempty"`
	// Gateway of the network
	Gateway string `json:"gateway,omitempty"`
}

// NewRuntimeNetworks returns a slice of networks
func NewRuntimeNetworks(networks map[string]*network.EndpointSettings) []*RuntimeNetwork {
	rnList := make([]*RuntimeNetwork, 0, len(networks))
	for key, value := range networks {
		mask := net.CIDRMask(value.IPPrefixLen, ipv4Length)
		maskIP := net.IP(mask).String()
		rnNetwork := &RuntimeNetwork{
			Name:    key,
			IP:      value.IPAddress,
			Mask:    maskIP,
			Gateway: value.Gateway,
		}
		rnList = append(rnList, rnNetwork)
	}
	return rnList
}
