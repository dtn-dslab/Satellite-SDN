package link

import (
	"ws/dtn-satellite-sdn/sdn/satellite"
)

// Function: GetSatConnectedToGroundStation
// Description: Given ground station, return satellite most closest to it.
// 1. s: The ground station.
// 2. cs: All constellations in StarNet
func GetSatConnectedToGroundStation(s satellite.GroundStation, cs []satellite.Constellation) satellite.Satellite {
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

// Function: GetGraphInConstellation
// Description: Return connection graph for satellites in the same orbit
// Return type: UUID -> []UUID
// 1. idx: Current constellation's index in cs
// 2. cs: All constellations in StarNet
func GetGraphInConstellation(idx int, cs []satellite.Constellation) map[string][]string {
	// Get edge set
	c := cs[idx]
	result := map[string][]string{}
	for i := 0; i <= c.Len(); i++ {
		result[c.Satellites[i].UUID] = append(
			result[c.Satellites[i].UUID], 
			c.Satellites[(i + 1) % c.Len()].UUID,
		)
		result[c.Satellites[(i + 1) % c.Len()].UUID] = append(
			result[c.Satellites[(i + 1) % c.Len()].UUID], 
			c.Satellites[i].UUID,
		)
	}
	return result
}

// Function: GetGraphAcrossConstellation
// Description: Return connection graph between different constellations.
// Return type: UUID -> []UUID
// 1. idx: Current constellation's index in cs
// 2. cs: All constellations in StarNet
// 3. name_map: UUID->int map in StarNet
// 4. distance_graph: distance between satellites(get index by name_map[UUID])
func GetGraphAcrossConstellation(
	idx int, cs []satellite.Constellation, distance_graph [][]float64,
	index_map map[string]int, uuid_map map[int]string) map[string][]string {
	// Initialize some variables
	result := map[string][]string{}
	c := cs[idx]
	current_idxs := []int{}
	for _, s := range c.Satellites {
		current_idxs = append(current_idxs, index_map[s.UUID])
	}
	
	// Get 2 nearest satellites for each satellite in current constellation
	for _, s := range c.Satellites {
		sat_id := index_map[s.UUID]
		// Get the nearst satellite
		min_distance, min_idx := 1e9, -1
		for other_sat_id, distance := range distance_graph[sat_id] {
			// If other_sat_id is in current_idxs, continue
			flag := false
			for _, index := range current_idxs {
				if other_sat_id == index { 
					flag = true
					break
				}
			}
			if flag { continue }
			// Update min_distance & min_idx
			if distance < min_distance {
				min_distance = distance
				min_idx = other_sat_id
			}
		}
		// Get the second nearst satellite
		second_min_distance, second_min_idx := 1e9, -1
		for other_sat_id, distance := range distance_graph[sat_id] {
			// If other_sat_id is in current_idxs or equals to min_idx, continue
			flag := false
			for _, index := range current_idxs {
				if other_sat_id == index || other_sat_id == min_idx {
					flag = true
					break
				}
			}
			if flag { continue }
			// Update second_min_distance & second_min_idx
			if distance < second_min_distance {
				second_min_distance = distance
				second_min_idx = other_sat_id
			}
		}
		// Update connection graph
		result[s.UUID] = []string{uuid_map[min_idx], uuid_map[second_min_idx]}
	}

	return result
}