package link

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"
	"ws/dtn-satellite-sdn/sdn/util"
)

func GenerateLinkSummaryFile(nameMap map[int]string, edgeSet []LinkEdge, outputPath string) error {
	// Initialize topologyList
	topologyList := util.TopologyList{}
	topologyList.APIVersion = "v1"
	topologyList.Kind = "List"
	podIntfMap := make([]int, len(nameMap))
	for idx := 0; idx < len(nameMap); idx++ {
		podIntfMap[idx] = 1
		topologyList.Items = append(topologyList.Items, util.Topology{
			APIVersion: "y-young.github.io/v1",
			Kind:       "Topology",
			MetaData: util.MetaData{
				Name: nameMap[idx],
			},
		})
	}

	// Construct topologyList according to edgeSet
	for _, edge := range edgeSet {
		topologyList.Items[edge.From].Spec.Links = append(
			topologyList.Items[edge.From].Spec.Links,
			util.Link{
				UID:       (edge.From << 12) + edge.To,
				PeerPod:   nameMap[edge.To],
				LocalIntf: fmt.Sprintf("sdneth%d", podIntfMap[edge.From]),
				PeefIntf:  fmt.Sprintf("sdneth%d", podIntfMap[edge.To]),
				LocalIP:   util.GetVxlanIP(uint(edge.From), uint(edge.To)),
				PeerIP:    util.GetVxlanIP(uint(edge.To), uint(edge.From)),
			},
		)
		topologyList.Items[edge.To].Spec.Links = append(
			topologyList.Items[edge.To].Spec.Links,
			util.Link{
				UID:       (edge.From << 12) + edge.To,
				PeerPod:   nameMap[edge.From],
				LocalIntf: fmt.Sprintf("sdneth%d", podIntfMap[edge.To]),
				PeefIntf:  fmt.Sprintf("sdneth%d", podIntfMap[edge.From]),
				LocalIP:   util.GetVxlanIP(uint(edge.To), uint(edge.From)),
				PeerIP:    util.GetVxlanIP(uint(edge.From), uint(edge.To)),
			},
		)
		podIntfMap[edge.From]++
		podIntfMap[edge.To]++
	}

	// Write to file
	conf, err := yaml.Marshal(topologyList)
	if err != nil {
		return fmt.Errorf("Error in parsing linkList to yaml: %v\n", err)
	}
	ioutil.WriteFile(outputPath, conf, 0644)

	return nil
}
