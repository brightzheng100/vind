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
	"os"

	"github.com/brightzheng100/vind/pkg/cluster"
	"github.com/spf13/cobra"
)

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:     "show [MACHINE_NAME1 [MACHINE_NAME2] [...]]",
	Aliases: []string{"status"},
	Short:   "Show all running machines or some specific machine(s) by the given machine name(s).",
	Long: `Shows all cluster machines or some specific machine(s) if the given machine name(s) are specified, in JSON or Table format.
`,
	RunE: show,
}

var showOptions struct {
	output string
}

func init() {
	showCmd.Flags().StringVarP(&showOptions.output, "output", "o", "table", "Output formatting options: {table,json,ansible,ssh}.")
	rootCmd.AddCommand(showCmd)
}

// show will show all machines in a given cluster.
func show(cmd *cobra.Command, args []string) error {
	c, err := cluster.NewFromFile(configFile(cfgFile.config))
	if err != nil {
		return err
	}
	var formatter cluster.Formatter
	switch showOptions.output {
	case "table":
		formatter = new(cluster.TableFormatter)
	case "json":
		formatter = new(cluster.JSONFormatter)
	case "ansible":
		formatter = new(cluster.AnsibleFormatter)
	case "ssh":
		formatter = new(cluster.SSHConfigFormatter)
	default:
		return fmt.Errorf("unknown formatter '%s'", showOptions.output)
	}
	machines, err := c.Show(args)
	if err != nil {
		return err
	}
	return formatter.Format(os.Stdout, c, machines)
}
