package util

import (
	"flag"
	"fmt"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/flowcontrol"
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
	config.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(1000, 1000)

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
	config.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(1000, 1000)

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
	config.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(1000, 1000)

	topoclient, err := rest.RESTClientFor(config)
	return topoclient, err
}

