package clientset

import (
	"container/list"
	"fmt"
	"sync"

	"ws/dtn-satellite-sdn/sdn/link"
	"ws/dtn-satellite-sdn/sdn/route"
	satv2 "ws/dtn-satellite-sdn/sdn/type/v2"
	"ws/dtn-satellite-sdn/sdn/util"

	"github.com/sirupsen/logrus"
)

type NetworkInterface interface {
	UpdateNetwork(info *OrbitInfo)
	CheckConnection(idx1, idx2 int) bool
	GetTopoInAscArray() [][]int
	GetRouteFromAndTo(idx1, idx2 int) []int
	GetRouteHops(idx int, idxList []int) []int
	GetDistance(idx1, idx2 int) float64
	GetSpreadArray(idx int) [][]int
}

type Network struct {
	// TopoGraph is the topology connection graph.
	// TopoGraph[i][j] means that node i and node j are directly connected.
	TopoGraph [][]bool

	// RouteGraph is the route connection graph.
	// RouteGraph[i][j] means that the next hop of i to j is RouteGraph[i][j].
	// if i == j, then RouteGraph[i][j] = 0
	RouteGraph [][]int

	// DistanceMap is the map of two nodes to the distance between them
	DistanceMap [][]float64

	// Metadata is the metadata of current orbit info
	Metadata *OrbitMeta
}

func NewNetwork(info *OrbitInfo) *Network {
	network := Network{}
	network.UpdateNetwork(info)
	return &network
}

func (n *Network) UpdateNetwork(info *OrbitInfo) {
	// 1. Init some variables
	n.Metadata = info.Metadata
	totalNodesNum :=
		n.Metadata.LowOrbitNum + n.Metadata.HighOrbitNum +
		n.Metadata.GroundStationNum + n.Metadata.MissileNum +
		n.Metadata.UserNum
	// totalGroupsNum := len(info.LowOrbitSats) + len(info.HighOrbitSats) + 2
	n.DistanceMap = make([][]float64, totalNodesNum)
	n.TopoGraph = make([][]bool, totalNodesNum)
	for i := 0; i < totalNodesNum; i++ {
		n.DistanceMap[i] = make([]float64, totalNodesNum)
		n.TopoGraph[i] = make([]bool, totalNodesNum)
	}

	// 2. Compute DistanceMap(pure, not combined with topology)
	var wg sync.WaitGroup
	wg.Add(util.ThreadNums)
	for threadId := 0; threadId < util.ThreadNums; threadId++ {
		go func(id int) {
			// Parition tasks with different nodes
			for nodeID := id; nodeID < totalNodesNum; nodeID += util.ThreadNums {
				node := n.Metadata.UUIDNodeMap[n.Metadata.IndexUUIDMap[nodeID]]
				for idx := 0; idx < totalNodesNum; idx++ {
					targetNode := n.Metadata.UUIDNodeMap[n.Metadata.IndexUUIDMap[idx]]
					n.DistanceMap[nodeID][idx] = node.DistanceWithNodeAtTime(targetNode, n.Metadata.TimeStamp)
				}
			}
			wg.Done()
		}(threadId)
	}
	wg.Wait()

	// 3. Compute low-orbit topology
	wg.Add(util.ThreadNums)
	lowOrbitGroupNum := len(info.LowOrbitSats)
	lowOrbitGroupKeys := make([]int, 0, lowOrbitGroupNum) // Store trackID in LowOrbitSats
	for key, _ := range info.LowOrbitSats {
		lowOrbitGroupKeys = append(lowOrbitGroupKeys, key)
	}
	for threadId := 0; threadId < util.ThreadNums; threadId++ {
		go func(id int) {
			// Partition tasks with trackID in LowOrbitSats
			for trackIDIdx := id; trackIDIdx < lowOrbitGroupNum; trackIDIdx += util.ThreadNums {
				curTrackID := lowOrbitGroupKeys[trackIDIdx]
				sameOrbitTopoMap := link.GetTopoInGroup(info.LowOrbitSats[curTrackID])
				diffOrbitTopoMap := link.GetTopoAmongLowOrbitGroup(
					info.LowOrbitSats[curTrackID], n.Metadata.LowOrbitNum, n.DistanceMap,
					n.Metadata.IndexUUIDMap, n.Metadata.UUIDIndexMap,
				)
				// Apply result to TopoGraph
				for uuid1, sameUUIDList := range sameOrbitTopoMap {
					uuid1_idx := n.Metadata.UUIDIndexMap[uuid1]
					for _, uuid2 := range sameUUIDList {
						uuid2_idx := n.Metadata.UUIDIndexMap[uuid2]
						n.TopoGraph[uuid1_idx][uuid2_idx] = true
					}
					diffUUIDList := diffOrbitTopoMap[uuid1]
					for _, uuid2 := range diffUUIDList {
						uuid2_idx := n.Metadata.UUIDIndexMap[uuid2]
						n.TopoGraph[uuid1_idx][uuid2_idx] = true
					}
				}
			}
			wg.Done()
		}(threadId)
	}
	wg.Wait()
	// Assert symmetry in topoGraph (Low-orbit Satellites)
	for idx1 := 0; idx1 < n.Metadata.LowOrbitNum; idx1++ {
		for idx2 := idx1 + 1; idx2 < n.Metadata.LowOrbitNum; idx2++ {
			n.TopoGraph[idx2][idx1] = n.TopoGraph[idx1][idx2]
		}
	}

	// 4. Compute ground station & missile & users topology (with low-orbit satellites)
	lowOrbitGroups := []*satv2.Group{}
	for _, group := range info.LowOrbitSats {
		lowOrbitGroups = append(lowOrbitGroups, group)
	}
	// Iterate GroundStations
	for _, gs := range info.GroundStations.Nodes {
		sat_uuid := link.GetMinDistanceNode(&gs, lowOrbitGroups, n.Metadata.TimeStamp)
		gs_idx, sat_idx := n.Metadata.UUIDIndexMap[gs.UUID], n.Metadata.UUIDIndexMap[sat_uuid]
		n.TopoGraph[gs_idx][sat_idx] = true
		n.TopoGraph[sat_idx][gs_idx] = true
	}
	// Iterate Missiles
	for _, missile := range info.Missiles.Nodes {
		sat_uuid := link.GetMinDistanceNode(&missile, lowOrbitGroups, n.Metadata.TimeStamp)
		missile_idx, sat_idx := n.Metadata.UUIDIndexMap[missile.UUID], n.Metadata.UUIDIndexMap[sat_uuid]
		n.TopoGraph[missile_idx][sat_idx] = true
		n.TopoGraph[sat_idx][missile_idx] = true
	}
	// Iterate Users
	for _, user := range info.Users.Nodes {
		sat_uuid := link.GetMinDistanceNode(&user, lowOrbitGroups, n.Metadata.TimeStamp)
		user_idx, sat_idx := n.Metadata.UUIDIndexMap[user.UUID], n.Metadata.UUIDIndexMap[sat_uuid]
		n.TopoGraph[user_idx][sat_idx] = true
		n.TopoGraph[sat_idx][user_idx] = true
	}

	// 5. Compute RouteGraph
	// Get distanceMap for routing(1e9 for edges not directly connected)
	distanceMapForRoute := make([][]float64, totalNodesNum)
	for i := 0; i < totalNodesNum; i++ {
		distanceMapForRoute[i] = make([]float64, totalNodesNum)
		for j := 0; j < totalNodesNum; j++ {
			distanceMapForRoute[i][j] = 1e9
		}
	}
	wg.Add(util.ThreadNums)
	for threadID := 0; threadID < util.ThreadNums; threadID++ {
		go func(id int) {
			// Partition tasks with TopoGraph's row set.
			for idx1 := id; idx1 < totalNodesNum; idx1 += util.ThreadNums {
				for idx2 := 0; idx2 < totalNodesNum; idx2++ {
					// Set distanceMapForRoute[idx1][idx2]
					if n.TopoGraph[idx1][idx2] {
						distanceMapForRoute[idx1][idx2] = n.DistanceMap[idx1][idx2]
					}
				}
			}
			wg.Done()
		}(threadID)
	}
	wg.Wait()
	// Call route calculation func in package route
	n.RouteGraph = route.ComputeRoutes(distanceMapForRoute, util.ThreadNums)

	logrus.WithFields(logrus.Fields{
		"name-map": n.Metadata.IndexUUIDMap,
		"route-graph": n.RouteGraph,
	}).Debug("update network finished")
}

func (n *Network) CheckConnection(idx1, idx2 int) bool {
	return n.TopoGraph[idx1][idx2]
}

func (n *Network) GetTopoInAscArray() [][]int {
	result := [][]int{}
	totalNodesNum := len(n.TopoGraph)
	for idx1 := 0; idx1 < totalNodesNum; idx1++ {
		for idx2 := idx1 + 1; idx2 < totalNodesNum; idx2++ {
			if n.TopoGraph[idx1][idx2] {
				result = append(result, []int{idx1, idx2})
			}
		}
	}
	return result
}

func (n *Network) GetRouteFromAndTo(idx1, idx2 int) []int {
	result := []int{idx1}
	for idx1 != idx2 {
		idx1 = n.RouteGraph[idx1][idx2]
		result = append(result, idx1)
	}
	return result
}

func (n *Network) GetRouteHops(idx int, idxList []int) []int {
	result := []int{}
	for _, target_idx := range idxList {
		hopList := n.GetRouteFromAndTo(idx, target_idx)
		result = append(result, len(hopList))
	}
	return result
}

func (n *Network) GetDistance(idx1, idx2 int) float64 {
	return n.DistanceMap[idx1][idx2]
}

type SpreadLink struct {
	Level int `json:"level"`
	Start string `json:"start"`
	End	  string `json:"end"`
}

func (n *Network) GetSpreadArray(idx int) []SpreadLink {
	visited := map[int]bool{}
	result := []SpreadLink{}
	q := new(list.List)
	q.Init().PushBack(idx)
	visited[idx] = true
	curLevel := 0
	lowOrbitMaxIdx := n.Metadata.LowOrbitNum - 1
	for q.Len() > 0 {
		curLength := q.Len()
		for i := 0; i < curLength; i++ {
			from := q.Front().Value.(int)
			for to := 0; to <= lowOrbitMaxIdx; to++ {
				if n.TopoGraph[from][to] && !visited[to] {
					q.PushBack(to)
					visited[to] = true
					result = append(result, SpreadLink{
						Level: curLevel,
						Start: fmt.Sprint(from),
						End: fmt.Sprint(to),
					})
				}
			}
			q.Remove(q.Front())
		}
		curLevel++
	}
	return result
}
