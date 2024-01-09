package util

const (
	POD_IMAGE_NAME = "electronicwaste/podserver"
	POD_IMAGE_TAG  = "v29"
	ThreadNums     = 64
)

var (
	NodeCapacity = map[string]int{
		"node1.dtn.lab":  1,
		"node2.dtn.lab":  1,
		"node3.dtn.lab":  1,
		"node4.dtn.lab":  2,
		"node5.dtn.lab":  2,
		"node6.dtn.lab":  2,
		"node7.dtn.lab":  1,
		"node11.dtn.lab": 1,
		"node12.dtn.lab": 4,
		"node13.dtn.lab": 4,
		"node14.dtn.lab": 4,
		"node15.dtn.lab": 4,
		"node16.dtn.lab": 4,
		"node17.dtn.lab": 4,
		"node18.dtn.lab": 4,
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
