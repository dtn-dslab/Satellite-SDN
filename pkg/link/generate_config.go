package link

import (
	"fmt"
	"os/exec"
	"strings"
	"io/ioutil"

	
	"gopkg.in/yaml.v3"
	"ws/dtn-satellite-sdn/pkg/satellite"
)

func GeneratePodSummaryFile(nameMap map[int]string, edgeSet []satellite.LinkEdge, outputPath string, expectedNodeNum int) error {
	podList := PodList{}
	podList.Kind = "PodList"
	podList.APIVersion = "v1"

	// Get available nodes
	nodes := []string{}
	cmd := exec.Command("kubectl", "get", "nodes")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Executing %v failed: %v", cmd, err)
	}
	lines := strings.Split(string(output), "\n")
	lines = lines[1:]
	for _, line := range lines {
		blocks := strings.Split(string(line), " ")
		if !strings.Contains(blocks[0], "master") {
			nodes = append(nodes, blocks[0])
		}
	}

	// Consturct partitionMap
	nodeSet := []int{}
	for idx := 0; idx < len(nameMap); idx++ {
		nodeSet = append(nodeSet, idx)
	}
	nodeMap := map[int]string{}	// map: satId -> node
	partitions := GraphCutLinear(nodeSet, edgeSet, expectedNodeNum)
	for nodeId, partition := range partitions {
		for _, satId := range partition {
			nodeMap[satId] = nodes[nodeId]
		}
	}

	// Construct pods
	for idx := 0; idx < len(nameMap); idx++ {
		pod := Pod{
			APIVersion: "v1",
			Kind: "Pod",
			MetaData: MetaData{
				Name: nameMap[idx],
			},
			Spec: PodSpec{
				Containers: []Container{
					{
						Name: "satellite",
						Image: "golang:latest",
						ImagePullPolicy: "IfNotPresent",
						Ports: []ContainerPort{
							{
								ContainerPort: 8080,
							},
						},
						Command: []string{
							"/bin/sh",
							"-c",
						},
						Args: []string{
							"sleep 2000000000000",
						},
					},
				},
				NodeSelector: map[string]string{
					"kubernetes.io/hostname": nodeMap[idx],
				},
			},
		}
		podList.Items = append(podList.Items, pod)
	}

	// Write to file
	conf, err := yaml.Marshal(podList)
	if err != nil {
		return fmt.Errorf("Error in parsing podList to yaml: %v\n", err)
	}
	ioutil.WriteFile(outputPath, conf, 0644)

	return nil
}

func GenerateLinkSummaryFile(nameMap map[int]string, edgeSet []satellite.LinkEdge, outputPath string) error {
	// Initialize topologyList
	topologyList := TopologyList{}
	topologyList.APIVersion = "v1"
	topologyList.Kind = "List"
	for idx := 0; idx < len(nameMap); idx++ {
		topologyList.Items = append(topologyList.Items, Topology{
			APIVersion: "y-young.github.io/v1",
			Kind: "Topology",
			MetaData: MetaData{
				Name: nameMap[idx],
			},
		})
	}

	// Construct topologyList according to edgeSet
	for _, edge := range edgeSet {
		topologyList.Items[edge.From].Spec.Links = append(
			topologyList.Items[edge.From].Spec.Links,
			Link{
				UID: (edge.From << 12) + edge.To,
				PeerPod: nameMap[edge.To],
				LocalIntf: nameMap[edge.From] + "-" + nameMap[edge.To],
				PeefIntf: nameMap[edge.To] + "-" + nameMap[edge.From],
				LocalIP: GenerateIP(uint(edge.From), uint(edge.To), true),
				PeerIP: GenerateIP(uint(edge.From), uint(edge.To), false),
			}, 
		)
		topologyList.Items[edge.To].Spec.Links = append(
			topologyList.Items[edge.To].Spec.Links, 
			Link{
				UID: (edge.From << 12) + edge.To,
				PeerPod: nameMap[edge.From],
				LocalIntf: nameMap[edge.To] + "-" + nameMap[edge.From],
				PeefIntf: nameMap[edge.From] + "-" + nameMap[edge.To],
				LocalIP: GenerateIP(uint(edge.From), uint(edge.To), false),
				PeerIP: GenerateIP(uint(edge.From), uint(edge.To), true),
			},
		)
	}

	// Write to file
	conf, err := yaml.Marshal(topologyList)
	if err != nil {
		return fmt.Errorf("Error in parsing linkList to yaml: %v\n", err)
	}
	ioutil.WriteFile(outputPath, conf, 0644)

	return nil
}

// IP = uid << 8 | 0x1/0x2 (according to isLower) 
// The highest bit of IP must be 1
// Can support allocating IP to at most 2048 pods
func GenerateIP(lowerId, higherId uint, isLower bool) string {
	uid := (lowerId << 12) + higherId
	netIP := make([]string, 4)
	netIP[0] = fmt.Sprint((uid >> 16) & 0xff | 0x80)
	netIP[1] = fmt.Sprint((uid >> 8) & 0xff)
	netIP[2] = fmt.Sprint(uid & 0xff)
	if isLower {
		netIP[3] = "1"
	} else {
		netIP[3] = "2"
	}
	
	return strings.Join(netIP, ".") + "/24"
}