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
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/brightzheng100/vind/pkg/config"
)

const (
	// NotCreated status of a machine
	NotCreated = "Not created"
	// Stopped status of a machine
	Stopped = "Stopped"
	// Running status of a machine
	Running = "Running"
)

// MachineStatus is the runtime status of a Machine.
type MachineStatus struct {
	Container       string            `json:"container"`
	State           string            `json:"state"`
	Spec            *config.Machine   `json:"spec,omitempty"`
	Ports           []port            `json:"ports"`
	MachineName     string            `json:"machineName"`
	Image           string            `json:"image"`
	Command         string            `json:"cmd"`
	IP              string            `json:"ip"`
	RuntimeNetworks []*RuntimeNetwork `json:"runtimeNetworks,omitempty"`
}

// Formatter formats a slice of machines and outputs the result
// in a given format.
type Formatter interface {
	Format(io.Writer, []*Machine) error
}

// JSONFormatter formats a slice of machines into a JSON and
// outputs it to stdout.
type JSONFormatter struct{}

// TableFormatter formats a slice of machines into a colored
// table like output and prints that to stdout.
type TableFormatter struct{}

type port struct {
	Guest int `json:"guest"`
	Host  int `json:"host"`
}

// Format will output to stdout in JSON format.
func (JSONFormatter) Format(w io.Writer, machines []*Machine) error {
	var statuses []MachineStatus
	for _, m := range machines {
		statuses = append(statuses, *m.Status())
	}

	m := struct {
		Machines []MachineStatus `json:"machines"`
	}{
		Machines: statuses,
	}
	ms, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	ms = append(ms, '\n')
	_, err = w.Write(ms)
	return err
}

// FormatSingle is a json formatter for a single machine.
func (JSONFormatter) FormatSingle(w io.Writer, m *Machine) error {
	status, err := json.MarshalIndent(m.Status(), "", "  ")
	if err != nil {
		return err
	}
	_, err = w.Write(status)
	return err
}

// writer contains writeColumns' error value to clean-up some error handling
type writer struct {
	err error
}

// writerColumns is a no-op if there was an error already
func (wr writer) writeColumns(w io.Writer, cols []string) {
	if wr.err != nil {
		return
	}
	_, err := fmt.Fprintln(w, strings.Join(cols, "\t"))
	wr.err = err
}

// Format will output to stdout in table format.
func (TableFormatter) Format(w io.Writer, machines []*Machine) error {
	const padding = 3
	wr := new(writer)
	var statuses []MachineStatus
	for _, m := range machines {
		statuses = append(statuses, *m.Status())
	}

	table := tabwriter.NewWriter(w, 0, 0, padding, ' ', 0)
	wr.writeColumns(table, []string{"CONTAINER NAME", "MACHINE NAME", "PORTS", "IP", "IMAGE", "CMD", "STATE", "BACKEND"})
	// we bail early here if there was an error so we don't process the below loop
	if wr.err != nil {
		return wr.err
	}
	for _, s := range statuses {
		var ports []string
		for _, port := range s.Ports {
			p := fmt.Sprintf("%d->%d", port.Host, port.Guest)
			ports = append(ports, p)
		}
		if len(ports) < 1 {
			for _, p := range s.Spec.PortMappings {
				port := fmt.Sprintf("%d->%d", p.HostPort, p.ContainerPort)
				ports = append(ports, port)
			}
		}
		ps := strings.Join(ports, ",")
		wr.writeColumns(table, []string{s.Container, s.MachineName, ps, s.IP, s.Image, s.Command, s.State, s.Spec.Backend})
	}

	if wr.err != nil {
		return wr.err
	}
	return table.Flush()
}
