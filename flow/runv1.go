package flow

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"ws/dtn-satellite-sdn/sdn/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ipSymbol = "10.233"
)

func CreateFlows(Bandwidth string) error {
	opts := metav1.ListOptions{
		LabelSelector: "type=flow",
	}
	clientset, err := util.GetClientset()
	if err != nil {
		return fmt.Errorf("CREATE CLIENTSET ERROR: %v", err)
	}

	// get current namespace
	namespace, err := util.GetNamespace()
	if err != nil {
		return fmt.Errorf("GET NAMESPACE ERROR: %v", err)
	}
	podList, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), opts)
	if err != nil {
		return fmt.Errorf("GET PODLIST ERROR: %v", err)
	}
	for i := 0; i < len(podList.Items); i += 2 {
		clientPod := podList.Items[i]
		serverPod := podList.Items[i+1]
		HostIP := GetIPByIfconfig(namespace, clientPod.GetName())
		ServerIP := GetIPByIfconfig(namespace, clientPod.GetName())

		cmd := exec.Command("bash", "-c", fmt.Sprintf("kubectl exec -it -n %s %s -- iperf3 -s -p %d –i 1", namespace, serverPod.GetName(), 5202))
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("START IPERF SERVER ERROR: %v", err)
		}
		cmd = exec.Command("bash", "-c", fmt.Sprintf("kubectl exec -it -n %s %s -- iperf3 -c %s -B %s -b %s -p %d –i 1 -t 1000", namespace, clientPod.GetName(), HostIP, Bandwidth, ServerIP, 5202))
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("START IPERF CLIENT ERROR: %v", err)
		}
	}
	return nil
}

func GetIPByIfconfig(namespace string, podname string) string {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("kubectl exec -it -n %s %s -- ifconfig", namespace, podname))
	stdout, _ := cmd.CombinedOutput()
	ifconfig := string(stdout)
	words := strings.Split(ifconfig, " ")
	var IP string
	for _, word := range words {
		if strings.Contains(word, ipSymbol) {
			IP = word[5:]
			break
		}
	}
	return IP
}
