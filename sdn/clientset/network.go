package clientset

type NetworkInterface interface {
	UpdateNetwork(info *OrbitInfo)
	CheckConnection(uuid1, uuid2 string, uuidIndexMap map[string]int) bool
	GetTopoInArray(indexUUIDMap map[int]string) [][]string
	GetRouteFromAndTo(uuid1, uuid2 string, indexUUIDMap map[int]string, uuidIndexMap map[string]int) []string
	GetRouteHops(uuid string, uuidList []string, indexUUIDMap map[int]string, uuidIndexMap map[string]int) []int
}

type Network struct {
	// TopoGraph is the topology connection graph.
	TopoGraph [][]bool

	// RouteGraph is the route connection graph.
	RouteGraph [][]int

	// DistanceMap is the map of two nodes to the distance between them
	DistanceMap [][]float64
}

func (n *Network) UpdateNetwork(info *OrbitInfo) {

}

func (n *Network) CheckConnection(uuid1, uuid2 string, uuidIndexMap map[string]int) bool {
	return false
}

func (n *Network) GetTopoInArray(indexUUIDMap map[int]string) [][]string {
	return [][]string{}
}

func (n *Network) GetRouteFromAndTo(uuid1, uuid2 string, indexUUIDMap map[int]string, uuidIndexMap map[string]int) []string {
	result := []string{}
	mid_uuid := uuid1
	for ; mid_uuid != uuid2; {
		mid_index := uuidIndexMap[mid_uuid]
		target_index := uuidIndexMap[uuid2]
		mid_uuid = indexUUIDMap[n.RouteGraph[mid_index][target_index]]
		result = append(result, mid_uuid)
	}
	return result
}

func (n *Network) GetRouteHops(uuid string, uuidList []string, indexUUIDMap map[int]string, uuidIndexMap map[string]int) []int {
	result := []int{}
	for _, target_uuid := range uuidList {
		hopList := n.GetRouteFromAndTo(uuid, target_uuid, indexUUIDMap, uuidIndexMap)
		result = append(result, len(hopList) - 1)
	}
	return result
}

