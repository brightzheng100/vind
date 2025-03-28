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
	"github.com/brightzheng100/vind/pkg/cluster"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start [MACHINE_NAME1 [MACHINE_NAME2] [...]]",
	Short: "Start all cluster machines or specific machine(s) by given name(s)",
	RunE:  start,
}

func init() {
	rootCmd.AddCommand(startCmd)
}

func start(cmd *cobra.Command, args []string) error {
	cluster, err := cluster.NewFromFile(configFile(cfgFile.config))
	if err != nil {
		return err
	}
	return cluster.Start(args)
}
