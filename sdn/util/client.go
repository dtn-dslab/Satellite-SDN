package util

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"

	// "k8s.io/apimachinery/pkg/runtime"
	// "k8s.io/apimachinery/pkg/runtime/schema"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	sdnv1 "ws/dtn-satellite-sdn/api/v1"
	topov1 "github.com/y-young/kube-dtn/api/v1"
)

var (
	kubeconfig *string = nil
	clientset *kubernetes.Clientset = nil
	routeclient *rest.RESTClient = nil
	topoclient *rest.RESTClient = nil
)

func init() {
	sdnv1.AddToScheme(scheme.Scheme)
	topov1.AddToScheme(scheme.Scheme)
}

func GetClientset() (*kubernetes.Clientset, error) {
	// If clientset is not empty, return clientset
	if clientset != nil {
		return clientset, nil
	}

	// init kubeconfig
	if kubeconfig == nil {
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
	}

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("CONFIG ERROR: %v", err)
	}

	// create the clientset
	clientset, err = kubernetes.NewForConfig(config)
	return clientset, err
}

func GetRouteClient() (*rest.RESTClient, error) {
	// If RouteClient is not empty, return RouteClient
	if routeclient != nil {
		return routeclient, nil
	}

	// init kubeconfig
	if kubeconfig == nil {
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
	}

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("CONFIG ERROR: %v", err)
	}

	config.APIPath = "/apis"
	config.ContentConfig.GroupVersion = &sdnv1.GroupVersion
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	config.UserAgent = rest.DefaultKubernetesUserAgent()

	routeclient, err = rest.RESTClientFor(config)
	return routeclient, err
}

func GetTopoClient() (*rest.RESTClient, error) {
	// If TopoClient is not empty, return TopoClient
	if topoclient != nil {
		return topoclient, nil
	}

	// init kubeconfig
	if kubeconfig == nil {
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
	}
	
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("CONFIG ERROR: %v", err)
	}

	config.APIPath = "/apis"
	config.ContentConfig.GroupVersion = &topov1.GroupVersion
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	config.UserAgent = rest.DefaultKubernetesUserAgent()

	topoclient, err := rest.RESTClientFor(config)
	return topoclient, err
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