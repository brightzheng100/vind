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
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchFilter(t *testing.T) {
	const refused = "ssh: connect to host 172.17.0.2 port 22: Connection refused"

	filter := matchFilter{
		writer: ioutil.Discard,
		regexp: connectRefused,
	}

	_, err := filter.Write([]byte("foo\n"))
	assert.NoError(t, err)
	assert.Equal(t, false, filter.matched)

	_, err = filter.Write([]byte(refused))
	assert.NoError(t, err)
	assert.Equal(t, false, filter.matched)
}

func TestNewClusterWithHostPort(t *testing.T) {
	cluster, err := NewFromYAML([]byte(`
cluster:
  name: cluster
  privateKey: cluster-key
machineSets:
- name: centos
  replicas: 2
  spec:
    image: quay.io/brightzheng100/centos7
    name: node%d
    portMappings:
    - containerPort: 22
      hostPort: 2222
`))
	assert.NoError(t, err)
	assert.NotNil(t, cluster)
	assert.Equal(t, 1, len(cluster.config.MachineSets))
	template := cluster.config.MachineSets[0]
	assert.Equal(t, "centos", template.Name)
	assert.Equal(t, 2, template.Replicas)
	assert.Equal(t, 1, len(template.Spec.PortMappings))
	portMapping := template.Spec.PortMappings[0]
	assert.Equal(t, uint16(22), portMapping.ContainerPort)
	assert.Equal(t, uint16(2222), portMapping.HostPort)

	machine0 := newMachine(&cluster.config.Cluster, &cluster.config.MachineSets[0], &template.Spec, 0)
	args0 := machine0.generateContainerRunArgs(cluster.Name())
	i := indexOf("-p", args0)
	assert.NotEqual(t, -1, i)
	assert.Equal(t, "2222:22", args0[i+1])

	machine1 := newMachine(&cluster.config.Cluster, &cluster.config.MachineSets[0], &template.Spec, 1)
	args1 := machine1.generateContainerRunArgs(cluster.Name())
	i = indexOf("-p", args1)
	assert.NotEqual(t, -1, i)
	assert.Equal(t, "2223:22", args1[i+1])
}

func indexOf(element string, array []string) int {
	for k, v := range array {
		if element == v {
			return k
		}
	}
	return -1 // element not found.
}
