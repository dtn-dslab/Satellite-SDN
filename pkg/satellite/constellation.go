package satellite

import (
	"fmt"
	// "log"
	"io/ioutil"
	"strings"
)

type Constellation struct {
	Satellites []Satellite
}

func NewConstellation(filePath string) (*Constellation, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("Error in reading file %s:%v\n", filePath, err)
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	var constellation Constellation
	for idx := 0; idx < len(lines); idx += 3 {
		name := strings.Replace(strings.Trim(lines[idx], " "), " ", "-", -1)
		name = strings.Replace(name, "-", "", -1)
		name = strings.Replace(name, "(", "", -1)
		name = strings.Replace(name, ")", "", -1)
		name = strings.ToLower(name)
		sat, err := NewStatellite(name, lines[idx+1], lines[idx+2])
		if err != nil {
			return nil, fmt.Errorf("Error in creating constellation: %v\n", err)
		}
		constellation.Satellites = append(constellation.Satellites, *sat)
	}

	return &constellation, nil
}

func (c *Constellation) FindSatelliteByName(name string) (Satellite, error) {
	for _, s := range c.Satellites {
		if s.Name == name {
			return s, nil
		}
	}
	return Satellite{}, fmt.Errorf("Could not find the satellite of which name is %s\n", name)
}

// Angle(sat2) - Angle(sat1) in WGS84's x-y plane
func (c *Constellation) AngleDelta(sat1, sat2 string) (float64, error) {
	// Find corresponding satellites
	satellite1, err := c.FindSatelliteByName(sat1)
	if err != nil {
		return 0, fmt.Errorf("The first satellite is not in the constellation\n")
	}
	satellite2, err := c.FindSatelliteByName(sat2)
	if err != nil {
		return 0, fmt.Errorf("The second satellite is not in the constellation\n")
	}

	// Calculate angleDelta
	return satellite1.AngleDelta(satellite2), nil
}

func (c *Constellation) Distance(sat1Name, sat2Name string) (float64, error) {
	satellite1, err := c.FindSatelliteByName(sat1Name)
	if err != nil {
		return -1, fmt.Errorf("%v", err)
	}

	satellite2, err := c.FindSatelliteByName(sat2Name)
	if err != nil {
		return -1, fmt.Errorf("%v", err)
	}

	return satellite1.Distance(satellite2), nil
}

