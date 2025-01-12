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
	"strings"

	log "github.com/sirupsen/logrus"
)

// Machine is the machine configuration.
type Machine struct {
	// Name is the machine name.
	//
	// When used in a MachineReplicas object, eg. in vind.yaml config files,
	// this field a format string. This format string needs to have a '%d', which
	// is populated by the machine index, a number between 0 and N-1, N being the
	// Count field of MachineReplicas. Name will default to "node%d"
	//
	// This name will also be used as the machine hostname.
	Name string `json:"name"`
	// Image is the container image to use for this machine.
	Image string `json:"image"`
	// User is the machine user used for SSH login.
	User string `json:"user,omitempty"`
	// Privileged controls whether to start the Machine as a privileged container
	// or not. Defaults to false.
	Privileged bool `json:"privileged,omitempty"`
	// Volumes is the list of volumes attached to this machine.
	Volumes []Volume `json:"volumes,omitempty"`
	// Networks is the list of user-defined docker networks this machine is
	// attached to. These networks have to be created manually before creating the
	// containers via "docker network create mynetwork"
	Networks []string `json:"networks,omitempty"`
	// PortMappings is the list of ports to expose to the host.
	PortMappings []PortMapping `json:"portMappings,omitempty"`
	// Cmd is a cmd which will be run in the container.
	Cmd string `json:"cmd,omitempty"`
	// PublicKey is the name of the public key to upload onto the machine for root
	// SSH access.
	PublicKey string `json:"publicKey,omitempty"`

	// Backend specifies the runtime backend for this machine
	Backend string `json:"backend,omitempty"`
}

// Volume is a volume that can be attached to a Machine.
type Volume struct {
	// Type is the volume type. One of "bind" or "volume".
	Type string `json:"type"`
	// Source is the volume source.
	// With type=bind, the volume source is a directory or a file in the host
	// filesystem.
	// With type=volume, source is either the name of a docker volume or "" for
	// anonymous volumes.
	Source string `json:"source"`
	// Destination is the mount point inside the container.
	Destination string `json:"destination"`
	// ReadOnly specifies if the volume should be read-only or not.
	ReadOnly bool `json:"readOnly"`
}

// PortMapping describes mapping a port from the machine onto the host.
type PortMapping struct {
	// Protocol is the layer 4 protocol for this mapping. One of "tcp" or "udp".
	// Defaults to "tcp".
	Protocol string `json:"protocol,omitempty"`
	// Address is the host address to bind to. Defaults to "0.0.0.0".
	Address string `json:"address,omitempty"`
	// HostPort is the base host port to map the containers ports to. As we
	// configure a number of machine replicas, each machine will use HostPort+i
	// where i is between 0 and N-1, N being the number of machine replicas. If 0,
	// a local port will be automatically allocated.
	HostPort uint16 `json:"hostPort,omitempty"`
	// ContainerPort is the container port to map.
	ContainerPort uint16 `json:"containerPort"`
}

// validate checks basic rules for Machine's fields
func (conf Machine) validate() error {
	validName := strings.Contains(conf.Name, "%d")
	if !validName {
		log.Warnf("Machine conf validation: machine name %v is not valid, it should contains %%d", conf.Name)
		return fmt.Errorf("Machine configuration not valid")
	}
	return nil
}
