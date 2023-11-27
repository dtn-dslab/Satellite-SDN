package sdn

import (
	"testing"
)

func TestRunSDNWithoutTimeout(t *testing.T) {
	inputFilePath := "./data/geodetic.txt"
	if err := RunSatelliteSDN(inputFilePath, 7, -1); err != nil {
		t.Error(err)
	}
}

func TestDelSDN(t *testing.T) {
	inputFilePath := "./data/geodetic.txt"
	if err := DelSatelliteSDN(inputFilePath); err != nil {
		t.Error(err)
	}
}

func TestSlice(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	newSlice := append(slice[:2], slice..., )
	t.Log(slice)
	t.Log(newSlice)
}