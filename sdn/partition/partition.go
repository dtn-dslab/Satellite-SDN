package partition

import (
	"reflect"
	"sort"
	// "fmt"

	"ws/dtn-satellite-sdn/sdn/link"
)

func GraphCutHash(nodeSet []int, expectedSplitsCount int) [][]int {
	ret := make([][]int, expectedSplitsCount)

	for _, nodeId := range nodeSet {
		idx := nodeId % expectedSplitsCount
		ret[idx] = append(ret[idx], nodeId)
	}

	return ret
}

func GraphCutLinear(nodeSet []int, edgeSet []link.LinkEdge, expectedSplitsCount int) [][]int {
	nodeCount, beta := len(nodeSet), 1.0
	// fmt.Println("nodeCount is ", nodeCount)
	allPartitions, nextPartitions := [][]int{}, [][]int{}
	for idx := 0; idx < expectedSplitsCount; idx++ {
		allPartitions = append(allPartitions, []int{})
		nextPartitions = append(nextPartitions, []int{})
	}

	firstIter := true
	for !reflect.DeepEqual(allPartitions, nextPartitions) || firstIter {
		firstIter = false
		copy(allPartitions, nextPartitions)
		// fmt.Println("Iteration!")
		for _, nodeId := range nodeSet {
			// fmt.Println("SubIteration")
			// Get corresponding neighbours
			neighbours := []int{}
			for _, edge := range edgeSet {
				if edge.From == nodeId {
					neighbours = append(neighbours, edge.To)
				} else if edge.To == nodeId {
					neighbours = append(neighbours, edge.From)
				}
			}
			// fmt.Printf("Neighbours of %d:%v\n", nodeId, neighbours)
			// Add new node to one of partitions
			nextPartitions = Partition(nextPartitions, nodeId, neighbours, expectedSplitsCount, nodeCount, beta)
			// fmt.Printf("%v\n", nextPartitions)
			for _, nextPartition := range nextPartitions {
				sort.Slice(nextPartition, func(i, j int) bool {
					return nextPartition[i] < nextPartition[j]
				})
			}
			sort.Slice(nextPartitions, func(i, j int) bool {
				if len(nextPartitions[i]) != 0 && len(nextPartitions[j]) != 0 {
					return nextPartitions[i][0] < nextPartitions[j][0]
				} else {
					return false
				}
			})
		}
	}

	return nextPartitions
}

func Partition(curPartitions [][]int, nodeId int, neighbours []int, expectedSplitsCount int, nodeCount int, beta float64) [][]int {
	targetIdx, targetScore := -1, -1.0
	for idx, partition := range curPartitions {
		// Deleting the same node
		newPartition := []int{}
		for i := range partition {
			if partition[i] != nodeId {
				newPartition = append(newPartition, partition[i])
			}
		}
		// fmt.Printf("newPartition: %v\n", newPartition)
		curPartitions[idx] = newPartition
		// Counting node in both neighbours and partition
		interserctCount := 0
		for _, neighbourId := range neighbours {
			for _, partitionNodeId := range curPartitions[idx] {
				if partitionNodeId == neighbourId {
					interserctCount++
				}
			}
		}
		// Calculate and update score
		C := beta * float64(nodeCount) / float64(expectedSplitsCount)
		weight := float64(1) - float64(len(curPartitions[idx]))/C
		score := float64(interserctCount) * weight
		if score > targetScore {
			targetIdx = idx
			targetScore = score
		}
	}
	// Update partitions
	nextPartitions := [][]int{}
	for idx := 0; idx < expectedSplitsCount; idx++ {
		nextPartitions = append(nextPartitions, []int{})
	}
	copy(nextPartitions, curPartitions)
	nextPartitions[targetIdx] = append(nextPartitions[targetIdx], nodeId)
	return nextPartitions
}

func ComputeEdgesAcrossSubgraphs(nodeSet []int, edgeSet []link.LinkEdge, partitions [][]int) []link.LinkEdge {
	ret := []link.LinkEdge{}
	for _, edge := range edgeSet {
		flag1, flag2 := false, false
		for _, partition := range partitions {
			for _, node := range partition {
				if node == edge.From {
					flag1 = true
				}
				if node == edge.To {
					flag2 = true
				}
			}
			// The edge does not cross graph
			if flag1 && flag2 {
				break
			}
			// Reset flag
			flag1, flag2 = false, false
		}
		if !flag1 && !flag2 {
			ret = append(ret, edge)
		}
	}
	return ret
}
