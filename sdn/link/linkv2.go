package link

import (
	"time"
	satv2 "ws/dtn-satellite-sdn/sdn/type/v2"
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
	curGroup *satv2.Group, distanceMap [][]float64,
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
		for otherNodeIdx, distance := range distanceMap[nodeIdx] {
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
			if distance < minDistance {
				minDistance = distance
				minIdx = otherNodeIdx
			}
		}
		// Get the second nearst satellite
		secondMinDistance, secondMinIdx := 1e9, -1
		for otherNodeIdx, distance := range distanceMap[nodeIdx] {
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
			if distance < secondMinDistance {
				secondMinDistance = distance
				secondMinIdx = otherNodeIdx
			}
		}
		// Update connection graph
		result[node.UUID] = []string{indexUUIDMap[minIdx], indexUUIDMap[secondMinIdx]}
	}

	return result
}