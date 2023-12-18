package util

const (
	DEBUG          = true
	POD_IMAGE_NAME = "electronicwaste/podserver"
	POD_IMAGE_TAG  = "v26"
	ThreadNums     = 64
)

var (
	NodeCapacity = map[string]int{
		"node1.dtn.lab":  1,
		"node2.dtn.lab":  1,
		"node3.dtn.lab":  1,
		"node4.dtn.lab":  3,
		"node5.dtn.lab":  3,
		"node6.dtn.lab":  3,
		"node7.dtn.lab":  1,
		"node12.dtn.lab": 4,
		"node13.dtn.lab": 4,
	}

	ImageName 				= POD_IMAGE_NAME + ":" + POD_IMAGE_TAG
	ImagePullPolicy 		= "IfNotPresent"

	FlowPVCName 			= "podserver-wangshao-pvc"
	FlowMountPath 			= "/flow"
	FlowPortName 			= "flow"
	FlowPort int32 			= 5202

	RoutePortName 			= "route"
	RoutePort int32 		= 8080

	PrometheusPortName		= "prometheus"
	PrometheusPort int32	= 2112

)
