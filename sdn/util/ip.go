package util

import (
	"fmt"
	"os/exec"
	"strings"
)

// Get a pod's ip via 'kubectl get pod <podName> -o wide' instruction by parsing the output.
func GetPodIP(podName string) (string, error) {
	// Executing 'kubectl get pod <podName> -o wide'
	cmd := exec.Command("kubectl", "get", "pod", podName, "-o", "wide")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Executing kubectl get pods failed: %v\n", err)
	}

	// Get pod ip
	lines := strings.Split(string(output), "\n")
	if len(lines) <= 1 {
		return "", fmt.Errorf("Can't find pod: %s\n", podName)
	}
	blocks := strings.Split(lines[1], " ")
	newBlocks := []string{}
	for _, block := range blocks {
		if block != "" {
			newBlocks = append(newBlocks, block)
		}
	}
	if newBlocks[5] != "<none>" {
		return newBlocks[5], nil
	} else {
		return "", fmt.Errorf("Can't find pod: %s\n", podName)
	}

}

// IP = uid << 2 | 0x1/0x2 (according to whether myID < peerID)
// The highest bit of IP must be 1
// Can support allocating IP to at most 2^14 pods
func GetVxlanIP(myID, peerID uint) string {
	var uid uint
	netIP := make([]string, 4)
	if myID < peerID {
		uid = ((myID << 15) + peerID) << 2
		netIP[3] = fmt.Sprint(uid & 0xff | 0x01)
	} else {
		uid = ((peerID << 15) + myID) << 2
		netIP[3] = fmt.Sprint(uid & 0xff | 0x02)
	}
	netIP[0] = fmt.Sprint((uid >> 24) & 0xff | 0x80)
	netIP[1] = fmt.Sprint((uid >> 16) & 0xff)
	netIP[2] = fmt.Sprint((uid >> 8) & 0xff)
	
	return strings.Join(netIP, ".") + "/30"
}

// Allocate global IP to pod according to its idx
// IP = 10.233.((idx >> 8)&0xff).(idx&0xff)
func GetGlobalIP(myID uint) string {
	netIP := make([]string, 4)
	netIP[0] = "10"
	netIP[1] = "233"
	netIP[2] = fmt.Sprint((myID >> 8) & 0xff)
	netIP[3] = fmt.Sprint(myID & 0xff)

	return strings.Join(netIP, ".")
}
