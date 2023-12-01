package pod

import (
	"context"
	"fmt"
	"ws/dtn-satellite-sdn/sdn/util"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
)

// Function: PodSyncLoopV2
// Description: Apply pods in which low-orbit satellites in one group are deployed to the same node.
// 1. indexUUIDMap: node's index -> node's uuid.
// 2. uuiAllocNodeMap: node's uuid -> allocNode's name(node1.dtn.lab), only stores low-orbit satellites's pairs.
func PodSyncLoopV2(indexUUIDMap map[int]string, uuidAllocNodeMap map[string]string) error {
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

	// Construct Pods
	podList := []*v1.PodApplyConfiguration{}
	for index, uuid := range indexUUIDMap {
		sat_name := uuid
		image_name := "electronicwaste/podserver:v12"
		image_pull_policy := "IfNotPresent"
		flowpvc := "podserver-wangshao-pvc"
		flow_mount_path := "/flow"
		prometheus_port_name := "prometheus"
		var port, prometheus_port, flow_port int32 = 8080, 2112, 5202
		labels := map[string]string{
			"k8s-app": "iperf",
		}
		podConfig := &v1.PodApplyConfiguration{}
		podConfig = podConfig.WithAPIVersion("v1")
		podConfig = podConfig.WithKind("Pod")
		podConfig = podConfig.WithName(sat_name)
		podConfig = podConfig.WithLabels(labels)
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
							{
								ContainerPort: &flow_port,
							},
						},
						VolumeMounts: []v1.VolumeMountApplyConfiguration{
							{
								MountPath: &flow_mount_path,
								Name:      &flowpvc,
							},
						},
						Command: []string {
							"/bin/sh",
							"-c",
						},
						Args: []string {
							fmt.Sprintf(
								"export PODNAME=%s;" +
								"./start.sh %s %d", 
								uuid, util.GetGlobalIP(uint(index)), index+5000,
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
				Volumes: []v1.VolumeApplyConfiguration{
					{
						Name: &flowpvc,
						VolumeSourceApplyConfiguration: v1.VolumeSourceApplyConfiguration{
							PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSourceApplyConfiguration{
								ClaimName: &flowpvc,
							},
						},
					},
				},
			},
		)
		if allocNode, ok := uuidAllocNodeMap[uuid]; ok {
			podConfig.Spec.NodeName = &allocNode
		}
		podList = append(podList, podConfig)
	}

	// TODO(ws): figure out why FieldManager is needed
	// When we delete key 'FieldManager', error occurred:
	// `PatchOptions.meta.k8s.io "" is invalid: fieldManager: Required value`
	// Related issue: https://github.com/kubernetes/client-go/issues/1036
	opts := metav1.ApplyOptions {
		FieldManager: "application/apply-patch",
	}
	// Apply pods
	for _, pod := range podList {
		if _, err := clientset.CoreV1().Pods(namespace).Apply(context.TODO(), pod, opts); err != nil {
			return fmt.Errorf("apply pod error: %v", err)
		}
	}
	return nil
}