package clientset

import (
	"sort"
	"time"

	satv2 "ws/dtn-satellite-sdn/sdn/type/v2"
)

type OrbitInterface interface {
	Update(params map[string]interface{}) *OrbitInfo
	RefreshMeta()
}

type OrbitMeta struct {
	// IndexUUIDMap stores index->UUID map
	IndexUUIDMap map[int]string

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

	// EnemyNum stores the number of enemies
	EnemyNum int
}

type OrbitInfo struct {
	// Metadata stores orbit metadata
	Metadata *OrbitMeta

	// TimeStamp is the last time this orbitInfo was updated
	TimeStamp time.Time	

	// LowOrbitSats stores groups of satellite nodes in low orbit
	LowOrbitSats map[int]*satv2.Group 

	// HighOrbitSats stores groups of satellite nodes in high orbit
	HighOrbitSats map[int]*satv2.Group	

	// GroundStations stores the group of ground station nodes
	GroundStations *satv2.Group	

	// Missiles stores the group of missile nodes
	Missiles *satv2.Group 

	// Enemies stores the group of enemy nodes
	Enemies *satv2.Group 
}

func NewOrbitInfo(params map[string]interface{}) *OrbitInfo {
	unixTimeStamp := params["unixTimeStamp"].(int64)
	satellites := params["satellites"].([]map[string]interface{})
	stations := params["stations"].([]map[string]interface{})
	missiles := params["missiles"].([]map[string]interface{})
	enemies := params["enemies"].([]map[string]interface{})
	info := OrbitInfo {
		TimeStamp: time.Unix(unixTimeStamp / 1000, 0),
		LowOrbitSats: make(map[int]*satv2.Group),
		HighOrbitSats: make(map[int]*satv2.Group),
		GroundStations: satv2.NewOtherGroup(satv2.GROUNDSTATION),
		Missiles: satv2.NewOtherGroup(satv2.MISSILE),
		Enemies: satv2.NewOtherGroup(satv2.ENEMY),
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
			trackID := node.TrackID
			node.Type = satv2.HIGHORBIT
			if _, ok := info.HighOrbitSats[trackID]; !ok {
				info.HighOrbitSats[trackID] = satv2.NewSatGroup(satv2.HIGHORBIT, trackID)
			}
			info.HighOrbitSats[trackID].Nodes = append(info.HighOrbitSats[trackID].Nodes, node)
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
	// Initialize enemy groups
	for _, enemy := range enemies {
		info.Enemies.Nodes = append(
			info.Enemies.Nodes, 
			satv2.NewOtherNode(satv2.ENEMY, enemy),
		)
	}
	// Update Metadata
	info.RefreshMeta()
	
	return &info
}

func (o *OrbitInfo) Update(params map[string]interface{}) *OrbitInfo {
	unixTimeStamp := params["unixTimeStamp"].(int64)
	satellites := params["satellites"].([]map[string]interface{})
	stations := params["stations"].([]map[string]interface{})
	missiles := params["missiles"].([]map[string]interface{})
	enemies := params["enemies"].([]map[string]interface{})
	// Initialize uuid->*Node map
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
	for _, enemy := range enemies {
		node := satv2.NewOtherNode(satv2.ENEMY, enemy)
		uuidNodeMap[node.UUID] = node
	}
	// Update node in o
	o.TimeStamp = time.Unix(unixTimeStamp / 1000, 0)
	for idx1, group := range o.LowOrbitSats {
		for idx2, node := range group.Nodes {
			o.LowOrbitSats[idx1].Nodes[idx2] = uuidNodeMap[node.UUID]
		}
	}
	for idx1, group := range o.HighOrbitSats {
		for idx2, node := range group.Nodes {
			o.HighOrbitSats[idx1].Nodes[idx2] = uuidNodeMap[node.UUID]
		}
	}
	for idx, station := range o.GroundStations.Nodes {
		o.GroundStations.Nodes[idx] = uuidNodeMap[station.UUID]
	}
	for idx, missile := range o.Missiles.Nodes {
		o.Missiles.Nodes[idx] = uuidNodeMap[missile.UUID]
	}
	for idx, enemy := range o.Enemies.Nodes {
		o.Enemies.Nodes[idx] = uuidNodeMap[enemy.UUID]
	}
	return o
}

func (o *OrbitInfo) RefreshMeta() {
	cur_idx := 0
	o.Metadata = &OrbitMeta{
		IndexUUIDMap: make(map[int]string),
		UUIDNodeMap: make(map[string]*satv2.Node),
		LowOrbitNum: 0,
		HighOrbitNum: 0,
		GroundStationNum: 0,
		MissileNum: 0,
		EnemyNum: 0,
	}
	for idx1, group := range o.LowOrbitSats {
		for idx2, node := range group.Nodes {
			o.Metadata.LowOrbitNum++
			o.Metadata.IndexUUIDMap[cur_idx] = node.UUID
			o.Metadata.UUIDNodeMap[node.UUID] = &o.LowOrbitSats[idx1].Nodes[idx2]
			cur_idx++
		}
	}
	for idx1, group := range o.HighOrbitSats {
		for idx2, node := range group.Nodes {
			o.Metadata.HighOrbitNum++
			o.Metadata.IndexUUIDMap[cur_idx] = node.UUID
			o.Metadata.UUIDNodeMap[node.UUID] = &o.HighOrbitSats[idx1].Nodes[idx2]
			cur_idx++
		}
	}
	for idx, node := range o.GroundStations.Nodes {
		o.Metadata.GroundStationNum++
		o.Metadata.IndexUUIDMap[cur_idx] = node.UUID
		o.Metadata.UUIDNodeMap[node.UUID] = &o.GroundStations.Nodes[idx]
		cur_idx++
	}
	for idx, node := range o.Missiles.Nodes {
		o.Metadata.MissileNum++
		o.Metadata.IndexUUIDMap[cur_idx] = node.UUID
		o.Metadata.UUIDNodeMap[node.UUID] = &o.Missiles.Nodes[idx]
		cur_idx++
	}
	for idx, node := range o.Enemies.Nodes {
		o.Metadata.EnemyNum++
		o.Metadata.IndexUUIDMap[cur_idx] = node.UUID
		o.Metadata.UUIDNodeMap[node.UUID] = &o.Enemies.Nodes[idx]
		cur_idx++
	}
}