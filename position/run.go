package position

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	sdnv1 "ws/dtn-satellite-sdn/sdn/type/v1"
)

type PositionServerInterface interface {
	GetLocationHanlder(http.ResponseWriter, *http.Request)
	ComputeSatsCache()
	UpdateCache() error
}

type PositionServer struct {
	c *sdnv1.Constellation
	cache *PositionCache
	fixedNum int
	timeStamp time.Time
}

type PositionCache struct {
	satCache map[string]*SatParams
	msCache []MSParams
	gsCache []GSParams
	fixedCache []FixedParams
}

func NewPositionServer(inputPath string, num int) *PositionServer {
	if constellation, err := sdnv1.NewConstellation(inputPath); err != nil {
		panic(err)
	} else {
		return &PositionServer{
			c: constellation,
			cache: nil,
			fixedNum: num,
		}
	}
}

// Function: GetLocationHanlder
// Description: A http hanlder for getting location of all types of node.
func (ps *PositionServer) GetLocationHandler(w http.ResponseWriter, req *http.Request) {
	ps.timeStamp = time.Now()
	if err := ps.UpdateCache(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	retParams := RetParams{
		TimeStamp: ps.timeStamp.UnixMilli(),
		Satellites: []SatParams{},
		Missiles: ps.cache.msCache,
		Stations: ps.cache.gsCache,
		FixedNodes: ps.cache.fixedCache,
	}
	for _, sat := range ps.cache.satCache {
		retParams.Satellites = append(retParams.Satellites, *sat)
	}
	fmt.Println(retParams)
	content, _ := json.Marshal(&retParams)
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

// Function: UpdateCache
// Description: Update cache in ps.cache for future use.
func (ps *PositionServer) UpdateCache() error {
	if ps.cache == nil {
		ps.cache = &PositionCache{
			satCache: make(map[string]*SatParams),
			msCache: GetMissiles(),
			gsCache: GetGroundStation(),
			fixedCache: GetFixedNodes(ps.fixedNum),
		}
		ps.ComputeSatsCache()
	} else {
		// Update longitude, latitude, altitude in Satellites
		year, month, day, hour, minute, second :=
			ps.timeStamp.Year(),
			int(ps.timeStamp.Month()),
			ps.timeStamp.Day(),
			ps.timeStamp.Hour(),
			ps.timeStamp.Minute(),
			ps.timeStamp.Second()
		for _, sat := range ps.c.Satellites {
			long, lat, alt := sat.LocationAtTime(
				year, month, day,
				hour, minute, second,
			)
			ps.cache.satCache[sat.Name].Longitude = long
			ps.cache.satCache[sat.Name].Latitude = lat
			ps.cache.satCache[sat.Name].Altitude = alt
		}
	}
	return nil
}

// Function: ComputeSatsCache
// Description: Compute all of sats' information when cache is recently created.
func (ps *PositionServer) ComputeSatsCache() {
	// Initialze satCache
	year, month, day, hour, minute, second :=
			ps.timeStamp.Year(),
			int(ps.timeStamp.Month()),
			ps.timeStamp.Day(),
			ps.timeStamp.Hour(),
			ps.timeStamp.Minute(),
			ps.timeStamp.Second()
	for _, sat := range ps.c.Satellites {
		long, lat, alt := sat.LocationAtTime(
			year, month, day,
			hour, minute, second,
		)
		s := SatParams{
			UUID: sat.Name,
			Longitude: long,
			Latitude: lat,
			Altitude: alt,
		}
		ps.cache.satCache[sat.Name] = &s
	}
	// Classify Satellites
	curTrackID, remainNode := 0, len(ps.cache.satCache)
	classifySatsUUIDList := [][]string{}
	visited := make(map[string]bool, remainNode)
	for k, _ := range ps.cache.satCache {
		visited[k] = false
	}
	for remainNode > 0 {
		var standardHeight float64
		// Find standard height for this track
		for key, sat := range ps.cache.satCache {
			if !visited[key] {
				standardHeight = sat.Altitude
				break
			}
		}
		// Iterate to find sats in the same track(|height - standardHeight| < 500)
		for key, sat := range ps.cache.satCache {
			if !visited[key] && math.Abs(sat.Altitude - standardHeight) < 500 {
				if len(classifySatsUUIDList) <= curTrackID {
					classifySatsUUIDList = append(classifySatsUUIDList, []string{})
				}
				classifySatsUUIDList[curTrackID] = append(classifySatsUUIDList[curTrackID], key) // Update result
				visited[key] = true	// Mark as visited
				remainNode--
			}
		}
		curTrackID++
	}
	// Bubble sort satellites according to Angle to get inTrackID
	for trackID, keyGroup := range classifySatsUUIDList {
		for idx1 := 0; idx1 < len(keyGroup) - 1; idx1++ {
			for idx2 := 0; idx2 < len(keyGroup) - 1 - idx1; idx2++ {
				sat1, _ := ps.c.FindSatelliteByName(keyGroup[idx2])
				sat2, _ := ps.c.FindSatelliteByName(keyGroup[idx2 + 1])
				if sat1.AngleDeltaAtTime(sat2, year, month, day, hour, minute, second) > 0 {
					keyGroup[idx2], keyGroup[idx2 + 1] = keyGroup[idx2 + 1], keyGroup[idx2]
				}
			}
		}
		// Assign TrackID and InTrackID
	
		for inTrackID, key := range keyGroup {
			ps.cache.satCache[key].TrackID = trackID
			ps.cache.satCache[key].InTrackID = inTrackID
		}
	}

	fmt.Println(classifySatsUUIDList)
}

// Function: RunPositionModule
// Description: Start Position Computing Module.
// 1. inputPath: TLE file's path.
// 2. fixedNum: The number of fixed network pod expected to generate.
func RunPositionModule(inputPath string, fixedNum int) {
	// Construct Constellation from file 
	ps := NewPositionServer(inputPath, fixedNum)

	// Bind handler and start server
	http.HandleFunc("/location", ps.GetLocationHandler)
	http.ListenAndServe(":30100", nil)
	
}