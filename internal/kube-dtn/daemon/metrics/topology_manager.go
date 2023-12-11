package metrics

import (
	"fmt"

	v1 "github.com/y-young/kube-dtn/api/v1"
)

type TopologyManager struct {
	topologies map[string]*v1.Topology
}

func NewTopologyManager() *TopologyManager {
	return &TopologyManager{
		topologies: make(map[string]*v1.Topology),
	}
}

func (m *TopologyManager) Init(topologies *v1.TopologyList) error {
	for _, topology := range topologies.Items {
		// https://medium.com/swlh/use-pointer-of-for-range-loop-variable-in-go-3d3481f7ffc9
		// Iteration variables are re-used each iteration,
		// through shadowing we create a new local variable for each iteration.
		topology := topology
		m.topologies[topology.Name] = &topology
	}
	return nil
}

func (m *TopologyManager) Add(topology *v1.Topology) {
	key := fmt.Sprintf("%s/%s", topology.Namespace, topology.Name)
	m.topologies[key] = topology
}

func (m *TopologyManager) Delete(name string, ns string) {
	key := fmt.Sprintf("%s/%s", ns, name)
	delete(m.topologies, key)
}

func (m *TopologyManager) Get(name string, ns string) *v1.Topology {
	key := fmt.Sprintf("%s/%s", ns, name)
	return m.topologies[key]
}

func (m *TopologyManager) List() []*v1.Topology {
	topologies := make([]*v1.Topology, 0, len(m.topologies))
	for _, topology := range m.topologies {
		topologies = append(topologies, topology)
	}
	return topologies
}

func (m *TopologyManager) Update(topology *v1.Topology) {
	m.topologies[topology.Name] = topology
}
