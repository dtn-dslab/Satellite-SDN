package util

import (
	"flag"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func GetClientset() (*kubernetes.Clientset, error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("CONFIG ERROR: %v", err)
	}

	// create the clientset
	return kubernetes.NewForConfig(config)
}

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