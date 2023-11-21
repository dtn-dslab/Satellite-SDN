package pod

import (
	"context"
	"fmt"

	"ws/dtn-satellite-sdn/sdn/util"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
)

func PodSyncLoop(nameMap map[int]string) error {
	// get clientset
	clientset, err := util.GetClientset()
	if err != nil {
		return fmt.Errorf("CREATE CLIENTSET ERROR: %v", err)
	}

	// get current namespace
	namespace, err := util.GetNamespace()
	if err != nil {
		return fmt.Errorf("GET NAMESPACE ERROR: %v", err)
	}
	
	// construct pods
	// TODO(ws): Store pod name in database
	for idx := 0; idx < len(nameMap); idx++ {
		sat_name := nameMap[idx]
		image_name := "electronicwaste/podserver:v10"
		image_pull_policy := "IfNotPresent"
		var port, prometheus_port int32 = 8080, 2112
		prometheus_port_name := "prometheus"
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
		podConfig = podConfig.WithName(sat_name)
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
							{
								Name: &prometheus_port_name,
								ContainerPort: &prometheus_port,
							},
						},
						Command: []string {
							"/bin/sh",
							"-c",
						},
						Args: []string {
							fmt.Sprintf(
								"export POD_IDX=%d;" +
								"export GLOBAL_IP=%s;" +
								"/bootstrap.sh", 
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

