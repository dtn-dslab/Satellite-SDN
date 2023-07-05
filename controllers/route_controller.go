/*
Copyright 2023.

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
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os/exec"
	"path"
	"reflect"
	"strings"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	sdnv1 "ws/dtn-satellite-sdn/api/v1"
)

// RouteReconciler reconciles a Route object
type RouteReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=sdn.dtn-satellite-sdn,resources=routes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=sdn.dtn-satellite-sdn,resources=routes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=sdn.dtn-satellite-sdn,resources=routes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Route object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *RouteReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var route sdnv1.Route
	if err := r.Get(ctx, req.NamespacedName, &route); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Route deleted")
		} else {
			log.Error(err, "Unable to fetch route")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Spec remains the same, nothing to do
	if reflect.DeepEqual(route.Status.SubPaths, route.Spec.SubPaths) {
		return ctrl.Result{}, nil
	}

	// Update route table in the pod
	var add, del, update []sdnv1.SubPath
	if route.Status.SubPaths == nil {
		add = route.Spec.DeepCopy().SubPaths
	} else {
		add, del, update = r.CalcDiff(route.Status.SubPaths, route.Spec.SubPaths)
	}

	if err := r.DelSubpaths(ctx, route.Name, del); err != nil {
		log.Error(err, "Failed to delete subpaths")
		return ctrl.Result{}, err
	}

	if err := r.AddSubpaths(ctx, route.Name, add); err != nil {
		log.Error(err, "Failed to add subpaths")
		return ctrl.Result{}, err
	}

	if err := r.UpdateSubpaths(ctx, route.Name, update); err != nil {
		log.Error(err, "Failed to update subpaths")
		return ctrl.Result{}, err
	}

	// Update route object in k8s cluster
	route.Status.SubPaths = route.Spec.DeepCopy().SubPaths
	if err := r.Status().Update(ctx, &route); err != nil {
		log.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *RouteReconciler) AddSubpaths(ctx context.Context, podName string, subpaths []sdnv1.SubPath) error {
	log := log.FromContext(ctx)

	var podIP string
	var err error
	for podIP, err = GetPodIP(podName); err != nil; {
		log.Info("Retry!")
		duration := 3000 + rand.Int31() % 2000
		time.Sleep(time.Duration(duration))
		podIP, err = GetPodIP(podName)
	}

	postURL :=  path.Join(podIP, "/route/apply")
	jsonVal, _ := json.Marshal(subpaths)
	resp, err := http.Post(
		postURL, 
		"application/json",
		strings.NewReader(string(jsonVal)),
	)
	if err != nil {
		log.Info("StatusCode is %d:%v", resp.StatusCode, err)
		return err
	}

	return nil
}


func (r *RouteReconciler) DelSubpaths(ctx context.Context, podName string, subpaths []sdnv1.SubPath) error {
	log := log.FromContext(ctx)

	var podIP string
	var err error
	for podIP, err = GetPodIP(podName); err != nil; {
		log.Info("Retry!")
		duration := 3000 + rand.Int31() % 2000
		time.Sleep(time.Duration(duration))
		podIP, err = GetPodIP(podName)
	}

	postURL :=  path.Join(podIP, "/route/del")
	jsonVal, _ := json.Marshal(subpaths)
	resp, err := http.Post(
		postURL, 
		"application/json",
		strings.NewReader(string(jsonVal)),
	)
	if err != nil {
		log.Info("StatusCode is %d:%v", resp.StatusCode, err)
		return err
	}

	return nil
}

func (r *RouteReconciler) UpdateSubpaths(ctx context.Context, podName string, subpaths []sdnv1.SubPath) error {
	log := log.FromContext(ctx)

	var podIP string
	var err error
	for podIP, err = GetPodIP(podName); err != nil; {
		log.Info("Retry!")
		duration := 3000 + rand.Int31() % 2000
		time.Sleep(time.Duration(duration))
		podIP, err = GetPodIP(podName)
	}

	postURL :=  path.Join(podIP, "/route/update")
	jsonVal, _ := json.Marshal(subpaths)
	resp, err := http.Post(
		postURL, 
		"application/json",
		strings.NewReader(string(jsonVal)),
	)
	if err != nil {
		log.Info("StatusCode is %d:%v", resp.StatusCode, err)
		return err
	}

	return nil
}

func (r *RouteReconciler) CalcDiff(old []sdnv1.SubPath, new []sdnv1.SubPath) (add []sdnv1.SubPath, del []sdnv1.SubPath, update []sdnv1.SubPath) {
	for _, oldSubpath := range old {
		found := false
		for _, newSubpath := range new {
			if oldSubpath.Name == newSubpath.Name {
				found = true
				if oldSubpath.NextIP != newSubpath.NextIP {
					update = append(update, newSubpath)
				}
				break
			}
		}
		if !found {
			del = append(del, oldSubpath)
		}
	}

	for _, newSubpath := range new {
		found := false
		for _, oldSubpath := range old {
			if oldSubpath.Name == newSubpath.Name {
				found = true
				break
			}
		}
		if !found {
			add = append(add, newSubpath)
		}
	}

	return
}

// SetupWithManager sets up the controller with the Manager.
func (r *RouteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sdnv1.Route{}).
		Complete(r)
}

func GetPodIP(podName string) (string, error) {
	cmd := exec.Command("kubectl", "get", "pods", "-o", "wide")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Executing kubectl get pods failed: %v\n", err)
	}

	lines := strings.Split(string(output), "\n")[1:]
	for _, line := range lines {
		blocks := strings.Split(line, " ")
		newBlocks := []string{}
		for _, block := range blocks {
			if block != "" {
				newBlocks = append(newBlocks, block)
			} 
		}
		if newBlocks[0] == podName && newBlocks[5] != "<none>"{
			return newBlocks[5], nil
		}
	}

	return "", fmt.Errorf("Can't find pod: %s\n", podName)
}
