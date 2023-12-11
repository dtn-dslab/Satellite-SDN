/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	pb "github.com/y-young/kube-dtn/proto/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TopologySpec defines the desired state of Topology
type TopologySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +optional
	Links []Link `json:"links"`
}

// TopologyStatus defines the observed state of Topology
type TopologyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// List of pods that are skipped by local pod
	// +optional
	Skipped []string `json:"skipped"`

	// Source IP of the pod
	// +optional
	SrcIP string `json:"src_ip"`

	// Network namespace of the pod
	// +optional
	NetNs string `json:"net_ns"`

	// Link statuses
	// +optional
	Links []Link `json:"links"`
}

// A complete definition of a p2p link
type Link struct {
	// Local interface name
	LocalIntf string `json:"local_intf"`

	// Local IP address
	// +optional
	// +kubebuilder:validation:Pattern=`^((([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])(\/(3[0-2]|[1-2][0-9]|[0-9]))?)?$`
	LocalIP string `json:"local_ip"`

	// Local MAC address, e.g. 00:00:5e:00:53:01 or 00-00-5e-00-53-01
	// +optional
	// +kubebuilder:validation:Pattern=`^(([0-9A-Fa-f]{2}[:-]){5}[0-9A-Fa-f]{2})?$`
	LocalMAC string `json:"local_mac"`

	// Peer interface name
	PeerIntf string `json:"peer_intf"`

	// Peer IP address
	// +optional
	// +kubebuilder:validation:Pattern=`^((([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])(\/(3[0-2]|[1-2][0-9]|[0-9]))?)?$`
	PeerIP string `json:"peer_ip"`

	// Peer MAC address, e.g. 00:00:5e:00:53:01 or 00-00-5e-00-53-01
	// +optional
	// +kubebuilder:validation:Pattern=`^(([0-9A-Fa-f]{2}[:-]){5}[0-9A-Fa-f]{2})?$`
	PeerMAC string `json:"peer_mac"`

	// Name of the peer pod
	PeerPod string `json:"peer_pod"`

	// Unique identifier of a p2p link
	UID int `json:"uid"`

	// Link properties, latency, bandwidth, etc
	// +optional
	Properties LinkProperties `json:"properties,omitempty"`
}

func (l *Link) ToProto() *pb.Link {
	return &pb.Link{
		PeerPod:    l.PeerPod,
		LocalIntf:  l.LocalIntf,
		PeerIntf:   l.PeerIntf,
		LocalIp:    l.LocalIP,
		PeerIp:     l.PeerIP,
		LocalMac:   l.LocalMAC,
		PeerMac:    l.PeerMAC,
		Uid:        int64(l.UID),
		Properties: l.Properties.ToProto(),
	}
}

// Float percentage between 0 and 100
// +kubebuilder:validation:Pattern=`^(100(\.0+)?|\d{1,2}(\.\d+)?)$`
type Percentage string

// Duration string, e.g. "300ms", "1.5s".
// +kubebuilder:validation:Pattern=`^(\d+(\.\d+)?(ns|us|µs|μs|ms|s|m|h))+$`
type Duration string

type LinkProperties struct {
	// Latency in duration string format, e.g. "300ms", "1.5s".
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	// +optional
	Latency Duration `json:"latency,omitempty"`

	// Latency correlation in float percentage
	// +optional
	LatencyCorr Percentage `json:"latency_corr,omitempty"`

	// Jitter in duration string format, e.g. "300ms", "1.5s".
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	// +optional
	Jitter Duration `json:"jitter,omitempty"`

	// Loss rate in float percentage
	// +optional
	Loss Percentage `json:"loss,omitempty"`

	// Loss correlation in float percentage
	// +optional
	LossCorr Percentage `json:"loss_corr,omitempty"`

	// Bandwidth rate limit, e.g. 1000(bit/s), 100kbit, 100Mbps, 1Gibps.
	// For more information, refer to https://man7.org/linux/man-pages/man8/tc.8.html.
	// +optional
	// +kubebuilder:validation:Pattern=`^\d+(\.\d+)?([KkMmGg]i?)?(bit|bps)?$`
	Rate string `json:"rate,omitempty"`

	// Gap every N packets
	// +optional
	// +kubebuilder:validation:Minimum=0
	Gap uint32 `json:"gap,omitempty"`

	// Duplicate rate in float percentage
	// +optional
	Duplicate Percentage `json:"duplicate,omitempty"`

	// Duplicate correlation in float percentage
	// +optional
	DuplicateCorr Percentage `json:"duplicate_corr,omitempty"`

	// Reorder probability in float percentage
	// +optional
	ReorderProb Percentage `json:"reorder_prob,omitempty"`

	// Reorder correlation in float percentage
	// +optional
	ReorderCorr Percentage `json:"reorder_corr,omitempty"`

	// Corrupt probability in float percentage
	// +optional
	CorruptProb Percentage `json:"corrupt_prob,omitempty"`

	// Corrupt correlation in float percentage
	// +optional
	CorruptCorr Percentage `json:"corrupt_corr,omitempty"`
}

func (p *LinkProperties) ToProto() *pb.LinkProperties {
	return &pb.LinkProperties{
		Latency:       string(p.Latency),
		LatencyCorr:   string(p.LatencyCorr),
		Jitter:        string(p.Jitter),
		Loss:          string(p.Loss),
		LossCorr:      string(p.LossCorr),
		Rate:          p.Rate,
		Gap:           p.Gap,
		Duplicate:     string(p.Duplicate),
		DuplicateCorr: string(p.DuplicateCorr),
		ReorderProb:   string(p.ReorderProb),
		ReorderCorr:   string(p.ReorderCorr),
		CorruptProb:   string(p.CorruptProb),
		CorruptCorr:   string(p.CorruptCorr),
	}
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Topology is the Schema for the topologies API
type Topology struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TopologySpec   `json:"spec,omitempty"`
	Status TopologyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TopologyList contains a list of Topology
type TopologyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Topology `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Topology{}, &TopologyList{})
}
