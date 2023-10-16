package partition

import (
	"testing"
	"ws/dtn-satellite-sdn/sdn/satellite"
	"ws/dtn-satellite-sdn/sdn/link"
)

func TestCutGraph(t *testing.T) {
	nodeSet := []int{1, 2, 3, 4, 5, 6, 7}
	expectedSplitsCount := 3
	graph := GraphCutHash(nodeSet, expectedSplitsCount)
	t.Logf("%v\n", graph)
}

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