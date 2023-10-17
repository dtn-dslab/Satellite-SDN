package pod

import (
	"context"
	"flag"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"ws/dtn-satellite-sdn/sdn/util"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func PodSyncLoop(nameMap map[int]string) error {
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
		return fmt.Errorf("CONFIG ERROR: %v", err)
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("CREATE CLIENTSET ERROR: %v", err)
	}

	// get current namespace
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
		return fmt.Errorf("EXEC COMMAND FAILURE: %v", err)
	}
	namespace := strings.Trim(string(output), "\"")

	// construct pods
	for idx := 0; idx < len(nameMap); idx++ {
		sat_name := "satellite"
		image_name := "electronicwaste/podserver:v6"
		image_pull_policy := "IfNotPresent"
		var port int32 = 8080
		// TODO(ws): figure out why FieldManager is needed
		// When we delete key 'FieldManager', error occurred:
		// `PatchOptions.meta.k8s.io "" is invalid: fieldManager: Required value`
		// Related issue: https://github.com/kubernetes/client-go/issues/1036
		opts := metav1.ApplyOptions {
			FieldManager: "application/apply-patch",
		}
		podConfig := &v1.PodApplyConfiguration{}
		podConfig = podConfig.WithAPIVersion("v1")
		podConfig = podConfig.WithKind("Pod")
		podConfig = podConfig.WithName(nameMap[idx])
		podConfig = podConfig.WithSpec(
			&v1.PodSpecApplyConfiguration{
				Containers: []v1.ContainerApplyConfiguration {
					{
						Name: &sat_name,
						Image: &image_name,
						ImagePullPolicy: (*corev1.PullPolicy)(&image_pull_policy),
						Ports: []v1.ContainerPortApplyConfiguration {
							{
								ContainerPort: &port,
							},
						},
						Command: []string {
							"/bin/sh",
							"-c",
						},
						Args: []string {
							fmt.Sprintf(
								"ifconfig eth0:sdneth0 %s netmask 255.255.255.255 up;" + 
								"iptables -t nat -A OUTPUT -d 10.233.0.0/16 -j MARK --set-mark %d;" +
								"iptables -t nat -A POSTROUTING -m mark --mark %d -d 10.233.0.0/16 -j SNAT --to-source %s;" +
								"/podserver", 
								util.GetGlobalIP(uint(idx)),
								idx + 5000,
								idx + 5000,
								util.GetGlobalIP(uint(idx)),
							),
						},
						SecurityContext: &v1.SecurityContextApplyConfiguration{
							Capabilities: &v1.CapabilitiesApplyConfiguration{
								Add: []corev1.Capability{
									"NET_ADMIN",
								},
							},
						},
					},
				},
			},
		)
		_, err := clientset.CoreV1().Pods(namespace).Apply(context.TODO(), podConfig, opts)
		if err != nil {
			return fmt.Errorf("CREATE POD FAILURE: %v", err)
		}
	}

	return nil
}

