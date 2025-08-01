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
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/brightzheng100/vind/pkg/config"
	"github.com/brightzheng100/vind/pkg/docker"
	"github.com/brightzheng100/vind/pkg/exec"
	"github.com/brightzheng100/vind/pkg/utils"
	"github.com/docker/docker/api/types/network"
	"github.com/pkg/errors"
)

const KEY_PATH_ROOT = "/root/.ssh/authorized_keys"
const KEY_PATH_NORMAL = "/home/%s/.ssh/authorized_keys"
const INIT_SCRIPT = `
set -e
rm -f /run/nologin
u=%s
if [[ "$u" == "root" ]]; then
	sshdir=/root/.ssh
	mkdir -p $sshdir; chmod 700 $sshdir
	touch $sshdir/authorized_keys; chmod 600 $sshdir/authorized_keys
else
	sshdir=/home/$u/.ssh
	mkdir -p $sshdir; chmod 700 $sshdir
	touch $sshdir/authorized_keys; chmod 600 $sshdir/authorized_keys
	chown -R $u:$u /home/$u/
fi
`

// defaultUser is the default container user.
const defaultUser = "root"

// Machine is a running machine instance.
type Machine struct {
	spec *config.Machine

	// index in the machine set
	index int

	// containerName is the container name in underlying platform.
	// Naming pattern: {cluster name}-{machineSet name}-{machineName with index}
	containerName string
	// machineName is the machine's name which is also the host's name.
	// Naming pattern: {machineSet name}-{node name with index}
	machineName string

	// runtimeNetwork are networks in Docker runtime
	runtimeNetworks []*RuntimeNetwork

	ports map[int]int
	// maps containerPort -> hostPort.
}

// newMachine inits a new indexed Machine in the cluster.
func newMachine(cluster *config.Cluster, machineSet *config.MachineSet, machine *config.Machine, i int) *Machine {
	return &Machine{
		index:         i,
		spec:          machine,
		containerName: f("%s-%s-"+machine.Name, cluster.Name, machineSet.Name, i),
		machineName:   f("%s-"+machine.Name, machineSet.Name, i),
	}
}

// CreateMachine creates and starts a new machine in the cluster.
func (m *Machine) Create(c *config.Cluster, publicKey []byte) error {
	// Start the container.
	utils.Logger.Infof("Creating machine: %s ...", m.containerName)

	if m.IsCreated() {
		utils.Logger.Infof("Machine %s is already created...", m.containerName)
		return nil
	}

	cmd := []string{"/sbin/init"}
	if strings.TrimSpace(m.spec.Cmd) != "" {
		cmd = strings.Split(strings.TrimSpace(m.spec.Cmd), " ")
	}

	// create the actual Docker container
	runArgs := m.generateContainerRunArgs(c.Name)
	_, err := docker.Create(m.spec.Image,
		runArgs,
		cmd,
	)
	if err != nil {
		return err
	}

	if len(m.spec.Networks) > 1 {
		for _, network := range m.spec.Networks[1:] {
			utils.Logger.Infof("Connecting %s to the %s network...", m.machineName, network)

			// if default "bridge" network is specified, connect to it
			if network == "bridge" {
				if err := docker.ConnectNetwork(m.containerName, network); err != nil {
					return err
				}
			} else {
				if err := docker.ConnectNetworkWithAlias(m.containerName, network, m.machineName); err != nil {
					return err
				}
			}
		}
	}

	// start up the container
	utils.Logger.Infof("Starting machine %s...", m.machineName)
	if err := docker.Start(m.containerName); err != nil {
		return err
	}

	// Initial provisioning.
	var keyPath = KEY_PATH_ROOT
	if m.User() != "root" {
		keyPath = f(KEY_PATH_NORMAL, m.User())
	}
	if err := containerRunShell(m.containerName, f(INIT_SCRIPT, m.User())); err != nil {
		return err
	}
	if err := copy(m.containerName, publicKey, keyPath); err != nil {
		return err
	}

	return nil
}

// generateContainerRunArgs generates the container creation args
func (m *Machine) generateContainerRunArgs(cluster string) []string {
	runArgs := []string{
		"-it",
		"--label", "creator=vind",
		"--label", f("cluster=%s", cluster),
		"--label", f("index=%d", m.index),
		"--name", m.containerName,
		"--hostname", m.machineName,
		"--tmpfs", "/run",
		"--tmpfs", "/run/lock",
		"--tmpfs", "/tmp:exec,mode=777",
		//"-v", "/sys/fs/cgroup:/sys/fs/cgroup:ro",
	}

	for _, volume := range m.spec.Volumes {
		mount := f("type=%s", volume.Type)
		if volume.Source != "" {
			mount += f(",src=%s", volume.Source)
		}
		mount += f(",dst=%s", volume.Destination)
		if volume.ReadOnly {
			mount += ",readonly"
		}
		runArgs = append(runArgs, "--mount", mount)
	}

	for _, mapping := range m.spec.PortMappings {
		publish := ""
		if mapping.Address != "" {
			publish += f("%s:", mapping.Address)
		}
		// add up the index to avoid host port conflicts
		if mapping.HostPort != 0 {
			publish += f("%d:", int(mapping.HostPort)+m.index)
		}
		publish += f("%d", mapping.ContainerPort)
		if mapping.Protocol != "" {
			publish += f("/%s", mapping.Protocol)
		}
		runArgs = append(runArgs, "-p", publish)
	}

	if m.spec.Privileged {
		runArgs = append(runArgs, "--privileged")
	}

	if len(m.spec.Networks) > 0 {
		network := m.spec.Networks[0]
		utils.Logger.Infof("Connecting %s to the %s network...", m.machineName, network)
		runArgs = append(runArgs, "--network", m.spec.Networks[0])
		if network != "bridge" {
			runArgs = append(runArgs, "--network-alias", m.machineName)
		}
	}

	return runArgs
}

// Delete deletes a Machine from the cluster.
func (m *Machine) Delete() error {
	if !m.IsCreated() {
		utils.Logger.Infof("Machine %s hasn't been created", m.machineName)
		return nil
	}

	if m.IsStarted() {
		utils.Logger.Infof("Machine %s is started, stopping and deleting machine...", m.machineName)
		err := docker.Kill("KILL", m.containerName)
		if err != nil {
			return err
		}
		cmd := exec.Command(
			"docker", "rm", "--volumes",
			m.containerName,
		)
		return cmd.Run()
	}
	utils.Logger.Infof("Deleting machine: %s ...", m.machineName)
	cmd := exec.Command(
		"docker", "rm", "--volumes",
		m.containerName,
	)
	return cmd.Run()
}

// Start starts a Machine
func (m *Machine) Start() error {
	if !m.IsCreated() {
		utils.Logger.Infof("Machine %s hasn't been created...", m.machineName)
		return nil
	}
	if m.IsStarted() {
		utils.Logger.Infof("Machine %s is already started...", m.machineName)
		return nil
	}
	utils.Logger.Infof("Starting machine: %s ...", m.machineName)

	// Run command while sigs.k8s.io/kind/pkg/container/docker doesn't
	// have a start command
	cmd := exec.Command(
		"docker", "start",
		m.containerName,
	)
	return cmd.Run()
}

// Stop stops a Machine
func (m *Machine) Stop() error {
	if !m.IsCreated() {
		utils.Logger.Infof("Machine %s hasn't been created...", m.containerName)
		return nil
	}
	if !m.IsStarted() {
		utils.Logger.Infof("Machine %s is already stopped...", m.containerName)
		return nil
	}
	utils.Logger.Infof("Stopping machine: %s ...", m.containerName)

	// Run command while sigs.k8s.io/kind/pkg/container/docker doesn't
	// have a start command
	cmd := exec.Command(
		"docker", "stop",
		m.containerName,
	)
	return cmd.Run()
}

// User gets the machine's OS user, defaults to root if not specified.
func (m *Machine) User() string {
	if m.spec.User == "" {
		return defaultUser
	}
	return m.spec.User
}

// IsCreated returns if a machine is has been created. A created machine could
// either be running or stopped.
func (m *Machine) IsCreated() bool {
	res, err := docker.Inspect(m.containerName, "{{.Name}}")
	if err != nil {
		return false
	}
	if len(res) > 0 && len(res[0]) > 0 {
		return true
	}
	return false
}

// IsStarted returns if a machine is currently started or not.
func (m *Machine) IsStarted() bool {
	res, _ := docker.Inspect(m.containerName, "{{.State.Running}}")
	parsed, _ := strconv.ParseBool(strings.Trim(res[0], `'`))
	return parsed
}

// HostKey returns the host key for a machine
func (m *Machine) HostKey() (string, error) {
	hostPort, err := m.HostPort(22)
	if err != nil {
		return "", err
	}
	mapping, err := mappingFromPort(m.spec, 22)
	if err != nil {
		return "", err
	}
	remote := "localhost"
	if mapping.Address != "" {
		remote = mapping.Address
	}

	cmd := exec.Command("ssh-keyscan",
		"-t", "rsa",
		"-p", f("%d", hostPort),
		remote,
	)
	var buff bytes.Buffer
	cmd.SetStdout(&buff)
	time.Sleep(500 * time.Millisecond)
	err = cmd.Run()
	if err != nil {
		return "error", err
	}
	return buff.String(), err
}

// HostPort returns the host port corresponding to the given container port.
func (m *Machine) HostPort(containerPort int) (int, error) {
	// Use the cached version first
	if hostPort, ok := m.ports[containerPort]; ok {
		return hostPort, nil
	}

	var hostPort int

	// retrieve the specific port mapping using docker inspect
	lines, err := docker.Inspect(m.containerName, fmt.Sprintf("{{(index (index .NetworkSettings.Ports \"%d/tcp\") 0).HostPort}}", containerPort))
	if err != nil {
		return -1, errors.Wrapf(err, "hostport: failed to inspect container: %v", lines)
	}
	if len(lines) != 1 {
		return -1, errors.Errorf("hostport: should only be one line, got %d lines", len(lines))
	}

	port := strings.Replace(lines[0], "'", "", -1)
	if hostPort, err = strconv.Atoi(port); err != nil {
		return -1, errors.Wrap(err, "hostport: failed to parse string to int")
	}

	if m.ports == nil {
		m.ports = make(map[int]int)
	}

	// Cache the result
	m.ports[containerPort] = hostPort
	return hostPort, nil
}

func (m *Machine) networks() ([]*RuntimeNetwork, error) {
	if len(m.runtimeNetworks) != 0 {
		return m.runtimeNetworks, nil
	}

	var networks map[string]*network.EndpointSettings
	if err := docker.InspectObject(m.containerName, ".NetworkSettings.Networks", &networks); err != nil {
		return nil, err
	}
	m.runtimeNetworks = NewRuntimeNetworks(networks)
	return m.runtimeNetworks, nil
}

func (m *Machine) dockerStatus(s *MachineStatus) error {
	var ports []port
	if m.IsCreated() {
		for _, v := range m.spec.PortMappings {
			hPort, err := m.HostPort(int(v.ContainerPort))
			if err != nil {
				hPort = 0
			}
			p := port{
				Host:  hPort,
				Guest: int(v.ContainerPort),
			}
			ports = append(ports, p)
		}
	}
	if len(ports) < 1 {
		for _, p := range m.spec.PortMappings {
			ports = append(ports, port{Host: 0, Guest: int(p.ContainerPort)})
		}
	}
	s.Ports = ports

	s.RuntimeNetworks, _ = m.networks()

	return nil
}

// Status returns the machine status.
func (m *Machine) Status() *MachineStatus {
	s := MachineStatus{}
	s.Container = m.containerName
	s.Image = m.spec.Image
	s.Command = m.spec.Cmd
	s.Spec = m.spec
	s.MachineName = m.machineName
	s.IP = strings.Join(m.IP(), ",")
	state := NotCreated

	if m.IsCreated() {
		state = Stopped
		if m.IsStarted() {
			state = Running
		}
	}
	s.State = state

	_ = m.dockerStatus(&s)

	return &s
}

func (m *Machine) IP() []string {
	ips := []string{}
	for _, network := range m.runtimeNetworks {
		ips = append(ips, network.IP)
	}
	return ips
}

// AutoCdTo is to cd into the current working directory if below bind mount exists.
// For example:
//   - type: bind
//     source: /
//     destination: /host
func (m *Machine) AutoCdTo() string {
	for _, volume := range m.spec.Volumes {
		if volume.Type == "bind" && volume.Destination == "/host" {
			pwd, err := os.Getwd()
			if err != nil {
				utils.Logger.Warn("can't get current working directory: %w", err)
			}
			return fmt.Sprintf("%s%s", "/host", pwd)
		}
	}
	return ""
}
