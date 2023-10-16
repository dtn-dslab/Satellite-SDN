package pod

import (
	"testing"
	"ws/dtn-satellite-sdn/sdn/satellite"
	"ws/dtn-satellite-sdn/sdn/link"
)

func TestGeneratePodYaml(t *testing.T) {
	constellation, err := satellite.NewConstellation("../data/starlink863.txt")
	if err != nil {
		t.Error(err)
	}

	nameMap := constellation.GetNameMap()
	connGraph := link.GenerateConnGraph(constellation)
	edgeSet := link.ConvertConnGraphToEdgeSet(connGraph)
	GeneratePodSummaryFile(nameMap, edgeSet, "../output/pod.yaml", 7)
}