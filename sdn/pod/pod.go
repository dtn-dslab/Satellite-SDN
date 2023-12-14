package pod

import (
	"context"
	"fmt"
	"log"
	"sync"
	"ws/dtn-satellite-sdn/sdn/util"

	corev1 "k8s.io/api/core/v1"
	// "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
)

type PodMetadata struct {
	IndexUUIDMap  map[int]string
	UserIdxMin int
	UserNum    int
}

func ParseLabels(index int, meta *PodMetadata) map[string]string {
	result := map[string]string {
		"k8s-app": "iperf",
	}
	if index >= meta.UserIdxMin && 
		index < meta.UserIdxMin + meta.UserNum {
		if index < meta.UserIdxMin + meta.UserNum / 2 {
			result["type"] = "client"
		} else {
			result["type"] = "server"
		}
	} 
	return result
}

func ParseArgs(index int, meta *PodMetadata) string {
	// TODO(yy): remove `PODNAME` environment variable,
	// replace it with command `hostname`.
	result := fmt.Sprintf(
		"export PODNAME=%s;" +
		"./start.sh %s %d",
		meta.IndexUUIDMap[index], 
		util.GetGlobalIP(uint(index)), index + 5000,
	)
	if index >= meta.UserIdxMin && 
		index < meta.UserIdxMin + meta.UserNum {
		if index < meta.UserIdxMin + meta.UserNum / 2 {
			serverIP := util.GetGlobalIP(uint(index + meta.UserNum / 2))
			result = fmt.Sprintf("export SERVERIP=%s;", serverIP) + result
		} 
	}
	return result
}

// Function: PodSyncLoopV2
// Description: Apply pods in which low-orbit satellites in one group are deployed to the same node.
// 1. indexUUIDMap: node's index -> node's uuid.
// 2. uuiAllocNodeMap: node's uuid -> allocNode's name(node1.dtn.lab), only stores low-orbit satellites's pairs.
func PodSyncLoop(meta *PodMetadata, uuidAllocNodeMap map[string]string) error {
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
	for index, uuid := range meta.IndexUUIDMap {
		podConfig := &v1.PodApplyConfiguration{}
		podConfig = podConfig.WithAPIVersion("v1")
		podConfig = podConfig.WithKind("Pod")
		podConfig = podConfig.WithName(uuid)
		podConfig = podConfig.WithLabels(ParseLabels(index, meta))
		podConfig = podConfig.WithSpec(
			&v1.PodSpecApplyConfiguration{
				Containers: []v1.ContainerApplyConfiguration{
					{
						Name:            &uuid,
						Image:           &util.ImageName,
						ImagePullPolicy: (*corev1.PullPolicy)(&util.ImagePullPolicy),
						Ports: []v1.ContainerPortApplyConfiguration{
							{
								Name:			&util.RoutePortName,
								ContainerPort: 	&util.RoutePort,
							},
							{
								Name:          	&util.PrometheusPortName,
								ContainerPort: 	&util.PrometheusPort,
							},
							{
								Name:			&util.FlowPortName,
								ContainerPort: 	&util.FlowPort,
							},
						},
						VolumeMounts: []v1.VolumeMountApplyConfiguration{
							{
								MountPath: 		&util.FlowMountPath,
								Name:     		&util.FlowPVCName,
							},
						},
						Command: []string{
							"/bin/sh",
							"-c",
						},
						Args: []string{
							ParseArgs(index, meta),
						},
						SecurityContext: &v1.SecurityContextApplyConfiguration{
							Capabilities: &v1.CapabilitiesApplyConfiguration{
								Add: []corev1.Capability{
									"NET_ADMIN",
								},
							},
						},
						// Resources: &v1.ResourceRequirementsApplyConfiguration{
						// 	Limits: &corev1.ResourceList{
						// 		corev1.ResourceCPU: *resource.NewMilliQuantity(500, resource.DecimalSI),
						// 	},
						// 	Requests: &corev1.ResourceList{
						// 		corev1.ResourceCPU: *resource.NewMilliQuantity(250, resource.DecimalSI),
						// 	},
						// },
					},
				},
				Volumes: []v1.VolumeApplyConfiguration{
					{
						Name: &util.FlowPVCName,
						VolumeSourceApplyConfiguration: v1.VolumeSourceApplyConfiguration{
							PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSourceApplyConfiguration{
								ClaimName: &util.FlowPVCName,
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
	opts := metav1.ApplyOptions{
		FieldManager: "application/apply-patch",
	}
	// Apply pods
	wg := new(sync.WaitGroup)
	wg.Add(util.ThreadNums)
	for threadId := 0; threadId < util.ThreadNums; threadId++ {
		go func(id int) {
			for podId := id; podId < len(podList); podId += util.ThreadNums {
				pod := podList[podId]
				if _, err := clientset.CoreV1().Pods(namespace).Apply(context.TODO(), pod, opts); err != nil {
					log.Fatalf("apply pod error: %v", err)
				}
			}
			wg.Done()
		}(threadId)
	}
	wg.Wait()
	return nil
}
