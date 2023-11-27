package link

import (
	"context"
	"fmt"
	"log"
	"time"
	satv2 "ws/dtn-satellite-sdn/sdn/type/v2"
	"ws/dtn-satellite-sdn/sdn/util"

	topov1 "github.com/y-young/kube-dtn/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Function: GetMinDistanceNode
// Description: Given a node(most likely ground station/missile), return UUID of node most closest to it.
// 1. node: The node(ground station / missile).
// 2. groups: Node groups to select min node.
// 3. curTime: Standard time to compute distance.
func GetMinDistanceNode(node *satv2.Node, groups []*satv2.Group, curTime time.Time) string {
	minDistance, minUUID := 1e9, ""
	for _, group := range groups {
		for _, group_node := range group.Nodes {
			if distance := node.DistanceWithNodeAtTime(&group_node, curTime); distance < minDistance {
				minDistance = distance
				minUUID = group_node.UUID
			}
		}
	}
	return minUUID
}

// Function: GetTopoInGroup
// Description: Return connection graph for nodes(mainly satellites) in the same group(sorted)
// Return type: UUID -> []UUID
// 1. group: The given Group(mainly satellite group)
func GetTopoInGroup(group *satv2.Group) map[string][]string {
	// Get edge set
	result := map[string][]string{}
	for i := 0; i < group.Len(); i++ {
		result[group.Nodes[i].UUID] = append(
			result[group.Nodes[i].UUID], 
			group.Nodes[(i + 1) % group.Len()].UUID,
		)
		result[group.Nodes[(i + 1) % group.Len()].UUID] = append(
			result[group.Nodes[(i + 1) % group.Len()].UUID], 
			group.Nodes[i].UUID,
		)
	}
	return result
}

// Function: GetTopoAmongLowOrbitGroup
// Description: Return connection graph between different constellations.
// Return type: UUID -> []UUID
// 1. curGroup: current low-orbit satellite group
// 2. distanceMap: distance between nodes
// 3. indexUUIDMap: map from index to uuid
// 4. uuidIndexMap: map from uuid to index
func GetTopoAmongLowOrbitGroup(
	curGroup *satv2.Group, lowOrbitNum int, distanceMap [][]float64,
	indexUUIDMap map[int]string, uuidIndexMap map[string]int) map[string][]string {
	// Initialize some variables
	result := map[string][]string{}
	curNodeIdxs := []int{}
	for _, node := range curGroup.Nodes {
		curNodeIdxs = append(curNodeIdxs, uuidIndexMap[node.UUID])
	}
	
	// Get 2 nearest satellites for each satellite in current groups
	for _, node := range curGroup.Nodes {
		nodeIdx := uuidIndexMap[node.UUID]
		// Get the nearst satellite
		minDistance, minIdx := 1e9, -1
		for otherNodeIdx := 0; otherNodeIdx < lowOrbitNum; otherNodeIdx++ {
			// If otherNodeIdx is in curNodeIdxs, continue
			flag := false
			for _, index := range curNodeIdxs {
				if otherNodeIdx == index { 
					flag = true
					break
				}
			}
			if flag { continue }
			// Update min_distance & min_idx
			if distanceMap[nodeIdx][otherNodeIdx] < minDistance {
				minDistance = distanceMap[nodeIdx][otherNodeIdx]
				minIdx = otherNodeIdx
			}
		}
		// Get the second nearst satellite
		secondMinDistance, secondMinIdx := 1e9, -1
		for otherNodeIdx := 0; otherNodeIdx < lowOrbitNum; otherNodeIdx++ {
			// If otherNodeIdx is in curNodeIdxs or equals to MinIdx, continue
			flag := false
			for _, index := range curNodeIdxs {
				if otherNodeIdx == index || otherNodeIdx == minIdx {
					flag = true
					break
				}
			}
			if flag { continue }
			// Update second_min_distance & second_min_idx
			if distanceMap[nodeIdx][otherNodeIdx] < secondMinDistance {
				secondMinDistance = distanceMap[nodeIdx][otherNodeIdx]
				secondMinIdx = otherNodeIdx
			}
		}
		// Update connection graph
		result[node.UUID] = []string{indexUUIDMap[minIdx], indexUUIDMap[secondMinIdx]}
	}

	return result
}

// Function: LinkSyncLoopV2
// Description: Apply topologies according to indexUUIDMap and topoAscArray
// 1. indexUUIDMap: node's index -> node's uuid
// 2. topoAscArray: Topology graph in ascend array
// 3. isFistTime: true->create, false->update.
func LinkSyncLoopV2(indexUUIDMap map[int]string, topoAscArray [][]int, isFirstTime bool) error {
	// Initialize topologyList
	topoList := topov1.TopologyList{}
	for idx := 0; idx < len(indexUUIDMap); idx++ {
		// podIntfMap[idx] = 1
		topoList.Items = append(topoList.Items, topov1.Topology{
			ObjectMeta: metav1.ObjectMeta{
				Name: indexUUIDMap[idx],
			},
		})
	}

	// Construct topologyList according to topoAscArray
	for _, linkPair := range topoAscArray {
		edgeFrom, edgeTo := linkPair[0], linkPair[1]
		topoList.Items[edgeFrom].Spec.Links = append(
			topoList.Items[edgeFrom].Spec.Links,
			topov1.Link{
				UID:       (edgeFrom << 12) + edgeTo,
				PeerPod:   indexUUIDMap[edgeTo],
				LocalIntf: util.GetLinkName(indexUUIDMap[edgeTo]),
				PeerIntf:  util.GetLinkName(indexUUIDMap[edgeFrom]),
				LocalIP:   util.GetVxlanIP(uint(edgeFrom), uint(edgeTo)),
				PeerIP:    util.GetVxlanIP(uint(edgeTo), uint(edgeFrom)),
			},
		)
		topoList.Items[edgeTo].Spec.Links = append(
			topoList.Items[edgeTo].Spec.Links, 
			topov1.Link{
				UID:       (edgeFrom << 12) + edgeTo,
				PeerPod:   indexUUIDMap[edgeFrom],
				LocalIntf: util.GetLinkName(indexUUIDMap[edgeFrom]),
				PeerIntf:  util.GetLinkName(indexUUIDMap[edgeTo]),
				LocalIP:   util.GetVxlanIP(uint(edgeTo), uint(edgeFrom)),
				PeerIP:    util.GetVxlanIP(uint(edgeFrom), uint(edgeTo)),
			},
		)
	}

	// Get current namespace
	namespace, err := util.GetNamespace()
	if err != nil {
		return fmt.Errorf("get namespace error: %v", err)
	}

	// Create/update topologyList with RESTClient according to variable isFirstTime
	restClient, err := util.GetTopoClient()
	if err != nil {
		return fmt.Errorf("config error: %v", err)
	}
	if isFirstTime {
		log.Println("creating topologies...")
		for _, topo := range topoList.Items {
			if util.DEBUG {
				util.ShowTopology(&topo)
			}
			if err := restClient.Post().
				Namespace(namespace).
				Resource("topologies").
				Body(&topo).
				Do(context.TODO()).
				Into(nil); err != nil {
				return fmt.Errorf("apply topology error: %v", err)
			}
		}
	} else {
		log.Println("updating topologies...")
		resourceVersionMap := map[string]string{}
		topoVersionList := topov1.TopologyList{}
		if err := restClient.Get().
			Namespace(namespace).
			Resource("topologies").
			Do(context.TODO()).
			Into(&topoVersionList); err != nil {
			return fmt.Errorf("get topologylist error: %v", err)
		}
		for _, topo := range topoVersionList.Items {
			resourceVersionMap[topo.Name] = topo.ResourceVersion
		}
		for _, topo := range topoList.Items {
			if util.DEBUG {
				util.ShowTopology(&topo)
			}
			topo.ResourceVersion = resourceVersionMap[topo.Name]
			if err := restClient.Put().
				Namespace(namespace).
				Resource("topologies").
				Name(topo.Name).
				Body(&topo).
				Do(context.TODO()).
				Into(nil); err != nil {
				return fmt.Errorf("update topology error: %v", err)
			}
		}
	}
	
	return nil
}