package clientset

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
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
	CheckConnectionHandler(w http.ResponseWriter, r *http.Request)
	GetTopoInAscArrayHandler(w http.ResponseWriter, r *http.Request)
	GetRouteFromAndToHandler(w http.ResponseWriter, r *http.Request)
	GetRouteHopsHandler(w http.ResponseWriter, r *http.Request)
	GetDistanceHanlder(w http.ResponseWriter, r *http.Request)
	ApplyPod(nodeNum int) error
	ApplyTopo() error
	ApplyRoute() error
	UpdateTopo() error
	UpdateRoute() error
}

type SDNClient struct {
	// Store OrbitInfo in SDNClient to encapsulate orbit operation
	OrbitClient *OrbitInfo

	// Store Network in SDNClient to encapsulate network operation
	NetworkClient *Network

	// PositionURL is the address of position computing module, by which SDNClient can get each node's position
	PositionURL string

	// RWLock is RWMutex for synchronizing writing threads and reading threads
	RWLock *sync.RWMutex
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
		OrbitClient:   orbit,
		NetworkClient: NewNetwork(orbit),
		PositionURL:   url,
		RWLock:        new(sync.RWMutex),
	}
}

// Function: FetchAndUpdate
// Description: Update OrbitClient and NetworkClient.
func (client *SDNClient) FetchAndUpdate() error {
	client.RWLock.Lock()
	defer client.RWLock.Unlock()
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
	client.RWLock.RLock()
	defer client.RWLock.RUnlock()
	uuidIndexMap := client.OrbitClient.GetUUIDIndexMap()
	if uuid1_index, ok := uuidIndexMap[uuid1]; !ok {
		return false, fmt.Errorf("uuid %s does not exist", uuid1)
	} else if uuid2_index, ok := uuidIndexMap[uuid2]; !ok {
		return false, fmt.Errorf("uuid %s does not exist", uuid2)
	} else {
		return client.NetworkClient.CheckConnection(uuid1_index, uuid2_index), nil
	}
}

// Function: CheckConnectionHandler
// Description: Http handler wrapper for func CheckConnection
func (client *SDNClient) CheckConnectionHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	uuid1, uuid2 := params.Get("src"), params.Get("dst")
	if isTrue, err := client.CheckConnection(uuid1, uuid2); err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
	} else {
		result := map[string]interface{}{
			"result": isTrue,
		}
		content, _ := json.Marshal(&result)
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}
}

// Function: GetTopoInAscArray
// Description: Return topology graph in the form of array
func (client *SDNClient) GetTopoInAscArray() ([][]string, error) {
	client.RWLock.RLock()
	defer client.RWLock.RUnlock()
	indexUUIDMap := client.OrbitClient.GetIndexUUIDMap()
	ascArray := client.NetworkClient.GetTopoInAscArray()
	result := [][]string{}
	for _, idx_pair := range ascArray {
		result = append(result, []string{indexUUIDMap[idx_pair[0]], indexUUIDMap[idx_pair[1]]})
	}
	return result, nil
}

// Function: GetTopoInAscArrayHandler
// Description: Http handler wrapper for func GetTopoInAscArray
func (client *SDNClient) GetTopoInAscArrayHandler(w http.ResponseWriter, r *http.Request) {
	topoArr, _ := client.GetTopoInAscArray()
	result := map[string]interface{}{
		"result": topoArr,
	}
	content, _ := json.Marshal(&result)
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

// Function: GetRouteFromAndTo
// Description: Return the uuid of nodes in the route hops(uuid1, ..., uuid2) from Node(uuid1) to Node(uuid2).
// 1. uuid1: The src node's uuid.
// 2. uuid2: The dst node's uuid.
func (client *SDNClient) GetRouteFromAndTo(uuid1, uuid2 string) ([]string, error) {
	client.RWLock.RLock()
	defer client.RWLock.RUnlock()
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

// Function: GetRouteFromAndToHandler
// Description: Description: Http handler wrapper for func GetRouteFromAndTo
func (client *SDNClient) GetRouteFromAndToHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	uuid1, uuid2 := params.Get("src"), params.Get("dst")
	if routeHops, err := client.GetRouteFromAndTo(uuid1, uuid2); err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
	} else {
		result := map[string]interface{}{
			"result": routeHops,
		}
		content, _ := json.Marshal(result)
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}
}

// Function: GetRouteHops
// Description: Return the list of hop nums from Node(uuid1) to the nodes of which uuid is in uuidList.
// 1. uuid: The src node's uuid.
// 2. uuidList: The dst nodes' uuid list.
func (client *SDNClient) GetRouteHops(uuid, uuidList string) (string, error) {
	client.RWLock.RLock()
	defer client.RWLock.RUnlock()
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
	result = result[:len(result)-1] // Trim the last ','
	return result, nil
}

// Function: GetRouteHopsHanlder
// Description: Http hanlder wrapper for func GetRouteHops
func (client *SDNClient) GetRouteHopsHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	uuid, uuidList := params.Get("preId"), params.Get("saIdList")
	if routeHops, err := client.GetRouteHops(uuid, uuidList); err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
	} else {
		result := map[string]interface{}{
			"result": routeHops,
		}
		content, _ := json.Marshal(result)
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}
}

// Function: GetDistance
// Description: Return the distance from Node(uuid1) to Node(uuid2)
// 1. uuid1: The src node's uuid.
// 2. uuid2: The dst node's uuid.
func (client *SDNClient) GetDistance(uuid1, uuid2 string) (float64, error) {
	client.RWLock.RLock()
	defer client.RWLock.RUnlock()
	uuidIndexMap := client.OrbitClient.GetUUIDIndexMap()
	if uuid1_index, ok := uuidIndexMap[uuid1]; !ok {
		return 0.0, fmt.Errorf("uuid %s does not exist", uuid1)
	} else if uuid2_index, ok := uuidIndexMap[uuid2]; !ok {
		return 0.0, fmt.Errorf("uuid %s does not exist", uuid2)
	} else {
		return client.NetworkClient.GetDistance(uuid1_index, uuid2_index), nil
	}
}

// Function: GetDistanceHandler
// Description: Http handler wrapper for GetDistance
func (client *SDNClient) GetDistanceHanlder(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	uuid1, uuid2 := params.Get("src"), params.Get("dst")
	if distance, err := client.GetDistance(uuid1, uuid2); err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
	} else {
		result := map[string]interface{}{
			"result": distance,
		}
		content, _ := json.Marshal(result)
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}
}

// Function: ApplyPod
// Description: Apply pods according to infos in SDNClient
func (client *SDNClient) ApplyPod(nodeNum int) error {
	allocIdx, uuidAllocNodeMap := 0, map[string]string{}
	kubeNodeList, _ := util.GetSlaveNodes(nodeNum)
	client.RWLock.RLock()
	defer client.RWLock.RUnlock()
	log.Println("Applying pod...")
	// Currently, we only need to allocate low-orbit satellites in one group to the same physical node.
	capacity := map[string]int{}
	for k, v := range util.NodeCapacity {
		capacity[k] = v
	}
	for _, group := range client.OrbitClient.LowOrbitSats {
		for _, node := range group.Nodes {
			uuidAllocNodeMap[node.UUID] = kubeNodeList[allocIdx]
		}
		capacity[kubeNodeList[allocIdx]]--
		if capacity[kubeNodeList[allocIdx]] == 0 {
			allocIdx = (allocIdx + 1) % len(kubeNodeList)
			capacity[kubeNodeList[allocIdx]] = util.NodeCapacity[kubeNodeList[allocIdx]] // restore the value at beginning for next round
		}
	}
	podMeta := pod.PodMetadata{
		IndexUUIDMap: client.OrbitClient.GetIndexUUIDMap(),
		StationIdxMin: client.OrbitClient.Metadata.LowOrbitNum +
			client.OrbitClient.Metadata.HighOrbitNum,
		StationNum: client.OrbitClient.Metadata.GroundStationNum,
	}
	return pod.PodSyncLoopV2(&podMeta, uuidAllocNodeMap)
}

// Function: ApplyTopo
// Description: Apply topologies according to infos in SDNClient
func (client *SDNClient) ApplyTopo() error {
	client.RWLock.RLock()
	defer client.RWLock.RUnlock()
	log.Println("Applying topology...")
	return link.LinkSyncLoopV2(client.OrbitClient.GetIndexUUIDMap(), client.NetworkClient.GetTopoInAscArray(), true)
}

// Function: UpdateTopo
// Description: Update topologies according to infos in SDNClient
func (client *SDNClient) UpdateTopo() error {
	client.RWLock.RLock()
	defer client.RWLock.RUnlock()
	log.Println("Updating topology...")
	return link.LinkSyncLoopV2(client.OrbitClient.GetIndexUUIDMap(), client.NetworkClient.GetTopoInAscArray(), false)
}

// Function: ApplyRoute
// Description: Apply routes according to infos in SDNClient
func (client *SDNClient) ApplyRoute() error {
	client.RWLock.RLock()
	defer client.RWLock.RUnlock()
	log.Println("Applying route...")
	return route.RouteSyncLoop(client.OrbitClient.GetIndexUUIDMap(), client.NetworkClient.RouteGraph, true)
}

// Function: UpdateRoute
// Description: Update routes according to infos in SDNClient
func (client *SDNClient) UpdateRoute() error {
	client.RWLock.RLock()
	defer client.RWLock.RUnlock()
	log.Println("Updating route...")
	return route.RouteSyncLoop(client.OrbitClient.GetIndexUUIDMap(), client.NetworkClient.RouteGraph, false)
}
