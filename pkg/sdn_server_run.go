package sdn

import (
	"fmt"
	"log"
	"os/exec"
	"path"
	"ws/dtn-satellite-sdn/pkg/link"
	"ws/dtn-satellite-sdn/pkg/route"
	"ws/dtn-satellite-sdn/pkg/satellite"
)

func ApplyConfig(inputFilePath, outputDir string, availableNodeNum int) error {
	// Initialize constellation
	constellation, err := satellite.NewConstellation(inputFilePath)
	if err != nil {
		return fmt.Errorf("Generating constellation failed: %v", err)
	}

	// Generate connGraph & edgeSet to construct pod & topology file
	nameMap, connGraph := constellation.GenerateConnGraph()
	edgeSet := satellite.ConvertConnGraphToEdgeSet(connGraph)
	distanceMap := constellation.GenerateDistanceMap(connGraph)
	podOutputPath := path.Join(outputDir, "pod.yaml")
	topoOutputPath := path.Join(outputDir, "topology.yaml")

	log.Println("Generating pod yaml...")
	err = link.GeneratePodSummaryFile(nameMap, edgeSet, podOutputPath, availableNodeNum)
	if err != nil {
		return fmt.Errorf("Generating pod yaml failed: %v\n", err)
	}

	log.Println("Generating topology yaml...")
	link.GenerateLinkSummaryFile(nameMap, edgeSet, topoOutputPath)
	if err != nil {
		return fmt.Errorf("Generating topology yaml failed: %v\n", err)
	}

	// Apply topology and pod
	// p.s. We need to apply topology first due to the implementation of kube-dtn.
	topoCmd := exec.Command("kubectl", "apply", "-f", topoOutputPath)
	if err = topoCmd.Run(); err != nil {
		return fmt.Errorf("Apply topology error: %v\n", err)
	}
	podCmd := exec.Command("kubectl", "apply", "-f", podOutputPath)
	if err = podCmd.Run(); err != nil {
		return fmt.Errorf("Apply pod error: %v\n", err)
	}

	// Compute route & Apply
	// p.s. We need to get pod's ip first so that crd controller can delegate route rules to them.
	// 		So, we need to wait until pods are started up successfully.
	routeTable := route.ComputeRoutes(distanceMap, 8)
	routeOutputPath := path.Join(outputDir, "route.yaml")
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