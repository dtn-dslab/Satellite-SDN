package link

type MetaData struct {
	Name string `json:"name" yaml:"name"`
}

type ContainerPort struct {
	ContainerPort int32 `json:"containerPort" yaml:"containerPort"`
}

type Container struct {
	Name string `json:"name" yaml:"name"`

	Image string `json:"image" yaml:"image"`

	Command []string `json:"command" yaml:"command"`

	Args []string `json:"args" yaml:"args"`

	Ports []ContainerPort `json:"ports" yaml:"ports"`

	ImagePullPolicy string `json:"imagePullPolicy" yaml:"imagePullPolicy"`
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