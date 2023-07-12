package route

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os/exec"
	"strings"
	"time"
	"ws/dtn-satellite-sdn/pkg/link"
	"ws/dtn-satellite-sdn/pkg/util"

	"gopkg.in/yaml.v3"
)

func GenerateRouteSummaryFile(nameMap map[int]string, routeTable [][]int, outputPath string) error {
	// Get pods' ip
	nodeCount := len(nameMap)
	podIPTable := []string{}
	for idx := 0; idx < nodeCount; idx++ {
		var podIP string
		var err error
		for podIP, err = GetPodIP(nameMap[idx]); err != nil; podIP, err = GetPodIP(nameMap[idx]) {
			log.Println("Retry")
			duration := 3000 + rand.Int31()%2000
			time.Sleep(time.Duration(duration) * time.Millisecond)
		}
		podIPTable = append(podIPTable, podIP)
	}

	// Create routeList
	routeList := util.RouteList{}
	routeList.APIVersion = "v1"
	routeList.Kind = "List"
	for idx1 := range routeTable {
		route := util.Route{
			APIVersion: "sdn.dtn-satellite-sdn/v1",
			MetaData: util.MetaData{
				Name: nameMap[idx1],
			},
			Kind: "Route",
			Spec: util.RouteSpec{
				SubPaths: []util.SubPath{},
				PodIP:    podIPTable[idx1],
			},
		}
		for idx2 := range routeTable[idx1] {
			if routeTable[idx1][idx2] != -1 {
				route.Spec.SubPaths = append(
					route.Spec.SubPaths,
					util.SubPath{
						Name:     nameMap[idx2],
						TargetIP: link.GenerateIP(uint(idx2)),
						NextIP:   link.GenerateIP(uint(routeTable[idx1][idx2])),
					},
				)
			}
		}
		routeList.Items = append(routeList.Items, route)
	}

	// Write to file
	conf, err := yaml.Marshal(routeList)
	if err != nil {
		return fmt.Errorf("Error in parsing routeList to yaml: %v\n", err)
	}
	ioutil.WriteFile(outputPath, conf, 0644)

	return nil
}

// Get a pod's ip via 'kubectl get pod <podName> -o wide' instruction by parsing the output.
func GetPodIP(podName string) (string, error) {
	// Executing 'kubectl get pod <podName> -o wide'
	cmd := exec.Command("kubectl", "get", "pod", podName, "-o", "wide")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Executing kubectl get pods failed: %v\n", err)
	}

	// Get pod ip
	lines := strings.Split(string(output), "\n")
	if len(lines) <= 1 {
		return "", fmt.Errorf("Can't find pod: %s\n", podName)
	}
	blocks := strings.Split(lines[1], " ")
	newBlocks := []string{}
	for _, block := range blocks {
		if block != "" {
			newBlocks = append(newBlocks, block)
		}
	}
	if newBlocks[5] != "<none>" {
		return newBlocks[5], nil
	} else {
		return "", fmt.Errorf("Can't find pod: %s\n", podName)
	}

}
