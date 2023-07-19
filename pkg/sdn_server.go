package sdn

import (
	"fmt"
	"log"
	"os/exec"
	"path"

	"ws/dtn-satellite-sdn/pkg/link"
	"ws/dtn-satellite-sdn/pkg/route"
	"ws/dtn-satellite-sdn/pkg/partition"
)

// Construct network emulation system with nameMap, edgeSet, routeTable and expectedNodeNum
func RunSDN(nameMap map[int]string, edgeSet []link.LinkEdge, routeTable [][]int, expectedNodeNum int) error {
	// Define the path that yaml files store in
	podOutputPath := path.Join("./output", "pod.yaml")
	topoOutputPath := path.Join("./output", "topology.yaml")
	routeOutputPath := path.Join("./output", "route.yaml")

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
	err = partition.GeneratePodSummaryFile(nameMap, edgeSet, podOutputPath, expectedNodeNum)
	if err != nil {
		return fmt.Errorf("Generating pod yaml failed: %v\n", err)
	}
	podCmd := exec.Command("kubectl", "apply", "-f", podOutputPath)
	if err = podCmd.Run(); err != nil {
		return fmt.Errorf("Apply pod error: %v\n", err)
	}

	// Generate route file & apply route
	log.Println("Generating route yaml...")
	err = route.GenerateRouteSummaryFile(nameMap, routeTable, routeOutputPath)
	if err != nil {
		return fmt.Errorf("Generating route yaml failed: %v\n", err)
	}
	routeCmd := exec.Command("kubectl", "apply", "-f", routeOutputPath)
	if err = routeCmd.Run(); err != nil {
		return fmt.Errorf("Apply route error: %v\n", err)
	}

	return nil
}

// Update network emulation system with nameMap, edgeSet and routeTable
func UpdateSDN(nameMap map[int]string, edgeSet []link.LinkEdge, routeTable [][]int) error {
	// Define the path that yaml files store in
	topoOutputPath := path.Join("./output", "topology.yaml")
	routeOutputPath := path.Join("./output", "route.yaml")

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
	err = route.GenerateRouteSummaryFile(nameMap, routeTable, routeOutputPath)
	if err != nil {
		return fmt.Errorf("Generating route yaml failed: %v\n", err)
	}
	routeCmd := exec.Command("kubectl", "apply", "-f", routeOutputPath)
	if err = routeCmd.Run(); err != nil {
		return fmt.Errorf("Apply route error: %v\n", err)
	}

	return nil
}

// Uninitialize the network emulation system
func DelSDN(nameMap map[int]string) error {
	// Define the path that yaml files store in
	podOutputPath := path.Join("./output", "pod.yaml")
	topoOutputPath := path.Join("./output", "topology.yaml")
	routeOutputPath := path.Join("./output", "route.yaml")

	// Delete Pod
	podCmd := exec.Command("kubectl", "delete", "-f", podOutputPath)
	if err := podCmd.Run(); err != nil {
		return fmt.Errorf("Delete pod error: %v\n", err)
	}
	// Delete Topology
	topoCmd := exec.Command("kubectl", "delete", "-f", topoOutputPath)
	if err := topoCmd.Run(); err != nil {
		return fmt.Errorf("Delete topology error: %v\n", err)
	}
	// Delete Route
	routeCmd := exec.Command("kubectl", "delete", "-f", routeOutputPath)
	if err := routeCmd.Run(); err != nil {
		return fmt.Errorf("Delete route error: %v\n", err)
	}

	return nil
}