package util

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	topov1 "github.com/y-young/kube-dtn/api/v1"
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

func ShowTopology(topology *topov1.Topology) {
	fmt.Printf("%s ->", topology.Name)
	for _, topo := range topology.Spec.Links {
		fmt.Printf(" %s", topo.PeerPod)
	}
	fmt.Print("\n")
}

func GetLinkName(name string) string {
	if len(name) > 15 {
		name = name[:15]
	}
	return name
}