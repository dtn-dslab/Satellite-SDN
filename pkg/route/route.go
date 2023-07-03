package route

import (
	"sync"
)

// Return route table for all nodes
func ComputeRoutes(distanceMap [][]float64, threadNum int) [][]int {
	// Initialzie routeTable
	nodeCount := len(distanceMap)
	routeTable := [][]int{}
	for i := 0; i < nodeCount; i++ {
		arr := []int{}
		for j := 0; j < nodeCount; j++ {
			arr = append(arr, -1)
		}
		arr[i] = i
		routeTable = append(routeTable, arr)
	}

	// Start goroutine to compute routes
	var wg sync.WaitGroup
	wg.Add(threadNum)
	for idx := 0; idx < threadNum; idx++ {
		go ComputeRouteThread(distanceMap, routeTable, idx, threadNum, &wg)
	}
	wg.Wait()
	
	return routeTable
}

// Return certain nodes' route table
// routeTable: -1 means the target node is neighbour, other value means the next node packets are forwarded.
func ComputeRouteThread(distanceMap [][]float64, routeTable [][]int, threadID int, threadNum int, wg *sync.WaitGroup) {
	defer wg.Done()
	nodeCount := len(distanceMap)
	for idx := threadID; idx < nodeCount; idx += threadNum {
		// Initialize vector result 
		result, notFound := []float64{}, []float64{}
		dijkstraPath := [][]int{}
		for i := 0; i < nodeCount; i++ {
			result = append(result, 1e9)
			notFound = append(notFound, distanceMap[idx][i])
			dijkstraPath = append(dijkstraPath, []int{})
		}
		result[idx] = 0
		notFound[idx] = -1

		// Begin Dijkstra algorithm
		for i := 1; i < nodeCount; i++ {
			// Find the shortest path point
			var min float64 = 1e9
			var minIndex = 0
			for j := 0; j < nodeCount; j++ {
				if notFound[j] > 0 && notFound[j] < min {
					min = notFound[j]
					minIndex = j
				}
			}

			// Store the shortest path point
			result[minIndex] = min
			notFound[minIndex] = -1

			// Refresh notfound vector
			for j := 0; j < nodeCount; j++ {
				if result[j] == 1e9 && distanceMap[minIndex][j] != 1e9 {
					newDistance := result[minIndex] + distanceMap[minIndex][j]
					if newDistance < notFound[j] {
						notFound[j] = newDistance
						dijkstraPath[j] = []int{}
						dijkstraPath[j] = append(dijkstraPath[j], dijkstraPath[minIndex]...)
						dijkstraPath[j] = append(dijkstraPath[j], minIndex)
					}
				}
			}
		}

		// Udpate routeTable
		for i := 0; i < nodeCount; i++ {
			if len(dijkstraPath[i]) != 0 {
				routeTable[idx][i] = dijkstraPath[i][0]
			} else {
				routeTable[idx][i] = -1
			}
		}
	}
}