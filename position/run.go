package position

import (
	"encoding/json"
	"math"
	"net/http"
	"time"
	"sync"

	sdnv1 "ws/dtn-satellite-sdn/sdn/type/v1"

	"github.com/sirupsen/logrus"
)

type PositionServerInterface interface {
	GetLocationHanlder(http.ResponseWriter, *http.Request)
	Init()
	Update() error
}

type PositionServer struct {
	c         *sdnv1.Constellation
	cache     *PositionCache
	fixedNum  int
	timeStamp time.Time
}

type PositionCache struct {
	satCache   map[string]*SatParams
	msCache    []MSParams
	gsCache    []GSParams
	fixedCache []FixedParams
}

const (
	MaxOrbitSize int = 50
)

func NewPositionServer(inputPath string, num int, maxNum int) *PositionServer {
	logger := logrus.WithFields(logrus.Fields{
		"input-path": 		  inputPath,
		"fixed-num":  		  num,
		"max-satellites-num": maxNum,
	})
	logger.Info("initializing position server...")
	if constellation, err := sdnv1.NewConstellation(inputPath); err != nil {
		logrus.WithError(err).Fatal("create constellation failed")
		return nil
	} else {
		satelliteNum := len(constellation.Satellites)
		if satelliteNum > maxNum {
			constellation.Satellites = constellation.Satellites[:maxNum]
		}
		ps := PositionServer{
			c:        constellation,
			cache:    &PositionCache{
				satCache:   make(map[string]*SatParams),
				msCache:    GetMissiles(),
				gsCache:    GetGroundStation(),
				fixedCache: GetFixedNodes(num),
			},
			fixedNum: num,
		}
		ps.Init()
		logger.Info("position server has been initailized")
		return &ps
	}
}

// Function: GetLocationHanlder
// Description: A http hanlder for getting location of all types of node.
func (ps *PositionServer) GetLocationHandler(w http.ResponseWriter, req *http.Request) {
	ps.timeStamp = time.Now()
	if err := ps.Update(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		logrus.WithError(err).Error("update cache failed.")
	}
	retParams := RetParams{
		TimeStamp:  ps.timeStamp.UnixMilli(),
		Satellites: []SatParams{},
		Missiles:   ps.cache.msCache,
		Stations:   ps.cache.gsCache,
		FixedNodes: ps.cache.fixedCache,
	}
	for _, sat := range ps.cache.satCache {
		retParams.Satellites = append(retParams.Satellites, *sat)
	}
	logrus.Debugf("return value is %v", retParams)
	content, _ := json.Marshal(&retParams)
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

// Function: Update
// Description: Update cache in ps.cache for future use.
func (ps *PositionServer) Update() error {
	// Update longitude, latitude, altitude in Satellites
	year, month, day, hour, minute, second :=
		ps.timeStamp.Year(),
		int(ps.timeStamp.Month()),
		ps.timeStamp.Day(),
		ps.timeStamp.Hour(),
		ps.timeStamp.Minute(),
		ps.timeStamp.Second()
	logrus.Info("update longitude, latitude, altitude in cache")
	
	var wg sync.WaitGroup
	wg.Add(len(ps.c.Satellites))
	for _, sat := range ps.c.Satellites {
		go func(s sdnv1.Satellite) {
			long, lat, alt := s.LocationAtTime(
				year, month, day,
				hour, minute, second,
			)
			ps.cache.satCache[s.Name].Longitude = long
			ps.cache.satCache[s.Name].Latitude = lat
			ps.cache.satCache[s.Name].Altitude = alt
			wg.Done()
		}(sat)
	}
	wg.Wait()
	return nil
}

// Function: Init
// Description: Compute all of sats' information when cache is recently created.
func (ps *PositionServer) Init() {
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
			UUID:      sat.Name,
			Longitude: long,
			Latitude:  lat,
			Altitude:  alt,
		}
		ps.cache.satCache[sat.Name] = &s
	}
	// Classify Satellites
	curTrackID, remainNode := 0, len(ps.cache.satCache)
	classifySatsUUIDList := [][]string{}
	visited := make(map[string]bool, remainNode)
	for k := range ps.cache.satCache {
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
		curOrbitSize := 0
		for key, sat := range ps.cache.satCache {
			if !visited[key] && math.Abs(sat.Altitude-standardHeight) < 500 {
				if len(classifySatsUUIDList) <= curTrackID {
					classifySatsUUIDList = append(classifySatsUUIDList, []string{})
				}
				classifySatsUUIDList[curTrackID] = append(classifySatsUUIDList[curTrackID], key) // Update result
				visited[key] = true                                                              // Mark as visited
				remainNode--
				curOrbitSize++
				if curOrbitSize >= MaxOrbitSize {
					curTrackID++
					curOrbitSize = 0
				}
			}
		}
		curTrackID++
	}
	// Bubble sort satellites according to Angle to get inTrackID
	for trackID, keyGroup := range classifySatsUUIDList {
		for idx1 := 0; idx1 < len(keyGroup)-1; idx1++ {
			for idx2 := 0; idx2 < len(keyGroup)-1-idx1; idx2++ {
				sat1, _ := ps.c.FindSatelliteByName(keyGroup[idx2])
				sat2, _ := ps.c.FindSatelliteByName(keyGroup[idx2+1])
				if sat1.AngleDeltaAtTime(sat2, year, month, day, hour, minute, second) > 0 {
					keyGroup[idx2], keyGroup[idx2+1] = keyGroup[idx2+1], keyGroup[idx2]
				}
			}
		}
		// Assign TrackID and InTrackID

		for inTrackID, key := range keyGroup {
			ps.cache.satCache[key].TrackID = trackID
			ps.cache.satCache[key].InTrackID = inTrackID
		}
	}
	
	logrus.WithField("group-num", len(classifySatsUUIDList)).Debug(classifySatsUUIDList)
}

// Function: RunPositionModule
// Description: Start Position Computing Module.
// 1. inputPath: TLE file's path.
// 2. fixedNum: The number of fixed network pod expected to generate.
// 3. maxNum: The max number of satellites.
func RunPositionModule(inputPath string, fixedNum int, maxNum int) {
	// Construct Constellation from file
	ps := NewPositionServer(inputPath, fixedNum, maxNum)

	// Bind handler and start server
	http.HandleFunc("/location", ps.GetLocationHandler)
	http.ListenAndServe(":30100", nil)

}
