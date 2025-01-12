/*
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
	"errors"
	"fmt"
	"strings"

	c "github.com/brightzheng100/vind/pkg/cluster"
	"github.com/spf13/cobra"
)

// cpCmd represents the cp command
var cpCmd = &cobra.Command{
	Use:     "cp",
	Aliases: []string{"copy"},
	Short:   "Copy files or folders between a machine and the host file system",
	Long: `
cp <MACHINE_NAME:SRC_PATH> <HOST_DEST_PATH>
cp <HOST_SRC_PATH> <MACHINE_NAME:DEST_PATH>

Copy files or folders between a machine and the host file system
`,
	Args: validateCpArgs,
	RunE: copy,
}

func init() {
	rootCmd.AddCommand(cpCmd)
}

func copy(cmd *cobra.Command, args []string) error {
	cluster, err := c.NewFromFile(configFile(cfgFile.config))
	if err != nil {
		return err
	}
	// from: machine to host
	// to: host to machine
	copyFrom := strings.Contains(args[0], ":")
	copyTo := strings.Contains(args[1], ":")

	if copyFrom && copyTo || !copyFrom && !copyTo {
		return errors.New("either copy from or to machine is supported")
	}

	var machineName, srcPath, destPath string

	if copyFrom { // copy from machine, like: cp machine:/root/ .
		src := strings.Split(args[0], ":")
		machineName = src[0]
		srcPath = src[1]

		destPath = args[1]

		machine, err := cluster.GetMachineByMachineName(machineName)
		if err != nil {
			return fmt.Errorf("machine name not found: %s", machineName)
		}

		return cluster.CopyFrom(machine, srcPath, destPath)
	} else { // copy to machine, like: cp ./file machine:/root/
		srcPath := args[0]

		dest := strings.Split(args[1], ":")
		machineName = dest[0]
		destPath = dest[1]

		machine, err := cluster.GetMachineByMachineName(machineName)
		if err != nil {
			return fmt.Errorf("machine name not found: %s", machineName)
		}

		return cluster.CopyTo(srcPath, machine, destPath)
	}
}

func validateCpArgs(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return errors.New("both src and dest must be provided")
	}
	return nil
}
