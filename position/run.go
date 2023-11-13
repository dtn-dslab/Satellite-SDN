package position

import (
	"encoding/json"
	"net/http"
	"time"

	sdnv1 "ws/dtn-satellite-sdn/sdn/type/v1"
)

type PositionServerInterface interface {
	GetLocation(http.ResponseWriter, *http.Request)
	ComputeSats() []SatParams
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

func (ps *PositionServer) GetLocation(w http.ResponseWriter, req *http.Request) {
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
	content, _ := json.Marshal(&retParams)
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

func (ps *PositionServer) UpdateCache() error {
	if ps.cache == nil {
		ps.cache = &PositionCache{
			satCache: make(map[string]*SatParams),
			msCache: GetMissiles(),
			gsCache: GetGroundStation(),
			fixedCache: GetFixedNodes(ps.fixedNum),
		}
		satParams := ps.ComputeSats()
		for _, sat := range satParams {
			ps.cache.satCache[sat.UUID] = &sat
		}
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

func (ps *PositionServer) ComputeSats() []SatParams {
	return nil
}

func RunPositionModule(inputPath string, num int) {
	// Construct Constellation from file 
	ps := NewPositionServer(inputPath, num)

	// Bind handler and start server
	http.HandleFunc("/location", ps.GetLocation)
	http.ListenAndServe(":30100", nil)
	
}