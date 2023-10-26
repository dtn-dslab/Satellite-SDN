package clientset

type NetworkInterface interface {
	UpdateNetwork(info *OrbitInfo)
	CheckConnection(idx1, idx2 int) bool
	GetTopoInAscArray() [][]int
	GetRouteFromAndTo(idx1, idx2 int) []int
	GetRouteHops(idx int, idxList []int) []int
	GetDistance(idx1, idx2 int) float64
}

type Network struct {
	// TopoGraph is the topology connection graph.
	// Key: (int) the index of node
	// Value: ([]int) the index array of nodes connected to node(Key)
	TopoGraph map[int][]int

	// RouteGraph is the route connection graph.
	RouteGraph [][]int

	// DistanceMap is the map of two nodes to the distance between them
	DistanceMap [][]float64

	// Metadata is the metadata of current orbit info
	Metadata *OrbitMeta
}

func (n *Network) UpdateNetwork(info *OrbitInfo) {

}

func (n *Network) CheckConnection(idx1, idx2 int) bool {
	for _, idx := range n.TopoGraph[idx1] {
		if idx == idx2 {
			return true
		}
	}
	return false
}

func (n *Network) GetTopoInAscArray() [][]int {
	result := [][]int{}
	for idx1, idxList := range n.TopoGraph {
		for _, idx2 := range idxList {
			if idx1 < idx2 {
				result = append(result, []int{idx1, idx2})
			}
		}
	}
	return result
}

func (n *Network) GetRouteFromAndTo(idx1, idx2 int) []int {
	result := []int{}
	for ; idx1 != idx2; {
		idx1 = n.RouteGraph[idx1][idx2]
		result = append(result, idx1)
	}
	return result
}

func (n *Network) GetRouteHops(idx int, idxList []int) []int {
	result := []int{}
	for _, target_idx := range idxList {
		hopList := n.GetRouteFromAndTo(idx, target_idx)
		result = append(result, len(hopList) - 1)
	}
	return result
}

func (n *Network) GetDistance(idx1, idx2 int) float64 {
	return n.DistanceMap[idx1][idx2]
}

