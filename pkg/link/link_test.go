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
		t.Error(err)
	}

	nameMap, connGraph := constellation.GenerateConnGraph()
	edgeSet := satellite.ConvertConnGraphToEdgeSet(connGraph)
	GeneratePodSummaryFile(nameMap, edgeSet, "../output/pod.yaml", 3)
}

func TestGenerateIP(t *testing.T) {
	ip := GenerateIP(1)
	t.Logf("IP is %s\n", ip)
	if ip != "128.0.0.1/32" {
		t.Errorf("IP Dismatch!\n")
	}
}

func TestGenerateLinkYaml(t *testing.T) {
	constellation, err := satellite.NewConstellation("../data/geodetic.txt")
	if err != nil {
		t.Error(err)
	}

	nameMap, connGraph := constellation.GenerateConnGraph()
	edgeSet := satellite.ConvertConnGraphToEdgeSet(connGraph)
	err = GenerateLinkSummaryFile(nameMap, edgeSet, "../output/topology.yaml")
	if err != nil {
		t.Error(err)
	}
}
