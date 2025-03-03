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
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/brightzheng100/vind/pkg/config"
	"github.com/brightzheng100/vind/pkg/docker"
	"github.com/brightzheng100/vind/pkg/exec"
	"github.com/brightzheng100/vind/pkg/utils"
	"github.com/ghodss/yaml"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

// cluster is a running cluster.
type cluster struct {
	config   config.Config
	keyStore *KeyStore
}

// Container represents a running machine.
type Container struct {
	ID string
}

// New creates a new cluster. It takes as input the description of the cluster
// and its machines.
func New(conf config.Config) (*cluster, error) {
	if err := conf.Validate(); err != nil {
		return nil, err
	}
	return &cluster{
		config: conf,
	}, nil
}

// NewFromYAML creates a new Cluster from a YAML serialization of its
// configuration available in the provided string.
func NewFromYAML(data []byte) (*cluster, error) {
	config := config.Config{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return New(config)
}

// NewFromFile creates a new Cluster from a YAML serialization of its
// configuration available in the provided file.
func NewFromFile(path string) (*cluster, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return NewFromYAML(data)
}

// forEachMachine loops through every Machine for doing something
func (c *cluster) forEachMachine(do func(*Machine) error) error {
	for _, machineSet := range c.config.MachineSets {
		for i := 0; i < machineSet.Replicas; i++ {
			machine := newMachine(&c.config.Cluster, &machineSet, &machineSet.Spec, i)
			if err := do(machine); err != nil {
				return err
			}
		}
	}
	return nil
}

// forEachMachine loops through all Machine and locates only specific ones for doing something
func (c *cluster) forSpecificMachines(do func(*Machine) error, machineNames []string) error {
	// machineToStart map is used to track machines to make actions and non existing machines
	machineToHandle := make(map[string]bool)
	for _, machine := range machineNames {
		machineToHandle[machine] = false
	}
	for _, machineSet := range c.config.MachineSets {
		for i := 0; i < machineSet.Replicas; i++ {
			machine := newMachine(&c.config.Cluster, &machineSet, &machineSet.Spec, i)
			if _, ok := machineToHandle[machine.machineName]; ok {
				if err := do(machine); err != nil {
					return err
				}
				machineToHandle[machine.machineName] = true
			}
		}
	}
	// log warning for non existing machines
	for key, value := range machineToHandle {
		if !value {
			utils.Logger.Warnf("machine %v does not exist", key)
		}
	}
	return nil
}

// Create creates the cluster.
func (c *cluster) Create() error {
	// make sure the SSH key pair exists
	if err := c.ensureSSHKey(); err != nil {
		return err
	}

	// make sure Docker is running
	if err := docker.IsRunning(); err != nil {
		return err
	}

	// pull the images if not exist
	for _, template := range c.config.MachineSets {
		if _, err := docker.PullIfNotPresent(template.Spec.Image, 2); err != nil {
			return err
		}
	}

	// create all machines
	return c.forEachMachine(func(m *Machine) error {
		pk, err := c.publicKey(m.spec)
		if err != nil {
			return errors.Wrap(err, "can't retrieve public key")
		}
		return m.Create(&c.config.Cluster, pk)
	})
}

// ensureSSHKey generates SSK key pair when needed
func (c *cluster) ensureSSHKey() error {
	if c.config.Cluster.PrivateKey == "" {
		return nil
	}
	path, _ := homedir.Expand(c.config.Cluster.PrivateKey)
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	utils.Logger.Infof("Creating SSH key: %s ...", path)
	return run(
		"ssh-keygen", "-q",
		"-t", "rsa",
		"-b", "4096",
		"-C", f("%s@vind.mail", c.Name()),
		"-f", path,
		"-N", "",
	)
}

// publicKey retrieves the public key content from machine or clus
func (c *cluster) publicKey(machine *config.Machine) ([]byte, error) {
	// Prefer the machine public key over the cluster-wide key.
	if machine.PublicKey != "" && c.keyStore != nil {
		data, err := c.keyStore.Get(machine.PublicKey)
		if err != nil {
			return nil, err
		}
		data = append(data, byte('\n'))
		return data, err
	}

	// Cluster global key
	if c.config.Cluster.PrivateKey == "" {
		return nil, errors.New("no SSH key provided")
	}

	path, err := homedir.Expand(c.config.Cluster.PrivateKey)
	if err != nil {
		return nil, errors.Wrap(err, "public key expand")
	}
	return os.ReadFile(path + ".pub")
}

func f(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

// SetKeyStore provides a store where to persist public keys for this Cluster.
func (c *cluster) SetKeyStore(keyStore *KeyStore) *cluster {
	c.keyStore = keyStore
	return c
}

// Name returns the cluster name.
func (c *cluster) Name() string {
	return c.config.Cluster.Name
}

// Save writes the Cluster configure to a file.
func (c *cluster) Save(path string) error {
	data, err := yaml.Marshal(c.config)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0666)
}

// Delete deletes the cluster.
func (c *cluster) Delete() error {
	if err := docker.IsRunning(); err != nil {
		return err
	}

	return c.forEachMachine(func(m *Machine) error {
		return m.Delete()
	})
}

// Show will generate information about cluster's running or stopped machines.
func (c *cluster) Show(machineNames []string) (machines []*Machine, err error) {
	if err = docker.IsRunning(); err != nil {
		return nil, err
	}

	// walk through the machineSets
	for _, machineSet := range c.config.MachineSets {
		// walk through the specific machine set
		for i := 0; i < machineSet.Replicas; i++ {
			m := newMachine(&c.config.Cluster, &machineSet, &machineSet.Spec, i)

			// Proceed only if no machine names specified or the machine name is included
			if len(machineNames) == 0 || slices.Contains(machineNames, m.machineName) {
				if !m.IsCreated() {
					utils.Logger.Warnf("machine not created: %s", m.machineName)
					continue
				}

				var inspect InspectContainerJSON
				if err := docker.InspectObject(m.containerName, ".", &inspect); err != nil {
					return machines, err
				}

				// Handle Ports
				ports := make([]config.PortMapping, 0)
				for k, v := range inspect.NetworkSettings.Ports {
					if len(v) < 1 {
						continue
					}
					p := config.PortMapping{}
					hostPort, _ := strconv.Atoi(v[0].HostPort)
					p.HostPort = uint16(hostPort)
					p.ContainerPort = uint16(k.Int())
					p.Address = v[0].HostIP
					ports = append(ports, p)
				}
				m.spec.PortMappings = ports

				// Handle Volumes
				var volumes []config.Volume
				for _, mount := range inspect.Mounts {
					v := config.Volume{
						Type:        string(mount.Type),
						Source:      mount.Source,
						Destination: mount.Destination,
						ReadOnly:    mount.RW,
					}
					volumes = append(volumes, v)
				}
				m.spec.Volumes = volumes

				// Handle network
				m.runtimeNetworks = NewRuntimeNetworks(inspect.NetworkSettings.Networks)

				m.spec.Cmd = strings.Join(inspect.Config.Cmd, ",")

				machines = append(machines, m)
			}
		}
	}
	return
}

// Start starts all or specific machines in cluster.
func (c *cluster) Start(machineNames []string) error {
	if err := docker.IsRunning(); err != nil {
		return err
	}

	startMachineFun := func(m *Machine) error {
		return m.Start()
	}

	// start all if no specific machines are specified
	if len(machineNames) < 1 {
		return c.forEachMachine(startMachineFun)
	}

	// Otherwise, start the specific machines only
	return c.forSpecificMachines(startMachineFun, machineNames)
}

// Stop stops all or specific machines in cluster.
func (c *cluster) Stop(machineNames []string) error {
	if err := docker.IsRunning(); err != nil {
		return err
	}

	stopMachineFun := func(m *Machine) error {
		return m.Stop()
	}

	// stop all if no specific machines are specified
	if len(machineNames) < 1 {
		return c.forEachMachine(stopMachineFun)
	}

	// Otherwise, stop the specific machines only
	return c.forSpecificMachines(stopMachineFun, machineNames)
}

// io.Writer filter that writes that it receives to writer. Keeps track if it
// has seen a write matching regexp.
type matchFilter struct {
	writer       io.Writer
	writeMatched bool // whether the filter should write the matched value or not.

	regexp  *regexp.Regexp
	matched bool
}

func (f *matchFilter) Write(p []byte) (n int, err error) {
	// Assume the relevant log line is flushed in one write.
	if match := f.regexp.Match(p); match {
		f.matched = true
		if !f.writeMatched {
			return len(p), nil
		}
	}
	return f.writer.Write(p)
}

// Matches:
//
//	ssh_exchange_identification: read: Connection reset by peer
var connectRefused = regexp.MustCompile("^ssh_exchange_identification: ")

// Matches:
//
//	Warning:Permanently added '172.17.0.2' (ECDSA) to the list of known hosts
var knownHosts = regexp.MustCompile("^Warning: Permanently added .* to the list of known hosts.")

// ssh returns true if the command should be tried again.
func ssh(args []string) (bool, error) {
	utils.Logger.Debug("ssh", args)
	cmd := exec.Command("ssh", args...)

	refusedFilter := &matchFilter{
		writer:       os.Stderr,
		writeMatched: false,
		regexp:       connectRefused,
	}

	errFilter := &matchFilter{
		writer:       refusedFilter,
		writeMatched: false,
		regexp:       knownHosts,
	}

	cmd.SetStdin(os.Stdin)
	cmd.SetStdout(os.Stdout)
	cmd.SetStderr(errFilter)

	err := cmd.Run()
	if err != nil && refusedFilter.matched {
		return true, err
	}
	return false, err
}

func (c *cluster) GetMachineByMachineName(machineName string) (*Machine, error) {
	for _, machineSet := range c.config.MachineSets {
		for i := 0; i < machineSet.Replicas; i++ {
			if machineName == f("%s-"+machineSet.Spec.Name, machineSet.Name, i) {
				return newMachine(&c.config.Cluster, &machineSet, &machineSet.Spec, i), nil
			}
		}
	}
	return nil, fmt.Errorf("Machine name not found: %s", machineName)
}

func (c *cluster) GetFirstMachine() (*Machine, error) {
	if len(c.config.MachineSets) == 0 {
		return nil, errors.New("no machineSet is configured")
	} else {
		machineSet := c.config.MachineSets[0]
		return newMachine(&c.config.Cluster, &machineSet, &machineSet.Spec, 0), nil
	}
}

func mappingFromPort(spec *config.Machine, containerPort int) (*config.PortMapping, error) {
	for i := range spec.PortMappings {
		if int(spec.PortMappings[i].ContainerPort) == containerPort {
			return &spec.PortMappings[i], nil
		}
	}
	return nil, fmt.Errorf("unknown containerPort %d", containerPort)
}

// SSH logs into the named machine with SSH.
func (c *cluster) SSH(machine *Machine, username string, extraSshArgs string) error {
	utils.Logger.Infof("SSH into machine [%s] with user [%s]", machine.machineName, username)

	hostPort, err := machine.HostPort(22)
	if err != nil {
		return err
	}
	mapping, err := mappingFromPort(machine.spec, 22)
	if err != nil {
		return err
	}
	remote := "localhost"
	if mapping.Address != "" {
		remote = mapping.Address
	}
	path, _ := homedir.Expand(c.config.Cluster.PrivateKey)
	args := []string{
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "StrictHostKeyChecking=no",
		"-o", "IdentitiesOnly=yes",
		"-i", path,
		"-p", f("%d", hostPort),
		"-l", username,
		"-t", remote, // https://stackoverflow.com/questions/626533/how-can-i-ssh-directly-to-a-particular-directory
	}

	if len(extraSshArgs) > 0 {
		// if there are any extra SSH args, let's respect them
		utils.Logger.Infof("With extra SSH args: %s", extraSshArgs)
		args = append(args, extraSshArgs)
	} else {
		// try to auto cd into currently mapped folder
		// if bind mount to "/host" exists
		cd := machine.AutoCdTo()
		if cd != "" {
			utils.Logger.Infof("Trying to cd into: %s", cd)
			args = append(args, fmt.Sprintf("cd %s; exec $SHELL -l", cd))
		}
	}

	// If we ssh in a bit too quickly after the container creation, ssh errors out
	// with:
	//   ssh_exchange_identification: read: Connection reset by peer
	// Let's loop a few times if we receive this message.
	retries := 25
	var retry bool
	for retries > 0 {
		retry, err = ssh(args)
		if !retry {
			break
		}
		retries--
		time.Sleep(200 * time.Millisecond)
	}

	return err
}

// CopyFrom copies files/folders from the machine to the host filesystem
func (c *cluster) CopyFrom(from *Machine, srcPath, destPath string) error {
	// CopyTo(hostPath, containerNameOrID, destPath string) error
	return docker.CopyFrom(from.containerName, srcPath, destPath)
}

// CopyTo copies files/folders from the host filesystem to the machine
func (c *cluster) CopyTo(srcPath string, to *Machine, destPath string) error {
	return docker.CopyTo(srcPath, to.containerName, destPath)
}
