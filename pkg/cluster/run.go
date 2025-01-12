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
	"bytes"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/brightzheng100/vind/pkg/docker"
	"github.com/brightzheng100/vind/pkg/exec"
)

// run runs a command in host. It will output the combined stdout/error on failure.
func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	output, err := exec.CombinedOutputLines(cmd)
	if err != nil {
		// log error output if there was any
		for _, line := range output {
			log.Error(line)
		}
	}
	return err
}

// Run a command in a container. It will output the combined stdout/error on failure.
func containerRun(nameOrID string, name string, args ...string) error {
	exe := docker.ContainerCmder(nameOrID)
	cmd := exe.Command(name, args...)
	output, err := exec.CombinedOutputLines(cmd)
	if err != nil {
		// log error output if there was any
		for _, line := range output {
			log.WithField("machine", nameOrID).Error(line)
		}
	}
	return err
}

func containerRunShell(nameOrID string, script string) error {
	return containerRun(nameOrID, "/bin/bash", "-c", script)
}

func copy(nameOrID string, content []byte, path string) error {
	buf := bytes.Buffer{}
	buf.WriteString(fmt.Sprintf("cat <<__EOF | tee -a %s\n", path))
	buf.Write(content)
	buf.WriteString("__EOF")
	return containerRunShell(nameOrID, buf.String())
}
