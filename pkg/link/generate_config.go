package link

import (
	"fmt"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v3"
	"ws/dtn-satellite-sdn/pkg/util"
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
				LocalIP:   GenerateIP(uint(edge.From)),
				PeerIP:    GenerateIP(uint(edge.To)),
			},
		)
		topologyList.Items[edge.To].Spec.Links = append(
			topologyList.Items[edge.To].Spec.Links,
			util.Link{
				UID:       (edge.From << 12) + edge.To,
				PeerPod:   nameMap[edge.From],
				LocalIntf: fmt.Sprintf("sdneth%d", podIntfMap[edge.To]),
				PeefIntf:  fmt.Sprintf("sdneth%d", podIntfMap[edge.From]),
				LocalIP:   GenerateIP(uint(edge.To)),
				PeerIP:    GenerateIP(uint(edge.From)),
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

// IP = 0x80000000 | dst << 12 | src
// The highest bit of IP must be 1
// Can support allocating IP to at most 2 ^ 16 pods
func GenerateIP(id uint) string {
	id = id + 1
	netIP := make([]string, 4)
	netIP[0] = fmt.Sprint((id >> 24) & 0xff | 0x80)
	netIP[1] = fmt.Sprint((id >> 16) & 0xff)
	netIP[2] = fmt.Sprint((id >> 8) & 0xff)
	netIP[3] = fmt.Sprint(id & 0xff)

	return strings.Join(netIP, ".") + "/24"
}
