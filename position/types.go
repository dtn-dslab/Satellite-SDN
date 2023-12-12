package position

type RetParams struct {
	TimeStamp  int64         `json:"unixTimeStamp"`
	Satellites []SatParams   `json:"satellites"`
	Missiles   []MSParams    `json:"missiles"`
	Stations   []GSParams    `json:"stations"`
	FixedNodes []FixedParams `json:"fixedNodes"`
}

type SatParams struct {
	SatGroupID int     `json:"satGroupID"`
	TrackID    int     `json:"trackID"`
	InTrackID  int     `json:"inTrackID"`
	UUID       string  `json:"uuid"`
	Longitude  float64 `json:"lon"`
	Latitude   float64 `json:"lat"`
	Altitude   float64 `json:"height"`
}

type GSParams struct {
	UUID      string  `json:"uuid"`
	Longitude float64 `json:"lon"`
	Latitude  float64 `json:"lat"`
	Altitude  float64 `json:"height"`
}

type FixedParams struct {
	UUID      string  `json:"uuid"`
	Longitude float64 `json:"lon"`
	Latitude  float64 `json:"lat"`
	Altitude  float64 `json:"height"`
}

type MSParams struct {
	UUID      string  `json:"uuid"`
	Longitude float64 `json:"lon"`
	Latitude  float64 `json:"lat"`
	Altitude  float64 `json:"height"`
}
