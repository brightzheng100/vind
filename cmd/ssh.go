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
	"fmt"
	"strings"

	"github.com/pkg/errors"

	c "github.com/brightzheng100/vind/pkg/cluster"
	"github.com/spf13/cobra"
)

// sshCmd represents the ssh command
var sshCmd = &cobra.Command{
	Use:   "ssh [[USER@]<MACHINE_NAME>]",
	Short: "SSH into a specific machine, or first machine if not specified",
	Args:  validateSSHArgs,
	RunE:  ssh,
}

var configOptions struct {
	extraSshArgs string
}

func init() {
	sshCmd.Flags().StringVarP(&configOptions.extraSshArgs, "extra-ssh-args", "e", "", "Extra args for SSH command")
	rootCmd.AddCommand(sshCmd)
}

func ssh(cmd *cobra.Command, args []string) error {
	cluster, err := c.NewFromFile(configFile(cfgFile.config))
	if err != nil {
		return err
	}

	var machine *c.Machine
	var machineName string
	var userName string

	if len(args) == 0 {
		machine, err = cluster.GetFirstMachine()
		if err != nil {
			return errors.Wrap(err, "SSH into the first machine failed")
		}
	} else {
		if strings.Contains(args[0], "@") {
			items := strings.Split(args[0], "@")
			if len(items) != 2 {
				return fmt.Errorf("bad syntax for user@machineName: %v", items)
			}
			userName = items[0]
			machineName = items[1]
		} else {
			machineName = args[0]
		}

		machine, err = cluster.GetMachineByMachineName(machineName)
		if err != nil {
			return fmt.Errorf("machine name not found: %s", machineName)
		}
	}

	if userName == "" {
		userName = machine.User()
	}

	return cluster.SSH(machine, userName, configOptions.extraSshArgs)
}

func validateSSHArgs(cmd *cobra.Command, args []string) error {
	if len(args) > 1 {
		return errors.New("too many args")
	}
	return nil
}
