package link

import (
	"context"
	"fmt"
	"math"
	"sort"
	"ws/dtn-satellite-sdn/sdn/util"

	satv1 "ws/dtn-satellite-sdn/sdn/type/v1"
	topov1 "github.com/y-young/kube-dtn/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const orbitAltDelta float64 = 500

type LinkEdge struct {
	From int
	To   int
}

func LinkSyncLoop(nameMap map[int]string, edgeSet []LinkEdge) error {
	// Initialize topologyList
	topoList := topov1.TopologyList{}
	podIntfMap := make([]int, len(nameMap))
	for idx := 0; idx < len(nameMap); idx++ {
		podIntfMap[idx] = 1
		topoList.Items = append(topoList.Items, topov1.Topology{
			ObjectMeta: metav1.ObjectMeta{
				Name: nameMap[idx],
			},
		})
	}

	// Construct topologyList according to edgeSet
	for _, edge := range edgeSet {
		topoList.Items[edge.From].Spec.Links = append(
			topoList.Items[edge.From].Spec.Links,
			topov1.Link{
				UID:       (edge.From << 12) + edge.To,
				PeerPod:   nameMap[edge.To],
				LocalIntf: fmt.Sprintf("sdneth%d", podIntfMap[edge.From]),
				PeerIntf:  fmt.Sprintf("sdneth%d", podIntfMap[edge.To]),
				LocalIP:   util.GetVxlanIP(uint(edge.From), uint(edge.To)),
				PeerIP:    util.GetVxlanIP(uint(edge.To), uint(edge.From)),
			},
		)
		topoList.Items[edge.To].Spec.Links = append(
			topoList.Items[edge.To].Spec.Links, 
			topov1.Link{
				UID:       (edge.From << 12) + edge.To,
				PeerPod:   nameMap[edge.From],
				LocalIntf: fmt.Sprintf("sdneth%d", podIntfMap[edge.To]),
				PeerIntf:  fmt.Sprintf("sdneth%d", podIntfMap[edge.From]),
				LocalIP:   util.GetVxlanIP(uint(edge.To), uint(edge.From)),
				PeerIP:    util.GetVxlanIP(uint(edge.From), uint(edge.To)),
			},
		)
		podIntfMap[edge.From]++
		podIntfMap[edge.To]++
	}

	// Get current namespace
	namespace, err := util.GetNamespace()
	if err != nil {
		return fmt.Errorf("GET NAMESPACE ERROR: %v", err)
	}

	// Create topologyList with RESTClient
	restClient, err := util.GetTopoClient()
	if err != nil {
		return fmt.Errorf("CONFIG ERROR: %v", err)
	}
	for _, topo := range topoList.Items {
		if err := restClient.Post().
			Namespace(namespace).
			Resource("topologies").
			Body(&topo).
			Do(context.TODO()).
			Into(nil); err != nil {
			return fmt.Errorf("APPLY TOPOLOGY FAILURE: %v", err)
		}
	}
	return nil
}

func isConnection(c *satv1.Constellation, sat1, sat2 string) (bool, error) {
	// Find corresponding satellites
	satellite1, err := c.FindSatelliteByName(sat1)
	if err != nil {
		return false, fmt.Errorf("The first satellite is not in the constellation\n")
	}
	satellite2, err := c.FindSatelliteByName(sat2)
	if err != nil {
		return false, fmt.Errorf("The second satellite is not in the constellation\n")
	}

	// Satellites connect with at most 4 neighbour satellites
	// 1. Same orbit & Closest satellite in positive direction
	// 2. Same orbit & Closest satellite in negative direction
	// 3. Higher orbit & Closest satellite
	// 4. Lower orbit & Closet satellite
	// p.s. If two satellites are in same orbit, we judge their distance with angleDelta.
	//		Otherwise, we judge in physical distance.
	distancesMap := map[string]float64{}
	_, _, alt1 := satellite1.Location()
	_, _, alt2 := satellite2.Location()
	angleDelta := satellite1.AngleDelta(satellite2)
	isSameOrbit := math.Abs((alt1-alt2)/orbitAltDelta) < 1
	isHigher := alt2-alt1 > 0
	isPostiveDir := angleDelta > 0
	if isSameOrbit && isPostiveDir {
		for _, sat := range c.Satellites {
			_, _, tmpAlt := sat.Location()
			if sat.Name != sat1 && math.Abs((tmpAlt-alt1)/orbitAltDelta) < 1 {
				distancesMap[sat.Name] = satellite1.AngleDelta(sat)
			}
		}
	} else if isSameOrbit && !isPostiveDir {
		for _, sat := range c.Satellites {
			_, _, tmpAlt := sat.Location()
			if sat.Name != sat1 && math.Abs((tmpAlt-alt1)/orbitAltDelta) < 1 {
				distancesMap[sat.Name] = -satellite1.AngleDelta(sat)
			}
		}
	} else if isHigher {
		for _, sat := range c.Satellites {
			_, _, tmpAlt := sat.Location()
			if sat.Name != sat2 && (tmpAlt-alt2)/orbitAltDelta <= -1 {
				distance, err := c.Distance(sat.Name, sat2)
				if err != nil {
					return false, fmt.Errorf("Calculating distance error in function isConnection\n")
				}
				distancesMap[sat.Name] = distance
			}
		}
	} else {
		for _, sat := range c.Satellites {
			_, _, tmpAlt := sat.Location()
			if sat.Name != sat1 && (tmpAlt-alt1)/orbitAltDelta <= -1 {
				distance, err := c.Distance(sat.Name, sat1)
				if err != nil {
					return false, fmt.Errorf("Calculating distance error in function isConnection\n")
				}
				distancesMap[sat.Name] = distance
			}
		}
	}

	// Sort distance from low to high
	type Pair struct {
		name     string
		distance float64
	}
	pairList := []Pair{}
	for k, v := range distancesMap {
		pairList = append(pairList, Pair{k, v})
	}
	sort.Slice(pairList, func(i, j int) bool {
		return pairList[i].distance < pairList[j].distance
	})
	if !isSameOrbit && isHigher && pairList[0].name == sat1 {
		return true, nil
	} else if pairList[0].name == sat2 {
		return true, nil
	} else {
		return false, nil
	}
}

// Return and connGraph
func GenerateConnGraph(c *satv1.Constellation) ([][]int) {
	// Initialize nameMap and connGraph
	nodeCount := len(c.Satellites)
	connGraph := [][]int{}
	for i := 0; i < nodeCount; i++ {
		tmpArr := []int{}
		for j := 0; j < nodeCount; j++ {
			tmpArr = append(tmpArr, 0)
		}
		connGraph = append(connGraph, tmpArr)
	}

	// Construct connGraph
	for idx1, sat1 := range c.Satellites {
		nextMinVal, nextMinIdx := 1e9, -1
		prevMinVal, prevMinIdx := 1e9, -1
		lowerMinVal, lowerMinIdx := 1e9, -1

		for idx2, sat2 := range c.Satellites {
			if idx1 != idx2 {
				_, _, alt1 := sat1.Location()
				_, _, alt2 := sat2.Location()
				angleDelta := sat1.AngleDelta(sat2)
				isSameOrbit := math.Abs((alt1-alt2)/orbitAltDelta) < 1
				isHigher := alt2-alt1 > 0
				isPostiveDir := angleDelta > 0

				if isSameOrbit && isPostiveDir {
					if math.Abs(angleDelta) < nextMinVal {
						nextMinVal, nextMinIdx = math.Abs(angleDelta), idx2
					}
				} else if isSameOrbit && !isPostiveDir {
					if math.Abs(angleDelta) < prevMinVal {
						prevMinVal, prevMinIdx = math.Abs(angleDelta), idx2
					}
				} else if !isHigher {
					if sat1.Distance(sat2) < lowerMinVal {
						lowerMinVal, lowerMinIdx = sat1.Distance(sat2), idx2
					}
				}
			}
		}

		if nextMinIdx != -1 {
			connGraph[idx1][nextMinIdx] = 1
		}
		if prevMinIdx != -1 {
			connGraph[idx1][prevMinIdx] = 1
		}
		if lowerMinIdx != -1 {
			connGraph[idx1][lowerMinIdx] = 1
			connGraph[lowerMinIdx][idx1] = 1
		}
	}

	return connGraph
}

func GenerateDistanceMap(c *satv1.Constellation, connGraph [][]int) [][]float64 {
	// Initialize distanceMap (distance is 1e9 at first)
	distanceMap := [][]float64{}
	for i := 0; i < len(connGraph); i++ {
		arr := []float64{}
		for j := 0; j < len(connGraph); j++ {
			arr = append(arr, 1e9)
		}
		distanceMap = append(distanceMap, arr)
	}

	// Compute distance
	for idx1, sat1 := range c.Satellites {
		for idx2, sat2 := range c.Satellites {
			if connGraph[idx1][idx2] == 1 {
				distanceMap[idx1][idx2] = sat1.Distance(sat2)
			} else if idx1 == idx2 {
				distanceMap[idx1][idx2] = 0
			}
		}
	}

	return distanceMap
}

func ConvertConnGraphToEdgeSet(connGraph [][]int) []LinkEdge {
	edgeSet := []LinkEdge{}
	for idx1 := 0; idx1 < len(connGraph); idx1++ {
		for idx2 := idx1 + 1; idx2 < len(connGraph); idx2++ {
			if connGraph[idx1][idx2] == 1 {
				edgeSet = append(edgeSet, LinkEdge{
					From: idx1, To: idx2,
				})
			}
		}
	}
	return edgeSet
}