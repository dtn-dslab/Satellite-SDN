package satellite

import (
	"fmt"
	"time"
	// "log"
	"math"

	gosate "github.com/joshuaferrara/go-satellite"
)

type Satellite struct {
	Name      string
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
		Name:      name,
		Satellite: gosate.TLEToSat(line1, line2, "wgs72"),
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
	// log.Printf("%s: x is %2f, y is %2f, z is %2f\n",
	// 	sat.Name, position.X, position.Y, position.Z)

	return position.X, position.Y, position.Z
}

// Return the position of satellite expressed by x/y/z at given time
func (sat *Satellite) PositionAtTime(year, month, day, hour, minute, second int) (x, y, z float64) {
	// Get current position
	position, _ := gosate.Propagate(
		sat.Satellite, year, month,
		day, hour, minute, second,
	)

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
	// log.Printf("%s: long is %2f, lat is %2f, alt is %2f\n",
	// 	sat.Name, latlong.Longitude, latlong.Latitude, altitude)

	return latlong.Longitude, latlong.Latitude, altitude
}

// Return the location of satellite expressed by longitude/latitude/altitude at given time
func (sat *Satellite) LocationAtTime(year, month, day, hour, minute, second int) (long, lat, alt float64) {
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
	// log.Printf("%s: long is %2f, lat is %2f, alt is %2f\n",
	// 	sat.Name, latlong.Longitude, latlong.Latitude, altitude)

	return latlong.Longitude, latlong.Latitude, altitude
}


// Return the distance between two satellites in kilometer.
func (sat *Satellite) Distance(anotherSat Satellite) float64 {
	// now := time.Now()
	x1, y1, z1 := sat.Position()
	x2, y2, z2 := anotherSat.Position()
	// log.Printf("year is %d, month is %d, day is %d, hour is %d, minute is %d, second is %d\n",
	// 				now.Year(), int(now.Month()), now.Day(), now.Hour(), now.Minute(), now.Second())
	distance := math.Sqrt(
		(x2-x1)*(x2-x1) +
			(y2-y1)*(y2-y1) +
			(z2-z1)*(z2-z1),
	)
	return distance
}

// Return the distance between two satellites in kilometer at given time.
func (sat *Satellite) DistanceAtTime(anotherSat Satellite, year, month, day, hour, minute, second int) float64 {
	x1, y1, z1 := sat.PositionAtTime(year, month, day, hour, minute, second)
	x2, y2, z2 := anotherSat.PositionAtTime(year, month, day, hour, minute, second)
	// log.Printf("year is %d, month is %d, day is %d, hour is %d, minute is %d, second is %d\n",
	// 				now.Year(), int(now.Month()), now.Day(), now.Hour(), now.Minute(), now.Second())
	distance := math.Sqrt(
		(x2-x1)*(x2-x1) +
			(y2-y1)*(y2-y1) +
			(z2-z1)*(z2-z1),
	)
	return distance
}

// Return the angle in WGS84's x-y plane.
// The return value ranges from 0 to 2*pi
func (sat *Satellite) Angle() float64 {
	x, y, _ := sat.Position()
	if x > 0 && y > 0 {
		return math.Atan(y / x)
	} else if x > 0 && y < 0 {
		return 2*math.Pi + math.Atan(y/x)
	} else {
		return math.Pi + math.Atan(y/x)
	}
}

// Return the angle in WGS84's x-y plane at given time.
// The return value ranges from 0 to 2*pi
func (sat *Satellite) AngleAtTime(year, month, day, hour, minute, second int) float64 {
	x, y, _ := sat.PositionAtTime(year, month, day, hour, minute, second)
	if x > 0 && y > 0 {
		return math.Atan(y / x)
	} else if x > 0 && y < 0 {
		return 2*math.Pi + math.Atan(y/x)
	} else {
		return math.Pi + math.Atan(y/x)
	}
}

// Angle(sat2) - Angle(self) in WGS84's x-y plane
// The return value ranges from -pi(does not contain) to pi
func (sat *Satellite) AngleDelta(anotherSat Satellite) float64 {
	angleDelta := anotherSat.Angle() - sat.Angle()
	if angleDelta > math.Pi {
		angleDelta = angleDelta - 2*math.Pi
	} else if angleDelta <= -math.Pi {
		angleDelta = 2*math.Pi + angleDelta
	}
	return angleDelta
}

// Angle(sat2) - Angle(self) in WGS84's x-y plane
// The return value ranges from -pi(does not contain) to pi
func (sat *Satellite) AngleDeltaAtTime(anotherSat Satellite, year, month, day, hour, minute, second int) float64 {
	angleDelta := 
		anotherSat.AngleAtTime(year, month, day, hour, minute, second) - 
		sat.AngleAtTime(year, month, day, hour, minute, second)
	if angleDelta > math.Pi {
		angleDelta = angleDelta - 2*math.Pi
	} else if angleDelta <= -math.Pi {
		angleDelta = 2*math.Pi + angleDelta
	}
	return angleDelta
}
