package route

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
	"ws/dtn-satellite-sdn/sdn/util"

	sdnv1 "ws/dtn-satellite-sdn/api/v1"
)

func RouteSyncLoop(nameMap map[int]string, routeTable [][]int) error {
	// Get RESTClient
	restClient, err := util.GetRouteClient()
	if err != nil {
		return fmt.Errorf("CONFIG ERROR: %v", err)
	}

	// Get current namespace
	namespace, err := util.GetNamespace()
	if err != nil {
		return fmt.Errorf("GET NAMESPACE ERROR: %v", err)
	}

	// Get pods' ip
	nodeCount := len(nameMap)
	podIPTable := []string{}
	for idx := 0; idx < nodeCount; idx++ {
		var podIP string
		var err error
		for podIP, err = util.GetPodIP(nameMap[idx]); err != nil; podIP, err = util.GetPodIP(nameMap[idx]) {
			log.Println("Retry")
			duration := 3000 + rand.Int31()%2000
			time.Sleep(time.Duration(duration) * time.Millisecond)
		}
		podIPTable = append(podIPTable, podIP)
	}

	// Construct routes and apply
	for idx1 := range routeTable {
		route := sdnv1.Route{
			Spec: sdnv1.RouteSpec{
				PodIP: podIPTable[idx1],
				SubPaths: []sdnv1.SubPath{},
			},
		}
		route.APIVersion  = "sdn.dtn-satellite-sdn/v1"
		route.Kind = "Route"
		route.Name = nameMap[idx1]
		for idx2 := range routeTable[idx1] {
			if idx1 != idx2 {
				// New routes for target Pod
				route.Spec.SubPaths = append(
					route.Spec.SubPaths,
					sdnv1.SubPath{
						Name:     nameMap[idx2],
						TargetIP: util.GetGlobalIP(uint(idx2)),
						NextIP:   util.GetVxlanIP(uint(routeTable[idx1][idx2]), uint(idx1)),
					},
				)
			}
		}
		if err := restClient.Post().
			Namespace(namespace).
			Resource("routes").
			Body(&route).
			Do(context.TODO()).
			Into(nil); err != nil {
			return fmt.Errorf("APPLY ROUTE FAILURE: %v", err)
		}
	}

	return nil	
}

// Return route table for all nodes
func ComputeRoutes(distanceMap [][]float64, threadNum int) [][]int {
	// Initialzie routeTable
	nodeCount := len(distanceMap)
	routeTable := [][]int{}
	for i := 0; i < nodeCount; i++ {
		routeTable = append(routeTable, make([]int, nodeCount))
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
// routeTable: means the next node packets forwarded.
func ComputeRouteThread(distanceMap [][]float64, routeTable [][]int, threadID int, threadNum int, wg *sync.WaitGroup) {
	defer wg.Done()
	nodeCount := len(distanceMap)
	for idx := threadID; idx < nodeCount; idx += threadNum {
		// Initialize some variables
		visited := make([]bool, nodeCount)
		dist := make([]float64, nodeCount)
		dijkstraPath := make([][]int, nodeCount)	// Record nodes in the middle(does not contain the leftmost and rightmost nodes)
		copy(dist, distanceMap[idx])
		visited[idx] = true

		// Use dijkstra algorithm to compute minimum distance and their routes
		for i := 1; i < nodeCount; i++ {
			// Use minDist and minIndex to record minimum distance and corresponding node's index
			minDist, minIndex := 1e9, 0

			// Iterate all nodes, find the next unvisted and nearst node to startNode.
			for j := 0; j < nodeCount; j++ {
				if !visited[j] && dist[j] < minDist {
					minDist = dist[j]
					minIndex = j
				}
			}

			// Iterate other nodes, update their minium distance to startNode
			for j := 0; j < nodeCount; j++ {
				if !visited[j] && dist[minIndex] + distanceMap[minIndex][j] < dist[j] {
					dist[j] = dist[minIndex] + distanceMap[minIndex][j]
					copy(dijkstraPath[j], dijkstraPath[minIndex])
					dijkstraPath[j] = append(dijkstraPath[j], minIndex)
				}
			}

			// Mark minIndex node as visited
			visited[minIndex] = true
		}

		// Update next-hop table(routeTable)
		for i := 0; i < nodeCount; i++ {
			if len(dijkstraPath[i]) != 0 {
				// Nodes not connected to startNode directly
				routeTable[idx][i] = dijkstraPath[i][0]
			} else {
				// Nodes connected to startNode directly
				routeTable[idx][i] = i
			}
		}
	}
}
