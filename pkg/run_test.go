package sdn

import "testing"

func TestRunWithoutTimeout(t *testing.T) {
	inputFilePath := "./data/geodetic.txt"
	outputDir := "./output"
	err := ApplyConfig(inputFilePath, outputDir, 3)
	if err != nil {
		t.Error(err)
	}
}