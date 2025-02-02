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
package cmd

import (
	"os"

	"github.com/brightzheng100/vind/pkg/utils"
	"github.com/spf13/cobra"
)

// Footloose is the default name of the footloose file.
const DEFAULT_CONFIG_FILE = "vind.yaml"

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage cluster configuration",
}

func init() {
	rootCmd.AddCommand(configCmd)
}

func configFile(file string) string {
	if file != "" {
		utils.Logger.Debugf("config file used: %s", file)
		return file
	} else {
		file = os.Getenv("VIND_CONFIG")
		utils.Logger.Debugf("no config file specified, try getting from $VIND_CONFIG: %s", file)
		if file != "" {
			utils.Logger.Debugf("config file used: %s", file)
		} else {
			utils.Logger.Debugf("fall back to default config file: %s", DEFAULT_CONFIG_FILE)
			file = DEFAULT_CONFIG_FILE
		}
		return file
	}
}
