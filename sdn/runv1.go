package sdn

import (
	"fmt"
	"log"
	"os/exec"
	"time"
	"ws/dtn-satellite-sdn/sdn/link"
	"ws/dtn-satellite-sdn/sdn/pod"
	"ws/dtn-satellite-sdn/sdn/route"
	satv1 "ws/dtn-satellite-sdn/sdn/type/v1"
)

// CLI Interface
// Compute and apply satellite/constellation configurations periodically
// If timeout < 0, do not update SDN
func RunSatelliteSDN(inputFilePath string, expectedNodeNum int, timeout int) error {
	// Compute configuration & Initialize SDN environment
	nameMap, edgeSet, routeTable, err := GenerateSatelliteConfig(inputFilePath)
	if err != nil {
		return fmt.Errorf("Generate satellite config error: %v\n", err)
	}
	if err := CreateSDN(nameMap, edgeSet, routeTable, expectedNodeNum); err != nil {
		return fmt.Errorf("Initialize satellite SDN failed: %v\n", err)
	}

	// Update SDN environment periodically
	if timeout < 0 {
		return nil
	}
	for ;; time.Sleep(time.Duration(timeout) * time.Second) {
		nameMap, edgeSet, routeTable, err := GenerateSatelliteConfig(inputFilePath)
		if err != nil {
			return fmt.Errorf("Generate satellite config error: %v\n", err)
		}
		if err := UpdateSDN(nameMap, edgeSet, routeTable); err != nil {
			return fmt.Errorf("Update satellite SDN failed: %v\n", err)
		}
	}
}

// CLI Interface
// Delete current SDN
func DelSatelliteSDN(inputFilePath string) error {
	// Initialize constellation
	constellation, err := satv1.NewConstellation(inputFilePath)
	if err != nil {
		return fmt.Errorf("Generating constellation failed: %v", err)
	}

	// Delete SDN environment
	if err := DelSDN(constellation.GetNameMap()); err != nil {
		return fmt.Errorf("Delete satellite SDN failed: %v\n", err)
	}

	return nil
}

// Return nameMap, edgeSet and routeTable
func GenerateSatelliteConfig(inputFilePath string) (map[int]string, []link.LinkEdge, [][]int, error) {
	// Initialize constellation
	constellation, err := satv1.NewConstellation(inputFilePath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Generating constellation failed: %v", err)
	}

	// Generate connGraph & edgeSet to construct pod & topology file
	nameMap := constellation.GetNameMap()
	connGraph := link.GenerateConnGraph(constellation)
	edgeSet := link.ConvertConnGraphToEdgeSet(connGraph)
	distanceMap := link.GenerateDistanceMap(constellation, connGraph)
	routeTable := route.ComputeRoutes(distanceMap, 64)

	return nameMap, edgeSet, routeTable, nil
}

// Construct network emulation system with nameMap, edgeSet, routeTable and expectedNodeNum
func CreateSDN(nameMap map[int]string, edgeSet []link.LinkEdge, routeTable [][]int, expectedNodeNum int) error {
	// Invoke link(topology)'s sync loop
	log.Println("Topology Sync...")
	err := link.LinkSyncLoop(nameMap, edgeSet, true)
	if err != nil {
		return fmt.Errorf("Topology sync failed: %v\n", err)
	}

	// Invoke pod's sync loop
	// p.s. We need to apply topology first due to the implementation of kube-dtn.
	log.Println("Pod Sync...")
	err = pod.PodSyncLoop(nameMap)
	if err != nil {
		return fmt.Errorf("Pod sync failed: %v\n", err)
	}

	// Invoke route's sync loop
	log.Println("Route Sync...")
	err = route.RouteSyncLoop(nameMap, routeTable, true)
	if err != nil {
		return fmt.Errorf("Route sync failed: %v\n", err)
	}

	return nil
}

// Update network emulation system with nameMap, edgeSet and routeTable
func UpdateSDN(nameMap map[int]string, edgeSet []link.LinkEdge, routeTable [][]int) error {
	// Invoke link(topology)'s sync loop
	log.Println("Topology Sync...")
	err := link.LinkSyncLoop(nameMap, edgeSet, false)
	if err != nil {
		return fmt.Errorf("Topology sync failed: %v\n", err)
	}

	// Invoke route's sync loop
	log.Println("Route Sync...")
	err = route.RouteSyncLoop(nameMap, routeTable, false)
	if err != nil {
		return fmt.Errorf("Route sync failed: %v\n", err)
	}

	return nil
}

// Uninitialize the network emulation system
func DelSDN(nameMap map[int]string) error {
	// Delete Pod
	// TODO(ws): Delete pods in smaller granularity
	podCmd := exec.Command("kubectl", "delete", "pod", "--all")
	if err := podCmd.Run(); err != nil {
		return fmt.Errorf("Delete pod error: %v\n", err)
	}
	// Delete Topology
	// TODO(ws): Delete topologies in smaller granularity
	topoCmd := exec.Command("kubectl", "delete", "topology", "--all")
	if err := topoCmd.Run(); err != nil {
		return fmt.Errorf("Delete topology error: %v\n", err)
	}
	// Delete Route
	// TODO(ws): Delete routes in smaller granularity
	routeCmd := exec.Command("kubectl", "delete", "route", "--all")
	if err := routeCmd.Run(); err != nil {
		return fmt.Errorf("Delete route error: %v\n", err)
	}

	return nil
}

