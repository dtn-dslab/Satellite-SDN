package sdn

import (
	"fmt"
	"log"
	"os/exec"
	"path"
	"time"
	"ws/dtn-satellite-sdn/sdn/link"
	"ws/dtn-satellite-sdn/sdn/pod"
	"ws/dtn-satellite-sdn/sdn/route"
	"ws/dtn-satellite-sdn/sdn/satellite"
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
	constellation, err := satellite.NewConstellation(inputFilePath)
	if err != nil {
		return fmt.Errorf("Generating constellation failed: %v", err)
	}

	// Delete SDN environment
	if err := DelSDN(constellation.GetNameMap()); err != nil {
		return fmt.Errorf("Delete satellite SDN failed: %v\n", err)
	}

	return nil
}

// 
// Return nameMap, edgeSet and routeTable
func GenerateSatelliteConfig(inputFilePath string) (map[int]string, []link.LinkEdge, [][]int, error) {
	// Initialize constellation
	constellation, err := satellite.NewConstellation(inputFilePath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Generating constellation failed: %v", err)
	}

	// Generate connGraph & edgeSet to construct pod & topology file
	nameMap := constellation.GetNameMap()
	connGraph := link.GenerateConnGraph(constellation)
	edgeSet := link.ConvertConnGraphToEdgeSet(connGraph)
	distanceMap := link.GenerateDistanceMap(constellation, connGraph)
	routeTable := route.ComputeRoutes(distanceMap, 8)

	return nameMap, edgeSet, routeTable, nil
}

// Construct network emulation system with nameMap, edgeSet, routeTable and expectedNodeNum
func CreateSDN(nameMap map[int]string, edgeSet []link.LinkEdge, routeTable [][]int, expectedNodeNum int) error {
	// Define the path that yaml files store in
	topoOutputPath := path.Join("./sdn/output", "topology.yaml")

	// Generate topology file & apply topology
	log.Println("Generating topology yaml...")
	err := link.GenerateLinkSummaryFile(nameMap, edgeSet, topoOutputPath)
	if err != nil {
		return fmt.Errorf("Generating topology yaml failed: %v\n", err)
	}
	topoCmd := exec.Command("kubectl", "apply", "-f", topoOutputPath)
	if err = topoCmd.Run(); err != nil {
		return fmt.Errorf("Apply topology error: %v\n", err)
	}

	// Generate pod file & apply pod
	// p.s. We need to apply topology first due to the implementation of kube-dtn.
	log.Println("Generate pod yaml...")
	err = pod.PodSyncLoop(nameMap)
	if err != nil {
		return fmt.Errorf("Pod sync failed: %v\n", err)
	}

	endInitTime := time.Now()

	// Generate route file & apply route
	log.Println("Generating route yaml...")
	err = route.RouteSyncLoop(nameMap, routeTable)
	if err != nil {
		return fmt.Errorf("Route sync failed: %v\n", err)
	}
	endApplyRouteTime := time.Now()
	log.Printf("Apply route time %vs", endApplyRouteTime.Sub(endInitTime).Seconds())

	return nil
}

// Update network emulation system with nameMap, edgeSet and routeTable
func UpdateSDN(nameMap map[int]string, edgeSet []link.LinkEdge, routeTable [][]int) error {
	// Define the path that yaml files store in
	topoOutputPath := path.Join("./sdn/output", "topology.yaml")

	// Generate topology file & apply topology
	log.Println("Generating topology yaml...")
	err := link.GenerateLinkSummaryFile(nameMap, edgeSet, topoOutputPath)
	if err != nil {
		return fmt.Errorf("Generating topology yaml failed: %v\n", err)
	}
	topoCmd := exec.Command("kubectl", "apply", "-f", topoOutputPath)
	if err = topoCmd.Run(); err != nil {
		return fmt.Errorf("Apply topology error: %v\n", err)
	}

	// Generate route file & apply route
	log.Println("Generating route yaml...")
	err = route.RouteSyncLoop(nameMap, routeTable)
	if err != nil {
		return fmt.Errorf("Generating route yaml failed: %v\n", err)
	}

	return nil
}

// Uninitialize the network emulation system
func DelSDN(nameMap map[int]string) error {
	// Define the path that yaml files store in
	topoOutputPath := path.Join("./sdn/output", "topology.yaml")

	// Delete Pod
	// TODO(ws): Delete pods in smaller granularity
	podCmd := exec.Command("kubectl", "delete", "pod", "--all")
	if err := podCmd.Run(); err != nil {
		return fmt.Errorf("Delete pod error: %v\n", err)
	}
	// Delete Topology
	topoCmd := exec.Command("kubectl", "delete", "-f", topoOutputPath)
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

