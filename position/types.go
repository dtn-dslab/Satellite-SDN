package position

type RetParams struct {
	TimeStamp int64			`json:"unixTimeStamp,omitempty"`
	Satellites []SatParams	`json:"satellites,omitempty"`
	Missiles []MSParams		`json:"missiles,omitempty"`
	Stations []GSParams		`json:"stations,omitempty"`
	FixedNodes []FixedParams`json:"fixedNodes,omitempty"`
}

type SatParams struct {
	SatGroupID int 		`json:"satGroupID,omitempty"`
	TrackID int			`json:"trackID,omitempty"`
	InTrackID int		`json:"inTrackID,omitempty"`
	UUID string			`json:"uuid,omitempty"`
	Longitude float64	`json:"lon,omitempty"`
	Latitude float64	`json:"lat,omitempty"`
	Altitude float64	`json:"height,omitempty"`
}

type GSParams struct {
	UUID string			`json:"uuid,omitempty"`
	Longitude float64	`json:"lon,omitempty"`
	Latitude float64	`json:"lat,omitempty"`
	Altitude float64	`json:"height,omitempty"`
}

type FixedParams struct {
	UUID string			`json:"uuid,omitempty"`
	Longitude float64	`json:"lon,omitempty"`
	Latitude float64	`json:"lat,omitempty"`
	Altitude float64	`json:"height,omitempty"`
}

type MSParams struct {
	UUID string			`json:"uuid,omitempty"`
	Longitude float64	`json:"lon,omitempty"`
	Latitude float64	`json:"lat,omitempty"`
	Altitude float64	`json:"height,omitempty"`
}