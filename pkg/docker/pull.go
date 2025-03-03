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
	"os"
	"time"

	"github.com/brightzheng100/vind/pkg/exec"
	"github.com/brightzheng100/vind/pkg/utils"
)

// PullIfNotPresent will pull an image if it is not present locally
// retrying up to retries times
// it returns true if it attempted to pull, and any errors from pulling
func PullIfNotPresent(image string, retries int) (pulled bool, err error) {
	// TODO(bentheelder): switch most (all) of the logging here to debug level
	// once we have configurable log levels
	// if this did not return an error, then the image exists locally
	cmd := exec.Command("docker", "inspect", "--type=image", image)
	if err := cmd.Run(); err == nil {
		utils.Logger.Infof("Docker Image: %s present locally", image)
		return false, nil
	}
	// otherwise try to pull it
	return true, Pull(image, retries)
}

// Pull pulls an image, retrying up to retries times
func Pull(image string, retries int) error {
	utils.Logger.Infof("Pulling image: %s ...", image)
	err := setPullCmd(image).Run()
	// retry pulling up to retries times if necessary
	if err != nil {
		for i := 0; i < retries; i++ {
			time.Sleep(time.Second * time.Duration(i+1))
			utils.Logger.WithError(err).Infof("Trying again to pull image: %s ...", image)
			// TODO(bentheelder): add some backoff / sleep?
			if err = setPullCmd(image).Run(); err == nil {
				break
			}
		}
	}
	if err != nil {
		utils.Logger.WithError(err).Infof("Failed to pull image: %s", image)
	}
	return err
}

// IsRunning checks if Docker is running properly
func IsRunning() error {
	cmd := exec.Command("docker", "version")
	if err := cmd.Run(); err != nil {
		utils.Logger.WithError(err).Infoln("Cannot connect to the Docker daemon. Is the docker daemon running?")
		return err
	}
	return nil
}

func setPullCmd(image string) exec.Cmd {
	cmd := exec.Command("docker", "pull", image)
	cmd.SetStderr(os.Stderr)
	return cmd
}
