package link

import (
	"sort"
	"ws/dtn-satellite-sdn/sdn/satellite"
)

func GetSatConnectedToGroundStation(s satellite.GroundStation, cs []satellite.Constellation) satellite.Satellite{
	min_distance, res := 1e9, satellite.Satellite{}
	for _, c := range cs {
		for _, sat := range c.Satellites {
			if distance := s.DistanceWithSat(&sat); distance < min_distance {
				res = sat
				min_distance = distance
			}
		}
	}
	return res
}

func GetGraphInConstellation(c satellite.Constellation) [][]string {
	// Sort satellites in angle order
	sort.Sort(c)

	// Get edge set
	result := [][]string{}
	for i := 0; i <= c.Len(); i++ {
		result = append(result, []string{
			c.Satellites[i].Name,
			c.Satellites[(i + 1) % c.Len()].Name,
		})
	}
	return result
}