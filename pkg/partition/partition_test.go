package partition

import (
	"testing"
	"ws/dtn-satellite-sdn/pkg/satellite"
	"ws/dtn-satellite-sdn/pkg/link"
)

func TestCutGraph(t *testing.T) {
	nodeSet := []int{1, 2, 3, 4, 5, 6, 7}
	edgeSet := []link.LinkEdge{{1, 2}, {1, 3}, {1, 4}, {2, 3}, {2, 4}, {3, 4}, {4, 5}, {5, 6}, {5, 7}, {6, 7}}
	expectedSplitsCount := 3
	graph := GraphCutLinear(nodeSet, edgeSet, expectedSplitsCount)
	t.Logf("%v\n", graph)
	edgesAcrossSubgraphs := ComputeEdgesAcrossSubgraphs(nodeSet, edgeSet, graph)
	t.Logf("%v\n", edgesAcrossSubgraphs)
}

func TestGeneratePodYaml(t *testing.T) {
	constellation, err := satellite.NewConstellation("../data/geodetic.txt")
	if err != nil {
		t.Error(err)
	}

	nameMap := constellation.GetNameMap()
	connGraph := link.GenerateConnGraph(constellation)
	edgeSet := link.ConvertConnGraphToEdgeSet(connGraph)
	GeneratePodSummaryFile(nameMap, edgeSet, "../output/pod.yaml", 3)
}