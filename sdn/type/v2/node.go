package v2

import (
	"math"
	"time"

	gosate "github.com/joshuaferrara/go-satellite"
)

const (
	UNKOWNTYPE = 0
	LOWORBIT = 1
	HIGHORBIT = 2
	GROUNDSTATION = 3
	MISSILE = 4
	FIXED = 5

	NAME_PREFIX_V2 = "sdn"
)

type NodeType int

type NodeInterface interface {
	Position() (x, y, z float64)
	PositionAtTime(t time.Time) (x, y, z float64)
	DistanceWithNode(node *Node) float64
	DistanceWithNodeAtTime(node *Node, t time.Time) float64
	Angle() float64
	AngleAtTime(t time.Time) float64
	AngleDeltaWithNode(node *Node) float64
	AngleDeltaWithNodeAtTime(node *Node, t time.Time) float64
}

type Node struct {
	Type NodeType 
	UUID string 
	TrackID int 
	InTrackID int 
	Latitude float64 
	Longitude float64 
	Altitude float64 
}

func NewSatNode(nodeType NodeType, params map[string]interface{}) Node {
	return Node{
		Type: nodeType,
		UUID: params["uuid"].(string),
		TrackID: (int) (params["trackID"].(float64)),
		InTrackID: (int) (params["inTrackID"].(float64)),
		Latitude: params["lat"].(float64),
		Longitude: params["lon"].(float64),
		Altitude: params["height"].(float64),
	}
}

func NewOtherNode(nodeType NodeType, params map[string]interface{}) Node {
	return Node{
		Type: nodeType,
		UUID: params["uuid"].(string),
		Latitude: params["lat"].(float64),
		Longitude: params["lon"].(float64),
		Altitude: params["height"].(float64),
	}
}

// Return the position of node expressed by x/y/z
func (n *Node) Position() (x, y, z float64) {
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
		Latitude: n.Latitude,
		Longitude: n.Longitude,
	}
	alt := n.Altitude
	jday := gosate.JDay(
		year, month, day,
		hour, minute, second,
	)

	// Convert LLA to ECI
	eciObs := gosate.LLAToECI(obsCoords, alt, jday)
	return eciObs.X, eciObs.Y, eciObs.Z
}

// Return the position of node expressed by x/y/z at given time
func (n *Node) PositionAtTime(t time.Time) (x, y, z float64) {
	// Declare time
	year, month, day, hour, minute, second :=
		t.Year(), int(t.Month()),
		t.Day(), t.Hour(),
		t.Minute(), t.Second()
	
	// Construct params for conversion
	obsCoords := gosate.LatLong{
		Latitude: n.Latitude,
		Longitude: n.Longitude,
	}
	alt := n.Altitude
	jday := gosate.JDay(
		year, month, day,
		hour, minute, second,
	)

	// Convert LLA to ECI
	eciObs := gosate.LLAToECI(obsCoords, alt, jday)
	return eciObs.X, eciObs.Y, eciObs.Z
}

// Return the distance between the node and the specified node in kilometer.
func (n *Node) DistanceWithNode(node *Node) float64 {
	x1, y1, z1 := n.Position()
	x2, y2, z2 := node.Position()
	distance := math.Sqrt(
		(x2 - x1) * (x2 - x1) +
		(y2 - y1) * (y2 - y1) +
		(z2 - z1) * (z2 - z1),
	)
	return distance
}

// Return the distance between the node and the specified node in kilometer at given time.
func (n *Node) DistanceWithNodeAtTime(node *Node, t time.Time) float64 {
	x1, y1, z1 := n.PositionAtTime(t)
	x2, y2, z2 := node.PositionAtTime(t)
	distance := math.Sqrt(
		(x2 - x1) * (x2 - x1) +
		(y2 - y1) * (y2 - y1) +
		(z2 - z1) * (z2 - z1),
	)
	return distance
}

// Return the angle in WGS84's x-y plane.
// The return value ranges from 0 to 2*pi
func (n *Node) Angle() float64 {
	x, y, _ := n.Position()
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
func (n *Node) AngleAtTime(t time.Time) float64 {
	x, y, _ := n.PositionAtTime(t)
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
func (n *Node) AngleDelta(node *Node) float64 {
	angleDelta := node.Angle() - n.Angle()
	if angleDelta > math.Pi {
		angleDelta = angleDelta - 2*math.Pi
	} else if angleDelta <= -math.Pi {
		angleDelta = 2*math.Pi + angleDelta
	}
	return angleDelta
}

// Angle(sat2) - Angle(self) in WGS84's x-y plane
// The return value ranges from -pi(does not contain) to pi
func (n *Node) AngleDeltaAtTime(node *Node, t time.Time) float64 {
	angleDelta := node.AngleAtTime(t) - n.AngleAtTime(t)
	if angleDelta > math.Pi {
		angleDelta = angleDelta - 2*math.Pi
	} else if angleDelta <= -math.Pi {
		angleDelta = 2*math.Pi + angleDelta
	}
	return angleDelta
}