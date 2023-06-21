package util

import (
	"fmt"
	"time"
	"log"
	"math"
	"strings"
	"io/ioutil"
	gosate"github.com/joshuaferrara/go-satellite"
)

type Satellite struct {
	Name string
	Satellite gosate.Satellite
}

func NewStatellite(name, line1, line2 string) (*Satellite, error) {
	if name == "" {
		return nil, fmt.Errorf("The name of satellite is empty...\n")
	}
	if line1 == "" {
		return nil, fmt.Errorf("The first line of satellite's TLE is empty\n")
	}
	if line2 == "" {
		return nil, fmt.Errorf("The second line of satellite's TLE is empty\n")
	}
	return &Satellite{
		Name: name,
		Satellite: gosate.TLEToSat(line1, line2, "wgs84"),
	}, nil
}

// Return the position of satellite expressed by x/y/z
func (sat *Satellite) Position() (x, y, z float64) {
	// Declare current time
	year, month, day, hour, minute, second := 
		time.Now().Year(),
		int(time.Now().Month()),
		time.Now().Day(),
		time.Now().Hour(),
		time.Now().Minute(),
		time.Now().Second()
	// Get current position
	position, _ := gosate.Propagate(
		sat.Satellite, year, month,
		day, hour, minute, second,
	)
	log.Printf("%s: x is %2f, y is %2f, z is %2f\n", 
		sat.Name, position.X, position.Y, position.Z)

	return position.X, position.Y, position.Z
}

// Return the location of satellite expressed by longitude/latitude/altitude
func (sat *Satellite) Location() (long, lat, alt float64) {
	// Declare current time
	year, month, day, hour, minute, second := 
		time.Now().Year(),
		int(time.Now().Month()),
		time.Now().Day(),
		time.Now().Hour(),
		time.Now().Minute(),
		time.Now().Second()
	// Get current position
	position, _ := gosate.Propagate(
		sat.Satellite, year, month,
		day, hour, minute, second,
	)
	// Convert the current time to Galileo system time (GST)
	gst := gosate.GSTimeFromDate(
		year, month, day,
		hour, minute, second,
	)
	// Get satellite's latitude & longtitude & altitude
	altitude, _, latlong := gosate.ECIToLLA(position, gst)
	latlong = gosate.LatLongDeg(latlong)
	log.Printf("%s: long is %2f, lat is %2f, alt is %2f\n", 
		sat.Name, latlong.Longitude, latlong.Latitude, altitude)

	return latlong.Longitude, latlong.Latitude, altitude
}

// Return the distance between two satellites in kilometer.
func (sat *Satellite) Distance(anotherSat Satellite) int {
	now := time.Now()
	x1, y1, z1 := sat.Position()
	x2, y2, z2 := anotherSat.Position()
	log.Printf("year is %d, month is %d, day is %d, hour is %d, minute is %d, second is %d\n",
					now.Year(), int(now.Month()), now.Day(), now.Hour(), now.Minute(), now.Second())
	distance := math.Sqrt(
		(x2 - x1) * (x2 - x1) + 
		(y2 - y1) * (y2 - y1) + 
		(z2 - z1) * (z2 - z1),
	)
	return int(distance)
}

type Constellation struct {
	Satellites []Satellite
}

func NewConstellation(filePath string) (*Constellation, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("Error in reading file %s:%v\n", filePath, err)
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) > 0 && lines[len(lines) - 1] == "" {
		lines = lines[:len(lines) - 1]
	}

	var constellation Constellation
	for idx := 0; idx < len(lines); idx += 3 {
		// Replace blank with -
		name := strings.Replace(strings.Trim(lines[idx], " "), " ", "-", -1)
		sat, err := NewStatellite(name, lines[idx + 1], lines[idx + 2])
		if err != nil {
			return nil, fmt.Errorf("Error in creating constellation: %v\n", err)
		}
		constellation.Satellites = append(constellation.Satellites, *sat)
	}

	return &constellation, nil
}

func (c *Constellation) findSatelliteByName(name string) (Satellite, error) {
	for _, s := range c.Satellites {
		if s.Name == name {
			return s, nil
		}
	}
	return Satellite{}, fmt.Errorf("Could not find the satellite of which name is %s\n", name)
}

func (c *Constellation) Distance(sat1Name, sat2Name string) (int, error) {
	satellite1, err := c.findSatelliteByName(sat1Name)
	if err != nil {
		return -1, fmt.Errorf("%v", err)
	}

	satellite2, err := c.findSatelliteByName(sat2Name)
	if err != nil {
		return -1, fmt.Errorf("%v", err)
	}

	return satellite1.Distance(satellite2), nil
}