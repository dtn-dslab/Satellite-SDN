package sdn

import "testing"

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