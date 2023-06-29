package link

import (
	"fmt"
	"os/exec"
	"strings"
	"io/ioutil"

	
	"gopkg.in/yaml.v3"
	"ws/dtn-satellite-sdn/pkg/satellite"
)

func GeneratePodSummaryFile(c *satellite.Constellation, outputPath string, expectedNodeNum int) error {
	nameMap, edgeSet :=  c.GenerateEdgeSet()
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

// func GenerateLinkSummaryFile(c *satellite.Constellation, outputPath string) error {

// }