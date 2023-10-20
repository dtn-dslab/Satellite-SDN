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
			if routeTable[idx1][idx2] != -1 {
				// New routes for target Pod
				route.Spec.SubPaths = append(
					route.Spec.SubPaths,
					sdnv1.SubPath{
						Name:     nameMap[idx2],
						TargetIP: util.GetGlobalIP(uint(idx2)),
						NextIP:   util.GetVxlanIP(uint(routeTable[idx1][idx2]), uint(idx1)),
					},
				)
			} else if idx1 != idx2 {
				// Exising routes for target Pod, rewrite it with global IP
				route.Spec.SubPaths = append(
					route.Spec.SubPaths, 
					sdnv1.SubPath{
						Name: nameMap[idx2],
						TargetIP: util.GetGlobalIP(uint(idx2)),
						NextIP: util.GetVxlanIP(uint(idx2), uint(idx1)),
					},
				)
			}

		}
		err := restClient.Post().
			Namespace(namespace).
			Resource("routes").
			Body(&route).
			Do(context.TODO()).
			Into(nil)
		if err != nil {
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
				// Record the next hop's idx
				routeTable[idx][i] = dijkstraPath[i][0]
			} else {
				// Means that they connect to each other directly.
				routeTable[idx][i] = -1
			}
		}
	}
}
