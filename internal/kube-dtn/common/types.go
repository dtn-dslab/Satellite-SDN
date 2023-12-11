package common

import "github.com/containernetworking/cni/pkg/types"

type NetConf struct {
	types.NetConf
	Delegate map[string]interface{} `json:"delegate"`
}

type K8sArgs struct {
	types.CommonArgs
	K8S_POD_NAME               types.UnmarshallableString
	K8S_POD_NAMESPACE          types.UnmarshallableString
	K8S_POD_INFRA_CONTAINER_ID types.UnmarshallableString
}
