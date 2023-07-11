package route

import (
	"fmt"
	"io/ioutil"
	"ws/dtn-satellite-sdn/pkg/link"
	"ws/dtn-satellite-sdn/pkg/util"

	"gopkg.in/yaml.v3"
)

func GenerateRouteSummaryFile(nameMap map[int]string, routeTable [][]int, outputPath string) error {
	// Initialize & Create vxlanIPTable
	nodeCount := len(nameMap)
	vxlanIPTable := [][]string{}
	for i := 0; i < nodeCount; i++ {
		vxlanIPTable = append(vxlanIPTable, make([]string, nodeCount))
	}
	for i := 0; i < nodeCount; i++ {
		vxlanIPTable[i][i] = ""
		for j := i + 1; j < nodeCount; j++ {
			fromIP := link.GenerateIP(uint(i), uint(j), true)
			toIP := link.GenerateIP(uint(i), uint(j), false)
			vxlanIPTable[i][j] = fromIP
			vxlanIPTable[j][i] = toIP
		}
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
			},
		}
		for idx2 := range routeTable[idx1] {
			if routeTable[idx1][idx2] != -1 {
				route.Spec.SubPaths = append(
					route.Spec.SubPaths,
					util.SubPath{
						Name:     nameMap[idx2],
						TargetIP: vxlanIPTable[idx1][idx2],
						NextIP:   vxlanIPTable[idx1][routeTable[idx1][idx2]],
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
