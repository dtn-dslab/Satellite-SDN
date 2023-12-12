package util

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"

	"k8s.io/client-go/util/homedir"
)

func GetNamespace() (string, error) {
	cmd := exec.Command(
		"/bin/sh",
		"-c",
		fmt.Sprintf(
			"cat %s | grep namespace | tr -d ' ' | sed 's/namespace://g' | tr -d '\n'",
			filepath.Join(homedir.HomeDir(), ".kube", "config"),
		),
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("EXEC COMMAND FAILURE: %v", err)
	}

	namespace := strings.Trim(string(output), "\"")
	return namespace, nil
}

func GetSlaveNodes(nodeNum int) ([]string, error) {
	// Execute `kubectl get nodes | grep none` to get slave nodes
	cmd := exec.Command("/bin/sh", "-c", "kubectl get nodes | grep none")
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("get nodes failed: %v", err)
	} else {
		result := []string{}
		lines := strings.Split(string(output), "\n")
		for idx := 0; idx < len(lines) && idx < nodeNum; idx++ {
			line := lines[idx]
			name, _, _ := strings.Cut(line, " ")
			result = append(result, name)
		}
		return result, nil
	}
}

func Fetch(url string) (map[string]interface{}, error) {
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch %s error! StatusCode: %d, Details: %v", url, resp.StatusCode, err)
	}
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body error: %v", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(content, &result); err != nil {
		return nil, fmt.Errorf("unmarshal failed: %v", err)
	} else {
		return result, nil
	}
}

func GetLinkName(name string) string {
	if len(name) > 15 {
		name = name[:15]
	}
	return name
}
