package partition

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v3"
	"ws/dtn-satellite-sdn/sdn/link"
	"ws/dtn-satellite-sdn/sdn/util"
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
	partitions := GraphCutHash(nodeSet, expectedNodeNum)
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
						Image:           "electronicwaste/podserver:v6",
						ImagePullPolicy: "IfNotPresent",
						Ports: []util.ContainerPort{
							{
								ContainerPort: 8080,
							},
						},
						Command: []string {
							"/bin/sh",
							"-c",
						},
						Args: []string {
							fmt.Sprintf(
								"ifconfig eth0:sdneth0 %s netmask 255.255.255.255 up;" + 
								"iptables -t nat -A OUTPUT -d 10.233.0.0/16 -j MARK --set-mark %d;" +
								"iptables -t nat -A POSTROUTING -m mark --mark %d -d 10.233.0.0/16 -j SNAT --to-source %s;" +
								"/podserver", 
								util.GetGlobalIP(uint(idx)),
								idx + 5000,
								idx + 5000,
								util.GetGlobalIP(uint(idx)),
							),
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

func GraphCutHash(nodeSet []int, expectedSplitsCount int) [][]int {
	ret := make([][]int, expectedSplitsCount)

	for _, nodeId := range nodeSet {
		idx := nodeId % expectedSplitsCount
		ret[idx] = append(ret[idx], nodeId)
	}

	return ret
}

