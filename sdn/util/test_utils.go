package util

import (
	"os/exec"
	"strings"
	"time"
)

func InitEnvTimeCounter(startTimer time.Time) (float64, error) {
	for ;!isPodOk() || !isTopoOk() || !isRouteOk(); {
		time.Sleep(3 * time.Second)
	}

	endTimer := time.Now()
	return endTimer.Sub(startTimer).Seconds(), nil
}

func isPodOk() bool {
	// Executing 'kubectl get pod <podName> -o wide'
	cmd := exec.Command("bash", "-c", "kubectl get pod -o wide | grep -v Running")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	
	// Parse file content to []string & Judge if all pods have been created
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 && lines[len(lines) - 1] == "" {
		lines = lines[:len(lines) - 1]
	}
	if len(lines) > 1 {
		return false
	}
	return true
}

// TODO(ws): Judge topology's state
func isTopoOk() bool {
	return true
}

// TODO(ws): Judge route's state
func isRouteOk() bool {
	return true
}