package route

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	sdnv1 "ws/dtn-satellite-sdn/api/v1"
	"ws/dtn-satellite-sdn/sdn/util"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func RouteSyncLoop(nameMap map[int]string, routeTable [][]int, isFirstTime bool) error {
	// Get RESTClient and clientset
	restClient, err := util.GetRouteClient()
	if err != nil {
		return fmt.Errorf("config error: %v", err)
	}
	clientset, err := util.GetClientset()
	if err != nil {
		return fmt.Errorf("create clientset error: %v", err)
	}

	// Get current namespace
	namespace, err := util.GetNamespace()
	if err != nil {
		return fmt.Errorf("get namespace error: %v", err)
	}

	// Get pods' ip
	podIPTable := map[string]string{}
	log.Println("Getting ip...")
	if isFirstTime {
		for {
			if podList, err := clientset.CoreV1().
				Pods(namespace).
				List(context.TODO(), v1.ListOptions{}); 
				err == nil {
				isContinue := false
				for _, pod := range podList.Items {
					if pod.Status.PodIP != "" {
						podIPTable[pod.Name] = pod.Status.PodIP
					} else {
						isContinue = true
						break
					}
				}
				if isContinue {
					log.Println("retry")
					duration := 3000 + rand.Int31() % 2000
					time.Sleep(time.Duration(duration) * time.Millisecond)
					continue
				}
			} else {
				log.Panic(err)
			}
			break
		}
	} else {
		if podList, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), v1.ListOptions{}); err == nil {
			for _, pod := range podList.Items {
				podIPTable[pod.Name] = pod.Status.PodIP
			}
		}
	}

	// Construct routes
	routeList := sdnv1.RouteList{}
	for idx1 := range routeTable {
		route := sdnv1.Route{
			Spec: sdnv1.RouteSpec{
				PodIP: podIPTable[nameMap[idx1]],
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
		routeList.Items = append(routeList.Items, route)
	}

	// Create/update routeList with RESTClient according to variable isFirstTime
	if isFirstTime {
		log.Println("Creating routes...")
		// for _, route := range routeList.Items {
		// 	if err := restClient.Post().
		// 		Namespace(namespace).
		// 		Resource("routes").
		// 		Body(&route).
		// 		Do(context.TODO()).
		// 		Into(nil); err != nil {
		// 		return fmt.Errorf("apply route failure: %v", err)
		// 	}
		// }
		wg := new(sync.WaitGroup)
		wg.Add(util.ThreadNums)
		for threadId := 0; threadId < util.ThreadNums; threadId++ {
			go func(id int) {
				for routeId := id; routeId < len(routeList.Items); routeId += util.ThreadNums {
					route := routeList.Items[routeId]
					if err := restClient.Post().
						Namespace(namespace).
						Resource("routes").
						Body(&route).
						Do(context.TODO()).
						Into(nil); err != nil {
						log.Fatalf("apply route failure: %v", err)
					}
				}
				wg.Done()
			}(threadId)
		}
		wg.Wait()
	} else {
		log.Println("Updating routes...")
		log.Println("Fetch route resources version")
		routeVersionList := sdnv1.RouteList{}
		resourceVersionMap := map[string]string{}
		if err := restClient.Get().
			Namespace(namespace).
			Resource("routes").
			Do(context.TODO()).
			Into(&routeVersionList); err != nil {
			return fmt.Errorf("get routelist error: %v", err)
		}
		log.Println("Update route resources version")
		for _, route := range routeVersionList.Items {
			resourceVersionMap[route.Name] = route.ResourceVersion
		}
		// for _, route := range routeList.Items {
		// 	route.ResourceVersion = resourceVersionMap[route.Name]
		// 	if err := restClient.Put().
		// 		Namespace(namespace).
		// 		Resource("routes").
		// 		Name(route.Name).
		// 		Body(&route).
		// 		Do(context.TODO()).
		// 		Into(nil); err != nil {
		// 		return fmt.Errorf("update route failure: %v", err)
		// 	}
		// }
		log.Println("Updating to API Server")
		wg := new(sync.WaitGroup)
		wg.Add(util.ThreadNums)
		for threadId := 0; threadId < util.ThreadNums; threadId++ {
			go func(id int) {
				for routeId := id; routeId < len(routeList.Items); routeId += util.ThreadNums {
					route := routeList.Items[routeId]
					route.ResourceVersion = resourceVersionMap[route.Name]
					if err := restClient.Put().
						Namespace(namespace).
						Resource("routes").
						Name(route.Name).
						Body(&route).
						Do(context.TODO()).
						Into(nil); err != nil {
						log.Fatalf("update route failure: %v", err)
					}
				}
				wg.Done()
			}(threadId)
		}
		wg.Wait()
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
		copy(dist, distanceMap[idx])
		visited[idx] = true
		dijkstraPath := make([]int, nodeCount)	// Record next hop nodes(default is node(idx))
		for i := range dijkstraPath {
			dijkstraPath[i] = i
		}

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
					dijkstraPath[j] = dijkstraPath[minIndex]
				}
			}
			
			// Mark minIndex node as visited
			visited[minIndex] = true
		}

		// Update next-hop table(routeTable)
		for i := 0; i < nodeCount; i++ {
			routeTable[idx][i] = dijkstraPath[i]
		}
	}
}
