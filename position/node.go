package position



func GetGroundStation() []GSParams {
	return []GSParams{
		{
			UUID: "station",
			Longitude: 121.4476,
			Latitude: 31.1618,
			Altitude: 0.02,
		},
	}
}

func GetMissiles() []MSParams {
	return []MSParams{
		{
			UUID: "missile",
			Longitude: 122.4476,
			Latitude: 32.1618,
			Altitude: 10.002,
		},
	}
}

func GetFixedNodes(num int) []FixedParams {
	return nil
}