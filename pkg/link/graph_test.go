package link

import (
	"testing"

	"ws/dtn-satellite-sdn/pkg/satellite"
)

func TestCutGraph(t *testing.T) {
	nodeSet := []int{1, 2, 3, 4, 5, 6, 7}
	edgeSet := []satellite.LinkEdge{{1, 2}, {1, 3}, {1, 4}, {2, 3}, {2, 4}, {3, 4}, {4, 5}, {5, 6}, {5, 7}, {6, 7}}
	expectedSplitsCount := 3
	graph := GraphCutLinear(nodeSet, edgeSet, expectedSplitsCount)
	t.Logf("%v\n", graph)
	edgesAcrossSubgraphs := ComputeEdgesAcrossSubgraphs(nodeSet, edgeSet, graph)
	t.Logf("%v\n", edgesAcrossSubgraphs)
}

func TestGeneratePodYaml(t *testing.T) {
	constellation, err := satellite.NewConstellation("../data/geodetic.txt")
	if err != nil {
		t.Errorf("%v\n", err)
	}

	GeneratePodSummaryFile(constellation, "../output/pod.yaml", 3)
}