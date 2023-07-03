package route

import "ws/dtn-satellite-sdn/pkg/link"

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
	// for idx1 := range routeTable {
	// 	for idx2 := range routeTable[idx1] {
	// 		if routeTable[idx1][idx2] != -1 {

	// 		}
	// 	}
	// }
	return nil
}