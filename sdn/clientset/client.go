package clientset

import (
	"fmt"
	"log"
	"strings"
	"ws/dtn-satellite-sdn/sdn/link"
	"ws/dtn-satellite-sdn/sdn/pod"
	"ws/dtn-satellite-sdn/sdn/route"
	"ws/dtn-satellite-sdn/sdn/util"
)

type ClientInterface interface {
	FetchAndUpdate() error
	CheckConnection(uuid1, uuid2 string) (bool, error)
	GetTopoInAscArray() ([][]string, error)
	GetRouteFromAndTo(uuid1, uuid2 string) ([]string, error)
	GetRouteHops(uuid, uuidList string) (string, error)
	GetDistance(uuid1, uuid2 string) (float64, error)
	ApplyPod() error
	ApplyTopo() error
	ApplyRoute() error
}

type SDNClient struct {
	// Store OrbitInfo in SDNClient to encapsulate orbit operation
	OrbitClient *OrbitInfo

	// Store Network in SDNClient to encapsulate network operation
	NetworkClient *Network

	// PositionURL is the address of position computing module, by which SDNClient can get each node's position
	PositionURL string
}

// Function: NewSDNClient
// Description: Create SDNClient with the address of the module that computes position of each node
// 1. url: The address of the module that computes position of each node
func NewSDNClient(url string) *SDNClient {
	params, err := util.Fetch(url)
	if err != nil {
		log.Fatal(err)
	}
	orbit := NewOrbitInfo(params)
	return &SDNClient{
		OrbitClient: orbit,
		NetworkClient: NewNetwork(orbit),
		PositionURL: url,
	}
}

// Function: FetchAndUpdate
// Description: Update OrbitClient and NetworkClient.
func (client *SDNClient) FetchAndUpdate() error {
	if params, err := util.Fetch(client.PositionURL); err != nil {
		return fmt.Errorf("failed to update SDN: %v", err)
	} else {
		client.OrbitClient.Update(params)
		client.NetworkClient.UpdateNetwork(client.OrbitClient)
		return nil
	}
}

// Function: CheckConnection
// Description: Check if Node(uuid1) connects to Node(uuid2) directly.
// 1. uuid1: The first node's uuid.
// 2. uuid2: The second node's uuid.
func (client *SDNClient) CheckConnection(uuid1, uuid2 string) (bool, error) {
	uuidIndexMap := client.OrbitClient.GetUUIDIndexMap()
	if uuid1_index, ok := uuidIndexMap[uuid1]; !ok {
		return false, fmt.Errorf("uuid %s does not exist", uuid1)
	} else if uuid2_index, ok := uuidIndexMap[uuid2]; !ok {
		return false, fmt.Errorf("uuid %s does not exist", uuid2)
	} else {
		return client.NetworkClient.CheckConnection(uuid1_index, uuid2_index), nil
	}
}

// Function: GetTopoInAscArray
// Description: Return topology graph in the form of array
func (client *SDNClient) GetTopoInAscArray() ([][]string, error) {
	indexUUIDMap := client.OrbitClient.GetIndexUUIDMap()
	ascArray := client.NetworkClient.GetTopoInAscArray()
	result := [][]string{}
	for _, idx_pair := range ascArray {
		result = append(result, []string{indexUUIDMap[idx_pair[0]], indexUUIDMap[idx_pair[1]]})
	}
	return result, nil
}

// Function: GetRouteFromAndTo
// Description: Return the uuid of nodes in the route hops(uuid1, ..., uuid2) from Node(uuid1) to Node(uuid2).
// 1. uuid1: The src node's uuid.
// 2. uuid2: The dst node's uuid.
func (client *SDNClient) GetRouteFromAndTo(uuid1, uuid2 string) ([]string, error) {
	uuidIndexMap := client.OrbitClient.GetUUIDIndexMap()
	indexUUIDMap := client.OrbitClient.GetIndexUUIDMap()
	if uuid1_index, ok := uuidIndexMap[uuid1]; !ok {
		return []string{}, fmt.Errorf("uuid %s does not exist", uuid1)
	} else if uuid2_index, ok := uuidIndexMap[uuid2]; !ok {
		return []string{}, fmt.Errorf("uuid %s does not exist", uuid2)
	} else {
		idxList := client.NetworkClient.GetRouteFromAndTo(uuid1_index, uuid2_index)
		result := []string{}
		for _, idx := range idxList {
			result = append(result, indexUUIDMap[idx])
		}
		return result, nil
	}
}

// Function: GetRouteHops
// Description: Return the list of hop nums from Node(uuid1) to the nodes of which uuid is in uuidList.
// 1. uuid: The src node's uuid.
// 2. uuidList: The dst nodes' uuid list.
func (client *SDNClient) GetRouteHops(uuid, uuidList string) (string, error) {
	var result string = ""
	uuidIndexMap := client.OrbitClient.GetUUIDIndexMap()
	uuid_index, ok := 0, false
	if uuid_index, ok = uuidIndexMap[uuid]; !ok {
		return "", fmt.Errorf("uuid %s does not exist", uuid)
	}
	uuidArray := strings.Split(uuidList, ",")
	for _, target_uuid := range uuidArray {
		if target_uuid_index, ok := uuidIndexMap[target_uuid]; !ok {
			return "", fmt.Errorf("uuid %s does not exist", target_uuid)
		} else {
			idxList := client.NetworkClient.GetRouteFromAndTo(uuid_index, target_uuid_index)
			result += fmt.Sprint(len(idxList)) + ","
		}
	}
	result = result[:len(result) - 1]	// Trim the last ','
	return result, nil
}

// Function: GetDistance
// Description: Return the distance from Node(uuid1) to Node(uuid2)
// 1. uuid1: The src node's uuid.
// 2. uuid2: The dst node's uuid.
func (client *SDNClient) GetDistance(uuid1, uuid2 string) (float64, error) {
	uuidIndexMap := client.OrbitClient.GetUUIDIndexMap()
	if uuid1_index, ok := uuidIndexMap[uuid1]; !ok {
		return 0.0, fmt.Errorf("uuid %s does not exist", uuid1)
	} else if uuid2_index, ok := uuidIndexMap[uuid2]; !ok {
		return 0.0, fmt.Errorf("uuid %s does not exist", uuid2)
	} else {
		return client.NetworkClient.GetDistance(uuid1_index, uuid2_index), nil
	}
}

// Function: ApplyPod
// Description: Apply pods according to infos in SDNClient
func (client *SDNClient) ApplyPod() error {
	allocIdx, uuidAllocNodeMap := 0, map[string]string{}
	kubeNodeList, _ := util.GetSlaveNodes()
	// Currently, we only need to allocate low-orbit satellites in one group to the same physical node.
	for _, group := range client.OrbitClient.LowOrbitSats {
		for _, node := range group.Nodes {
			uuidAllocNodeMap[node.UUID] = kubeNodeList[allocIdx]
		}
		allocIdx = (allocIdx + 1) % len(kubeNodeList)
	}
	return pod.PodSyncLoopV2(client.OrbitClient.GetIndexUUIDMap(), uuidAllocNodeMap)
}

// Function: ApplyTopo
// Description: Apply topologies according to infos in SDNClient
func (client *SDNClient) ApplyTopo() error {
	return link.LinkSyncLoopV2(client.OrbitClient.GetIndexUUIDMap(), client.NetworkClient.GetTopoInAscArray())
}

// Function: ApplyRoute
// Description: Apply routes according to infos in SDNClient
func (client *SDNClient) ApplyRoute() error {
	return route.RouteSyncLoop(client.OrbitClient.GetIndexUUIDMap(), client.NetworkClient.RouteGraph)
}



