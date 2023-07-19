package sdn

import (
	"fmt"
	"time"
	"ws/dtn-satellite-sdn/pkg/link"
	"ws/dtn-satellite-sdn/pkg/route"
	"ws/dtn-satellite-sdn/pkg/satellite"
)

// Special case for satellite/constellation:
// Compute edgeSet & route according to my specific logics
// Return nameMap, edgeSet and routeTable
func GenerateSatelliteConfig(inputFilePath string, expectedNodeNum int) (map[int]string, []link.LinkEdge, [][]int, error) {
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

// Compute and apply satellite/constellation configurations periodically
// If timeout < 0, do not update SDN
func RunSatelliteSDN(inputFilePath string, expectedNodeNum int, timeout int) error {
	// Compute configuration & Initialize SDN environment
	nameMap, edgeSet, routeTable, err := GenerateSatelliteConfig(inputFilePath, expectedNodeNum)
	if err != nil {
		return fmt.Errorf("Generate satellite config error: %v\n", err)
	}
	if err := RunSDN(nameMap, edgeSet, routeTable, expectedNodeNum); err != nil {
		return fmt.Errorf("Initialize satellite SDN failed: %v\n", err)
	}

	// Update SDN environment periodically
	if timeout < 0 {
		return nil
	}
	for ;; time.Sleep(time.Duration(timeout) * time.Second) {
		nameMap, edgeSet, routeTable, err := GenerateSatelliteConfig(inputFilePath, expectedNodeNum)
		if err != nil {
			return fmt.Errorf("Generate satellite config error: %v\n", err)
		}
		if err := UpdateSDN(nameMap, edgeSet, routeTable); err != nil {
			return fmt.Errorf("Update satellite SDN failed: %v\n", err)
		}
	}
}

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

