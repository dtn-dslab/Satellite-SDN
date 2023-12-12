package pod

import (
	"testing"
)

func TestGeneratePodYaml(t *testing.T) {
	nameMap := map[int]string{
		0: "first",
		1: "second",
	}
	if err := PodSyncLoop(nameMap); err != nil {
		t.Errorf("Pod sync loop failed: %v", err)
	}
}
