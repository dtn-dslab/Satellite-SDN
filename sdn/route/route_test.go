package route

import (
	"reflect"
	"testing"
)

func TestComputeRoutes(t *testing.T) {
	var distanceMap = [][]float64{
        {0, 100, 1200, 1e9, 1e9, 1e9},
        {100, 0, 900, 300, 1e9, 1e9},
        {1200, 900, 0, 400, 500, 1e9},
        {1e9, 300, 400, 0, 1300, 1400},
        {1e9, 1e9, 500, 1300, 0, 1500},
        {1e9, 1e9, 1e9, 1400, 1500, 0},
    }
	var result = [][]int{
		{0, 1, 1, 1, 1, 1},
		{0, 1, 3, 3, 3, 3},
		{3, 3, 2, 3, 4, 3},
		{1, 1, 2, 3, 2, 5},
		{2, 2, 2, 2, 4, 5},
		{3, 3, 3, 3, 4, 5},
	}
	routeTable := ComputeRoutes(distanceMap, 6)
	t.Log(routeTable)
	t.Log(distanceMap)
	if !reflect.DeepEqual(result, routeTable) {
		t.Errorf("Result error!")
	}
}
