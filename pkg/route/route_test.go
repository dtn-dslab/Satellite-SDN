package route

import (
	"testing"
	"ws/dtn-satellite-sdn/pkg/link"
	"ws/dtn-satellite-sdn/pkg/satellite"
)

func TestComputeRoutes(t *testing.T) {
	constellation, err := satellite.NewConstellation("../data/geodetic.txt")
	if err != nil {
		t.Errorf("%v\n", err)
	}

	nameMap, connGraph := link.GenerateConnGraph(constellation)
	edgeSet := link.ConvertConnGraphToEdgeSet(connGraph)
	distanceMap := link.GenerateDistanceMap(constellation, connGraph)
	t.Logf("NameMap: %v\n", nameMap)
	t.Logf("edgeSet: %v\n", edgeSet)

	routeTable := ComputeRoutes(distanceMap, 8)
	t.Logf("RouteTable: %v\n", routeTable)

	err = GenerateRouteSummaryFile(nameMap, routeTable, "../output/route.yaml")
	if err != nil {
		t.Error(err)
	}
}
