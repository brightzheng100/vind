//go:build linux

package cluster

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"syscall"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/moby/sys/signal"
)

type InspectContainerJSON struct {
	*types.ContainerJSONBase
	Mounts          []types.MountPoint
	Config          *InspectContainerConfig
	NetworkSettings *types.NetworkSettings
}

type InspectContainerConfig struct {
	container.Config
}

// UnmarshalJSON allow compatibility with podman V4 API
func (insp *InspectContainerConfig) UnmarshalJSON(data []byte) error {
	type Alias InspectContainerConfig
	aux := &struct {
		Entrypoint interface{} `json:"Entrypoint"`
		StopSignal interface{} `json:"StopSignal"`
		*Alias
	}{
		Alias: (*Alias)(insp),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	switch entrypoint := aux.Entrypoint.(type) {
	case string:
		insp.Entrypoint = strings.Split(entrypoint, " ")
	case []string:
		insp.Entrypoint = entrypoint
	case []interface{}:
		insp.Entrypoint = []string{}
		for _, entry := range entrypoint {
			if str, ok := entry.(string); ok {
				insp.Entrypoint = append(insp.Entrypoint, str)
			}
		}
	case nil:
		insp.Entrypoint = []string{}
	default:
		return fmt.Errorf("cannot unmarshal Config.Entrypoint of type  %T", entrypoint)
	}

	switch stopsignal := aux.StopSignal.(type) {
	case string:
		insp.StopSignal = stopsignal
	case float64:
		insp.StopSignal = ToDockerFormat(uint(stopsignal))
	case nil:
		break
	default:
		return fmt.Errorf("cannot unmarshal Config.StopSignal of type  %T", stopsignal)
	}
	return nil
}

// ParseSysSignalToName translates syscall.Signal to its name in the operating system.
// For example, syscall.Signal(9) will return "KILL" on Linux system.
func ParseSysSignalToName(s syscall.Signal) (string, error) {
	for k, v := range signal.SignalMap {
		if v == s {
			return k, nil
		}
	}
	return "", fmt.Errorf("unknown syscall signal: %s", s)
}

func ToDockerFormat(s uint) string {
	signalStr, err := ParseSysSignalToName(syscall.Signal(s))
	if err != nil {
		return strconv.FormatUint(uint64(s), 10)
	}
	return fmt.Sprintf("SIG%s", signalStr)
}
