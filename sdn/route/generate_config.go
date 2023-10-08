package route

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"time"

	"ws/dtn-satellite-sdn/sdn/util"

	"gopkg.in/yaml.v3"
)

func GenerateRouteSummaryFile(nameMap map[int]string, routeTable [][]int, outputPath string) error {
	// Get pods' ip
	nodeCount := len(nameMap)
	podIPTable := []string{}
	for idx := 0; idx < nodeCount; idx++ {
		var podIP string
		var err error
		for podIP, err = util.GetPodIP(nameMap[idx]); err != nil; podIP, err = util.GetPodIP(nameMap[idx]) {
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
				// New routes for target Pod
				route.Spec.SubPaths = append(
					route.Spec.SubPaths,
					util.SubPath{
						Name:     nameMap[idx2],
						TargetIP: util.GetGlobalIP(uint(idx2)),
						NextIP:   util.GetVxlanIP(uint(routeTable[idx1][idx2]), uint(idx1)),
					},
				)
			} else if idx1 != idx2 {
				// Exising routes for target Pod, rewrite it with global IP
				route.Spec.SubPaths = append(
					route.Spec.SubPaths, 
					util.SubPath{
						Name: nameMap[idx2],
						TargetIP: util.GetGlobalIP(uint(idx2)),
						NextIP: util.GetVxlanIP(uint(idx2), uint(idx1)),
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
