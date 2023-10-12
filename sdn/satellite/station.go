package satellite

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	gosate "github.com/joshuaferrara/go-satellite"
)

type GroundStation struct {
	Name string `json:"name,omitempty"`
	Latitude float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	Altitude float64 `json:"altitude,omitempty"`
}

func NewGroundStation(params interface{}) (*GroundStation, error) {
	var result GroundStation
	if err := json.Unmarshal(params.([]byte), &result); err != nil {
		return nil, fmt.Errorf("Create ground station error: %v", err)
	}
	return &result, nil
}

// Return the position of ground station expressed by x/y/z
func (s *GroundStation) Position() (x, y, z float64) {
	// Declare current time
	year, month, day, hour, minute, second :=
		time.Now().Year(),
		int(time.Now().Month()),
		time.Now().Day(),
		time.Now().Hour(),
		time.Now().Minute(),
		time.Now().Second()

	// Construct params for conversion
	obsCoords := gosate.LatLong{
		Latitude: s.Latitude,
		Longitude: s.Longitude,
	}
	alt := s.Altitude
	jday := gosate.JDay(
		year, month, day,
		hour, minute, second,
	)

	// Convert LLA to ECI
	eciObs := gosate.LLAToECI(obsCoords, alt, jday)
	return eciObs.X, eciObs.Y, eciObs.Z
}

// Return the position of satellite expressed by x/y/z at given time
func (s *GroundStation) PositionAtTime(year, month, day, hour, minute, second int) (x, y, z float64) {
	// Construct params for conversion
	obsCoords := gosate.LatLong{
		Latitude: s.Latitude,
		Longitude: s.Longitude,
	}
	alt := s.Altitude
	jday := gosate.JDay(
		year, month, day,
		hour, minute, second,
	)

	// Convert LLA to ECI
	eciObs := gosate.LLAToECI(obsCoords, alt, jday)
	return eciObs.X, eciObs.Y, eciObs.Z
}

// Return the distance between the ground station and the specified satellite in kilometer.
func (s *GroundStation) DistanceWithSat(sat *Satellite) float64 {
	x1, y1, z1 := s.Position()
	x2, y2, z2 := sat.Position()
	distance := math.Sqrt(
		(x2-x1)*(x2-x1) +
			(y2-y1)*(y2-y1) +
			(z2-z1)*(z2-z1),
	)
	return distance
}

// Return the distance between the ground station and the specified satellite in kilometer at given time.
func (s *GroundStation) DistanceWithSatAtTime(sat *Satellite, year, month, day, hour, minute, second int) float64 {
	x1, y1, z1 := s.PositionAtTime(
		year, month, day,
		hour, minute, second,
	)
	x2, y2, z2 := sat.PositionAtTime(
		year, month, day, 
		hour, minute, second,
	)
	distance := math.Sqrt(
		(x2-x1)*(x2-x1) +
			(y2-y1)*(y2-y1) +
			(z2-z1)*(z2-z1),
	)
	return distance
}

