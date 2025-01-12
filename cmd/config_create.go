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

var configCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a cluster configuration",
	Long: `Create a cluster configuration

A sample of the cluster configuration file may include a list of configurable elements.

For example, below command will create a new vind.yaml as follows:

vind config create -n my-cluster -k key -s ubuntu --networks my-network -r 3 -i brightzheng100/vind-ubuntu22:arm64

cluster:
  name: my-cluster
  privateKey: key
machines:
- count: 3
  name: ubuntu
  spec:
    backend: docker
    image: brightzheng100/vind-ubuntu22:arm64
    name: node%d
    networks:
    - my-network
    portMappings:
    - containerPort: 22
`,
	RunE: configCreate,
}

var configCreateOptions struct {
	override bool
	file     string
}

func init() {
	configCreateCmd.Flags().StringVarP(&configCreateOptions.file, "config", "c", DEFAULT_CONFIG_FILE, "Generated cluster configuration file")
	configCreateCmd.Flags().BoolVarP(&configCreateOptions.override, "override", "o", false, "Override configuration file if it exists")

	name := &defaultConfig.Cluster.Name
	configCreateCmd.PersistentFlags().StringVarP(name, "name", "n", *name, "Name of the cluster")

	private := &defaultConfig.Cluster.PrivateKey
	configCreateCmd.PersistentFlags().StringVarP(private, "key", "k", *private, "Name of the private and public key files")

	machineSetName := &defaultConfig.MachineSets[0].Name
	configCreateCmd.PersistentFlags().StringVarP(machineSetName, "machineset", "s", *machineSetName, "Name of the MachineSet")

	networks := &defaultConfig.MachineSets[0].Spec.Networks
	configCreateCmd.PersistentFlags().StringSliceVar(networks, "networks", *networks, "Networks names the machines are assigned to")

	replicas := &defaultConfig.MachineSets[0].Replicas
	configCreateCmd.PersistentFlags().IntVarP(replicas, "replicas", "r", *replicas, "Number of MachineSet's machine replicas")

	image := &defaultConfig.MachineSets[0].Spec.Image
	configCreateCmd.PersistentFlags().StringVarP(image, "image", "i", *image, "Docker image to use in the containers")

	privileged := &defaultConfig.MachineSets[0].Spec.Privileged
	configCreateCmd.PersistentFlags().BoolVar(privileged, "privileged", *privileged, "Create privileged containers")

	cmd := &defaultConfig.MachineSets[0].Spec.Cmd
	configCreateCmd.PersistentFlags().StringVarP(cmd, "cmd", "d", *cmd, "The command to execute on the container")

	configCmd.AddCommand(configCreateCmd)
}

func configCreate(cmd *cobra.Command, args []string) error {
	opts := &configCreateOptions

	log.Infof("Creating config file %s", opts.file)

	cluster, err := cluster.New(defaultConfig)
	if err != nil {
		return err
	}
	if configExists(configFile(opts.file)) && !opts.override {
		log.Warnf("Failed due to configuration file at %s already exists", opts.file)
		return fmt.Errorf("Configuration file at %s already exists. Override it by specifying --override or -o", opts.file)
	}
	return cluster.Save(configFile(opts.file))
}

// configExists checks whether a configuration file exists.
// Returns false if not true if it already exists.
func configExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) || os.IsPermission(err) {
		return false
	}
	return !info.IsDir()
}
