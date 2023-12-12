package util

const (
	DEBUG          = false
	POD_IMAGE_NAME = "electronicwaste/podserver"
	POD_IMAGE_TAG  = "v23"
	ThreadNums     = 64
)

var (
	NodeCapacity = map[string]int{
		"node1.dtn.lab":  1,
		"node2.dtn.lab":  1,
		"node4.dtn.lab":  3,
		"node5.dtn.lab":  3,
		"node6.dtn.lab":  3,
		"node7.dtn.lab":  1,
		"node12.dtn.lab": 4,
		"node13.dtn.lab": 4,
	}
)
