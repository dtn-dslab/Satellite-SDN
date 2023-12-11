package clientset

import (
	"sort"
	"time"

	satv2 "ws/dtn-satellite-sdn/sdn/type/v2"
)

type OrbitInterface interface {
	Update(params map[string]interface{})
	UpdateMeta()
	GetUUIDIndexMap() map[string]int
	GetIndexUUIDMap() map[int]string
}

type OrbitMeta struct {
	// TimeStamp is the last time this orbitInfo was updated
	TimeStamp time.Time	

	// IndexUUIDMap stores index->UUID map
	IndexUUIDMap map[int]string

	// UUIDIndexMap stores UUID->index map
	UUIDIndexMap map[string]int

	// UUIDNodeMap stores UUID->Node map
	UUIDNodeMap map[string]*satv2.Node

	// LowOrbitNum stores the number of low-orbit satellites
	LowOrbitNum int

	// HighOrbitNum stores the number of high-orbit satellites
	HighOrbitNum int

	// GroundStationNum stores the number of ground stations
	GroundStationNum int

	// MissileNum stores the number of missiles
	MissileNum int
}

type OrbitInfo struct {
	// Metadata stores orbit metadata
	Metadata *OrbitMeta

	// LowOrbitSats stores groups of satellite nodes in low orbit
	LowOrbitSats map[int]*satv2.Group 

	// HighOrbitSats stores groups of satellite nodes in high orbit
	HighOrbitSats map[int]*satv2.Group	

	// GroundStations stores the group of ground station nodes
	GroundStations *satv2.Group	

	// Missiles stores the group of missile nodes
	Missiles *satv2.Group 
}

// Function: ParseParamsQimeng
// Description: Parsing params in http response for Qimeng
// 1. param: Message from Qimeng
func ParseParamsQimeng(params map[string]interface{}) (unixTimeStamp int64, satellites, stations, missiles []map[string]interface{}) {
	unixTimeStamp = (int64) (params["unixTimeStamp"].(float64))
	if params["satellites"] != nil {
		for _, intf := range params["satellites"].([]interface{}) {
			satellites = append(satellites, intf.(map[string]interface{}))
		}
	}
	if params["stations"] != nil {
		for _, intf := range params["stations"].([]interface{}) {
			stations = append(stations, intf.(map[string]interface{}))
		}
	}
	if params["missiles"] != nil {
		for _, intf := range params["missiles"].([]interface{}) {
			missiles = append(missiles, intf.(map[string]interface{}))
		}
	}
	return
}

// Function: NewOrbitInfo
// Description: Create orbit info with JSON params
// 1. params: Message from Qimeng
func NewOrbitInfo(params map[string]interface{}) *OrbitInfo {
	unixTimeStamp, satellites, stations, missiles := ParseParamsQimeng(params)
	info := OrbitInfo {
		LowOrbitSats: make(map[int]*satv2.Group),
		HighOrbitSats: make(map[int]*satv2.Group),
		GroundStations: satv2.NewOtherGroup(satv2.GROUNDSTATION),
		Missiles: satv2.NewOtherGroup(satv2.MISSILE),
		Metadata: &OrbitMeta{},
	}
	// Initialize low-orbit and high-orbit satellite groups
	for _, sat := range satellites {
		node := satv2.NewSatNode(satv2.LOWORBIT, sat)
		// Classify satellites by altitude 30000km
		if (node.Altitude < 30000) {
			trackID := node.TrackID
			if _, ok := info.LowOrbitSats[trackID]; !ok {
				info.LowOrbitSats[trackID] = satv2.NewSatGroup(satv2.LOWORBIT, trackID)
			}
			info.LowOrbitSats[trackID].Nodes = append(info.LowOrbitSats[trackID].Nodes, node)
		} else {
			// trackID := node.TrackID
			// node.Type = satv2.HIGHORBIT
			// if _, ok := info.HighOrbitSats[trackID]; !ok {
			// 	info.HighOrbitSats[trackID] = satv2.NewSatGroup(satv2.HIGHORBIT, trackID)
			// }
			// info.HighOrbitSats[trackID].Nodes = append(info.HighOrbitSats[trackID].Nodes, node)
		}
	}
	for _, group := range info.LowOrbitSats {
		sort.Sort(group)
	}
	for _, group := range info.HighOrbitSats {
		sort.Sort(group)
	}
	// Initialzie station groups
	for _, station := range stations {
		info.GroundStations.Nodes = append(
			info.GroundStations.Nodes, 
			satv2.NewOtherNode(satv2.GROUNDSTATION, station),
		)
	}
	// Initialize missile groups
	for _, missile := range missiles {
		info.Missiles.Nodes = append(
			info.Missiles.Nodes, 
			satv2.NewOtherNode(satv2.MISSILE, missile),
		)
	}
	// Update Metadata
	info.Metadata = info.UpdateMeta(unixTimeStamp)
	
	return &info
}

// Function: Update
// Description: Update orbit info with JSON params, also timestamp & uuidNodeMap in orbit metadata.
// 1. params: Message from Qimeng
func (o *OrbitInfo) Update(params map[string]interface{}) {
	unixTimeStamp, satellites, stations, missiles := ParseParamsQimeng(params)
	// Initialize uuid->Node map
	uuidNodeMap := make(map[string]satv2.Node)
	for _, sat := range satellites {
		node := satv2.NewSatNode(satv2.LOWORBIT, sat)
		// Classify satellites by altitude 30000km
		if (node.Altitude < 30000) {
			uuidNodeMap[node.UUID] = node
		} else {
			node.Type = satv2.HIGHORBIT
			uuidNodeMap[node.UUID] = node
		}
	}
	for _, station := range stations {
		node := satv2.NewOtherNode(satv2.GROUNDSTATION, station)
		uuidNodeMap[node.UUID] = node
	}
	for _, missile := range missiles {
		node := satv2.NewOtherNode(satv2.MISSILE, missile)
		uuidNodeMap[node.UUID] = node
	}
	// Update variables in o.Metadata
	o.Metadata.TimeStamp = time.UnixMilli(unixTimeStamp)
	// Update node and uuidNodeMap in o
	for idx1, group := range o.LowOrbitSats {
		for idx2, node := range group.Nodes {
			o.LowOrbitSats[idx1].Nodes[idx2] = uuidNodeMap[node.UUID]
			o.Metadata.UUIDNodeMap[node.UUID] = &o.LowOrbitSats[idx1].Nodes[idx2]
		}
	}
	for idx1, group := range o.HighOrbitSats {
		for idx2, node := range group.Nodes {
			o.HighOrbitSats[idx1].Nodes[idx2] = uuidNodeMap[node.UUID]
			o.Metadata.UUIDNodeMap[node.UUID] = &o.HighOrbitSats[idx1].Nodes[idx2]
		}
	}
	for idx, station := range o.GroundStations.Nodes {
		o.GroundStations.Nodes[idx] = uuidNodeMap[station.UUID]
		o.Metadata.UUIDNodeMap[station.UUID] = &o.GroundStations.Nodes[idx]
	}
	for idx, missile := range o.Missiles.Nodes {
		o.Missiles.Nodes[idx] = uuidNodeMap[missile.UUID]
		o.Metadata.UUIDNodeMap[missile.UUID] = &o.Missiles.Nodes[idx]
	}
}

// Function: UpdateMeta
// Description: Update metadata when orbit info is created
// 1. unixTimeStamp: the timestamp passed by Qimeng
func (o *OrbitInfo) UpdateMeta(unixTimeStamp int64) *OrbitMeta {
	cur_idx := 0
	meta := OrbitMeta{
		TimeStamp: time.UnixMilli(unixTimeStamp),
		IndexUUIDMap: make(map[int]string),
		UUIDIndexMap: make(map[string]int),
		UUIDNodeMap: make(map[string]*satv2.Node),
		LowOrbitNum: 0,
		HighOrbitNum: 0,
		GroundStationNum: 0,
		MissileNum: 0,
	}
	// Since we do not modify bucket, we can get same result each time we iterate map.
	for idx1, group := range o.LowOrbitSats {
		for idx2, node := range group.Nodes {
			meta.LowOrbitNum++
			meta.IndexUUIDMap[cur_idx] = node.UUID
			meta.UUIDIndexMap[node.UUID] = cur_idx
			meta.UUIDNodeMap[node.UUID] = &o.LowOrbitSats[idx1].Nodes[idx2]
			cur_idx++
		}
	}
	for idx1, group := range o.HighOrbitSats {
		for idx2, node := range group.Nodes {
			meta.HighOrbitNum++
			meta.IndexUUIDMap[cur_idx] = node.UUID
			meta.UUIDIndexMap[node.UUID] = cur_idx
			meta.UUIDNodeMap[node.UUID] = &o.HighOrbitSats[idx1].Nodes[idx2]
			cur_idx++
		}
	}
	for idx, node := range o.GroundStations.Nodes {
		meta.GroundStationNum++
		meta.IndexUUIDMap[cur_idx] = node.UUID
		meta.UUIDIndexMap[node.UUID] = cur_idx
		meta.UUIDNodeMap[node.UUID] = &o.GroundStations.Nodes[idx]
		cur_idx++
	}
	for idx, node := range o.Missiles.Nodes {
		meta.MissileNum++
		meta.IndexUUIDMap[cur_idx] = node.UUID
		meta.UUIDIndexMap[node.UUID] = cur_idx
		meta.UUIDNodeMap[node.UUID] = &o.Missiles.Nodes[idx]
		cur_idx++
	}
	return &meta
}

func (o *OrbitInfo) GetIndexUUIDMap() map[int]string {
	return o.Metadata.IndexUUIDMap
}

func (o *OrbitInfo) GetUUIDIndexMap() map[string]int {
	return o.Metadata.UUIDIndexMap
}