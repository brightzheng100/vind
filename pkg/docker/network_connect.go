/*
Copyright 2018 The Kubernetes Authors.
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

package docker

import (
	"github.com/brightzheng100/vind/pkg/exec"
)

// ConnectNetwork connects network to container.
func ConnectNetwork(container, network string) error {
	cmd := exec.Command("docker", "network", "connect", network, container)
	return runWithLogging(cmd)
}

// ConnectNetworkWithAlias connects network to container adding a network-scoped
// alias for the container.
func ConnectNetworkWithAlias(container, network, alias string) error {
	cmd := exec.Command("docker", "network", "connect", network, container, "--alias", alias)
	return runWithLogging(cmd)
}
