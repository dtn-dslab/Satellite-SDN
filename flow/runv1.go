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
	ipSymbol    = "10.233"
	ClientLabel = "type=client"
	ServerLabel = "type=server"
)

func StartServer() error {
	opts := metav1.ListOptions{
		LabelSelector: ServerLabel,
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
	for _, pod := range podList.Items {
		cmd := exec.Command("bash", "-c", fmt.Sprintf("kubectl exec -it -n %s %s -- sh -c ./flow/server.sh ", namespace, pod.GetName()))
		go cmd.Run()
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

func StartClient(bandwidth string) error {
	opts := metav1.ListOptions{
		LabelSelector: ClientLabel,
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
	for _, pod := range podList.Items {
		cmd := exec.Command("bash", "-c", fmt.Sprintf("kubectl exec -it -n %s %s -- sh -c \"./flow/client %s &\"", namespace, pod.GetName(), bandwidth))
		go cmd.Run()
	}
	return nil
}
