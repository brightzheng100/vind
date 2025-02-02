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
	"github.com/brightzheng100/vind/pkg/utils"
)

func runWithLogging(cmd exec.Cmd) error {
	output, err := exec.CombinedOutputLines(cmd)
	if err != nil {
		// log error output if there was any
		for _, line := range output {
			utils.Logger.Error(line)
		}
	}
	return err

}

// Start starts a container.
func Start(container string) error {
	cmd := exec.Command("docker", "start", container)
	return runWithLogging(cmd)
}
