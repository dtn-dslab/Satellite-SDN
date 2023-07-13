package partition

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v3"
	"ws/dtn-satellite-sdn/pkg/link"
	"ws/dtn-satellite-sdn/pkg/util"
)



func GeneratePodSummaryFile(nameMap map[int]string, edgeSet []link.LinkEdge, outputPath string, expectedNodeNum int) error {
	podList := util.PodList{}
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
	nodeMap := map[int]string{} // map: satId -> node
	partitions := GraphCutLinear(nodeSet, edgeSet, expectedNodeNum)
	for nodeId, partition := range partitions {
		for _, satId := range partition {
			nodeMap[satId] = nodes[nodeId]
		}
	}

	// Construct pods
	for idx := 0; idx < len(nameMap); idx++ {
		pod := util.Pod{
			APIVersion: "v1",
			Kind:       "Pod",
			MetaData: util.MetaData{
				Name: nameMap[idx],
			},
			Spec: util.PodSpec{
				Containers: []util.Container{
					{
						Name:            "satellite",
						Image:           "electronicwaste/podserver:v3",
						ImagePullPolicy: "IfNotPresent",
						Ports: []util.ContainerPort{
							{
								ContainerPort: 8080,
							},
						},
						SecurityContext: util.SecurityContext{
							Capabilities: util.Capabilities{
								Add: []util.Capability{
									"NET_ADMIN",
								},
							},
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