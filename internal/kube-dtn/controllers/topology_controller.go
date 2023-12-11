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

package controllers

import (
	"context"
	"reflect"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1 "github.com/y-young/kube-dtn/api/v1"
	"github.com/y-young/kube-dtn/common"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/y-young/kube-dtn/proto/v1"
)

// TopologyReconciler reconciles a Topology object
type TopologyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=y-young.github.io,resources=topologies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=y-young.github.io,resources=topologies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=y-young.github.io,resources=topologies/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Topology object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *TopologyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var topology v1.Topology
	if err := r.Get(ctx, req.NamespacedName, &topology); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Topology deleted")
		} else {
			log.Error(err, "Unable to fetch Topology")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Spec remains the same, nothing to do
	if reflect.DeepEqual(topology.Status.Links, topology.Spec.Links) {
		return ctrl.Result{}, nil
	}

	if topology.Status.Links == nil {
		log.Info("Topology created")
		// Saw topology for the first time, assume all links in spec have been set up,
		// we'll copy initial links to status later
	} else {
		add, del, propertiesChanged := r.CalcDiff(topology.Status.Links, topology.Spec.Links)
		log.Info("Topology changed", "add", add, "del", del, "update", propertiesChanged)

		if err := r.DelLinks(ctx, &topology, del); err != nil {
			log.Error(err, "Failed to delete links")
			return ctrl.Result{}, err
		}

		if err := r.AddLinks(ctx, &topology, add); err != nil {
			log.Error(err, "Failed to add links")
			return ctrl.Result{}, err
		}

		if err := r.UpdateLinks(ctx, &topology, propertiesChanged); err != nil {
			log.Error(err, "Failed to update links")
			return ctrl.Result{}, err
		}
	}

	// Since kubedtn CNI will also update the status, retry on conflict
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var topology v1.Topology
		if err := r.Get(ctx, req.NamespacedName, &topology); err != nil {
			if apierrors.IsNotFound(err) {
				log.Info("Topology deleted")
			} else {
				log.Error(err, "Unable to fetch Topology")
			}
			return client.IgnoreNotFound(err)
		}

		topology.Status.Links = topology.Spec.Links
		return r.Status().Update(ctx, &topology)
	})

	if err != nil {
		log.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *TopologyReconciler) AddLinks(ctx context.Context, topology *v1.Topology, links []v1.Link) error {
	if len(links) == 0 {
		return nil
	}

	log := log.FromContext(ctx)

	conn, err := ConnectDaemon(ctx, topology.Status.SrcIP)
	if err != nil {
		return err
	}
	defer conn.Close()
	kubedtnClient := pb.NewLocalClient(conn)

	result, err := kubedtnClient.AddLinks(ctx, &pb.LinksBatchQuery{
		LocalPod: &pb.Pod{
			Name:   topology.Name,
			SrcIp:  topology.Status.SrcIP,
			NetNs:  topology.Status.NetNs,
			KubeNs: topology.Namespace,
		},
		Links: common.Map(links, func(link v1.Link) *pb.Link { return link.ToProto() }),
	})
	if err != nil || !result.GetResponse() {
		log.Error(err, "Failed to add links")
		return err
	}
	log.Info("Successfully added links", "links", common.Map(links, func(link v1.Link) int { return link.UID }))
	return nil
}

func (r *TopologyReconciler) DelLinks(ctx context.Context, topology *v1.Topology, links []v1.Link) error {
	if len(links) == 0 {
		return nil
	}

	log := log.FromContext(ctx)

	conn, err := ConnectDaemon(ctx, topology.Status.SrcIP)
	if err != nil {
		return err
	}
	defer conn.Close()
	kubedtnClient := pb.NewLocalClient(conn)

	result, err := kubedtnClient.DelLinks(ctx, &pb.LinksBatchQuery{
		LocalPod: &pb.Pod{
			Name:   topology.Name,
			SrcIp:  topology.Status.SrcIP,
			NetNs:  topology.Status.NetNs,
			KubeNs: topology.Namespace,
		},
		Links: common.Map(links, func(link v1.Link) *pb.Link { return link.ToProto() }),
	})
	if err != nil || !result.GetResponse() {
		log.Error(err, "Failed to delete links")
		return err
	}
	log.Info("Successfully deleted links", "links", common.Map(links, func(link v1.Link) int { return link.UID }))
	return nil
}

func (r *TopologyReconciler) UpdateLinks(ctx context.Context, topology *v1.Topology, links []v1.Link) error {
	if len(links) == 0 {
		return nil
	}

	log := log.FromContext(ctx)

	conn, err := ConnectDaemon(ctx, topology.Status.SrcIP)
	if err != nil {
		return err
	}
	defer conn.Close()
	kubedtnClient := pb.NewLocalClient(conn)

	result, err := kubedtnClient.UpdateLinks(ctx, &pb.LinksBatchQuery{
		LocalPod: &pb.Pod{
			Name:   topology.Name,
			SrcIp:  topology.Status.SrcIP,
			NetNs:  topology.Status.NetNs,
			KubeNs: topology.Namespace,
		},
		Links: common.Map(links, func(link v1.Link) *pb.Link { return link.ToProto() }),
	})
	if err != nil || !result.GetResponse() {
		log.Error(err, "Failed to delete link")
		return err
	}
	log.Info("Successfully updated links", "links", common.Map(links, func(link v1.Link) int { return link.UID }))
	return err
}

// Calculate difference between two old links and new links, returns a list of links to be added and a list of links to be deleted
func (r *TopologyReconciler) CalcDiff(old []v1.Link, new []v1.Link) (add []v1.Link, del []v1.Link, propertiesChanged []v1.Link) {
	for _, oldLink := range old {
		found := false
		for _, newLink := range new {
			if EqualWithoutProperties(oldLink, newLink) {
				found = true
				if !reflect.DeepEqual(oldLink.Properties, newLink.Properties) {
					propertiesChanged = append(propertiesChanged, newLink)
				}
				break
			}
		}
		if !found {
			del = append(del, oldLink)
		}
	}

	for _, newLink := range new {
		found := false
		for _, oldLink := range old {
			if EqualWithoutProperties(oldLink, newLink) {
				found = true
				break
			}
		}
		if !found {
			add = append(add, newLink)
		}
	}
	return
}

func ConnectDaemon(ctx context.Context, ip string) (*grpc.ClientConn, error) {
	log := log.FromContext(ctx)
	daemonAddr := "passthrough:///" + ip + ":" + common.DefaultPort
	conn, err := grpc.Dial(daemonAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error(err, "Failed to connect to daemon", "address", daemonAddr)
		return nil, err
	}
	return conn, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TopologyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Topology{}).
		Complete(r)
}

// EqualWithoutProperties compares two links without comparing link properties
func EqualWithoutProperties(a, b v1.Link) bool {
	return a.LocalIntf == b.LocalIntf &&
		a.LocalIP == b.LocalIP &&
		a.LocalMAC == b.LocalMAC &&
		a.PeerIntf == b.PeerIntf &&
		a.PeerIP == b.PeerIP &&
		a.PeerMAC == b.PeerMAC &&
		a.PeerPod == b.PeerPod &&
		a.UID == b.UID
}
