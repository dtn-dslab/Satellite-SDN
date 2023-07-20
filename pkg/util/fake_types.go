package util

// These types are only used for generating yaml file.

type MetaData struct {
	Name string `json:"name" yaml:"name"`
}

type ContainerPort struct {
	ContainerPort int32 `json:"containerPort" yaml:"containerPort"`
}

type Capability string

type Capabilities struct {
	// Added capabilities
	// +optional
	Add []Capability `json:"add,omitempty" yaml:"add,omitempty"`
	// Removed capabilities
	// +optional
	Drop []Capability `json:"drop,omitempty" yaml:"drop,omitempty"`
}

type SecurityContext struct {
	// The capabilities to add/drop when running containers.
	// Defaults to the default set of capabilities granted by the container runtime.
	// Note that this field cannot be set when spec.os.name is windows.
	// +optional
	Capabilities Capabilities `json:"capabilities" yaml:"capabilities"`
}

type Container struct {
	Name string `json:"name" yaml:"name"`

	Image string `json:"image" yaml:"image"`

	Command []string `json:"command" yaml:"command"`

	Args []string `json:"args" yaml:"args"`

	Ports []ContainerPort `json:"ports" yaml:"ports"`

	ImagePullPolicy string `json:"imagePullPolicy" yaml:"imagePullPolicy"`

	SecurityContext SecurityContext `json:"securityContext" yaml:"securityContext"`
}

type PodSpec struct {
	Containers []Container `json:"containers,omitempty" yaml:"containers,omitempty"`

	NodeSelector map[string]string `json:"nodeSelector,omitempty" yaml:"nodeSelector,omitempty"`
}

type Pod struct {
	APIVersion string `json:"apiVersion" yaml:"apiVersion"`

	MetaData MetaData `json:"metadata" yaml:"metadata"`

	Kind string `json:"kind" yaml:"kind"`

	Spec PodSpec `json:"spec" yaml:"spec"`
}

type PodList struct {
	APIVersion string `json:"apiVersion" yaml:"apiVersion"`

	Kind string `json:"kind" yaml:"kind"`

	Items []Pod `json:"items" yaml:"items"`
}

type Link struct {
	UID int `json:"uid" yaml:"uid"`

	LocalIntf string `json:"local_intf" yaml:"local_intf"`

	LocalIP string `json:"local_ip" yaml:"local_ip"`

	PeefIntf string `json:"peer_intf" yaml:"peer_intf"`

	PeerIP string `json:"peer_ip" yaml:"peer_ip"`

	PeerPod string `json:"peer_pod" yaml:"peer_pod"`
}

type TopologySpec struct {
	Links []Link `json:"links" yaml:"links"`
}

type Topology struct {
	APIVersion string `json:"apiVersion" yaml:"apiVersion"`

	Kind string `json:"kind" yaml:"kind"`

	MetaData MetaData `json:"metadata" yaml:"metadata"`

	Spec TopologySpec `json:"spec" yaml:"spec"`
}

type TopologyList struct {
	APIVersion string `json:"apiVersion" yaml:"apiVersion"`

	Kind string `json:"kind" yaml:"kind"`

	Items []Topology `json:"items" yaml:"items"`
}

type SubPath struct {
	Name string `json:"name" yaml:"name"`

	TargetIP string `json:"targetip" yaml:"targetip"`

	NextIP string `json:"nextip" yaml:"nextip"`
}

type RouteSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	PodIP string `json:"podip" yaml:"podip"`

	SubPaths []SubPath `json:"subpaths" yaml:"subpaths"`
}

type Route struct {
	APIVersion string `json:"apiVersion" yaml:"apiVersion"`

	MetaData MetaData `json:"metadata" yaml:"metadata"`

	Kind string `json:"kind" yaml:"kind"`

	Spec RouteSpec `json:"spec" yaml:"spec"`
}

type RouteList struct {
	APIVersion string `json:"apiVersion" yaml:"apiVersion"`

	Kind string `json:"kind" yaml:"kind"`

	Items []Route `json:"items" yaml:"items"`
}
