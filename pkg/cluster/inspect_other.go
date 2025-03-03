//go:build !linux

package cluster

import (
	"github.com/docker/docker/api/types"
)

type InspectContainerJSON struct {
	types.ContainerJSON
}
