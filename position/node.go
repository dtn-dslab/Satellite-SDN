package position

import "fmt"

func GetGroundStation() []GSParams {
	return []GSParams{
		// Shanghai Sheshan Station.
		{
			UUID:      "station",
			Longitude: 121.4476,
			Latitude:  31.1618,
			Altitude:  0.02,
		},
	}
}

func GetMissiles() []MSParams {
	return []MSParams{
		{
			UUID:      "missile",
			Longitude: 122.4476,
			Latitude:  32.1618,
			Altitude:  10.002,
		},
	}
}

func GetFixedNodes(num int) []FixedParams {
	result := make([]FixedParams, 0, num)
	for idx := 0; idx < num; idx++ {
		result = append(result, FixedParams{
			UUID:      "fixed" + fmt.Sprint(idx),
			Longitude: 0.0 + 0.1*float64(idx),
			Latitude:  0.0 + 0.05*float64(idx),
			Altitude:  0.0,
		})
	}
	return result
}
