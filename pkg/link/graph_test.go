package link

import (
	"testing"
)

func TestCutGraph(t *testing.T) {
	nodeSet := []int{1, 2, 3, 4, 5, 6, 7}
	edgeSet := []Edge{{1, 2}, {1, 3}, {1, 4}, {2, 3}, {2, 4}, {3, 4}, {4, 5}, {5, 6}, {5, 7}, {6, 7}}
	expectedSplitsCount := 3
	graph := GraphCutLinear(nodeSet, edgeSet, expectedSplitsCount)
	t.Logf("%v\n", graph)
	edgesAcrossSubgraphs := ComputeEdgesAcrossSubgraphs(nodeSet, edgeSet, graph)
	t.Logf("%v\n", edgesAcrossSubgraphs)
}