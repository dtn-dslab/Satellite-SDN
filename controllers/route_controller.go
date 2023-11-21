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
	"net/http"
	"reflect"
	"strings"

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

	log.Info("route %s -> add: %v, del: %v, update: %v", route.Name, add, del, update)

	if err := r.DelSubpaths(ctx, route.Spec.PodIP, del); err != nil {
		log.Error(err, "Failed to delete subpaths")
		return ctrl.Result{}, err
	}

	if err := r.AddSubpaths(ctx, route.Spec.PodIP, add); err != nil {
		log.Error(err, "Failed to add subpaths")
		return ctrl.Result{}, err
	}

	if err := r.UpdateSubpaths(ctx, route.Spec.PodIP, update); err != nil {
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

func (r *RouteReconciler) AddSubpaths(ctx context.Context, podIP string, subpaths []sdnv1.SubPath) error {
	log := log.FromContext(ctx)

	// No subpath in subpaths: return nil.
	if len(subpaths) == 0 {
		log.Info("No subpath in add.")
		return nil
	}

	postURL := "http://" + podIP + ":8080" + "/route/apply"
	jsonVal, _ := json.Marshal(subpaths)
	resp, err := http.Post(
		postURL,
		"application/json",
		strings.NewReader(string(jsonVal)),
	)
	if err != nil {
		log.Info("Post Error", "PostURL", postURL, "Error", err)
	} else if resp != nil && resp.StatusCode != http.StatusOK {
		log.Info("Add subpaths failed", "StatusCode", resp.StatusCode, "PostURL", postURL)
	}

	return err
}

func (r *RouteReconciler) DelSubpaths(ctx context.Context, podIP string, subpaths []sdnv1.SubPath) error {
	log := log.FromContext(ctx)

	// No subpath in subpaths: return nil.
	if len(subpaths) == 0 {
		log.Info("No subpath in del.")
		return nil
	}

	postURL := "http://" + podIP + ":8080" + "/route/del"
	jsonVal, _ := json.Marshal(subpaths)
	resp, err := http.Post(
		postURL,
		"application/json",
		strings.NewReader(string(jsonVal)),
	)
	if err != nil {
		log.Info("Post Error", "PostURL", postURL, "Error", err)
	} else if resp != nil && resp.StatusCode != http.StatusOK {
		log.Info("Delete subpaths failed", "StatusCode", resp.StatusCode, "PostURL", postURL)
	}

	return err
}

func (r *RouteReconciler) UpdateSubpaths(ctx context.Context, podIP string, subpaths []sdnv1.SubPath) error {
	log := log.FromContext(ctx)

	// No subpath in subpaths: return nil.
	if len(subpaths) == 0 {
		log.Info("No subpath in update.")
		return nil
	}

	postURL := "http://" + podIP + ":8080" + "/route/update"
	jsonVal, _ := json.Marshal(subpaths)
	resp, err := http.Post(
		postURL,
		"application/json",
		strings.NewReader(string(jsonVal)),
	)
	if err != nil {
		log.Info("Post Error", "PostURL", postURL, "Error", err)
	} else if resp != nil && resp.StatusCode != http.StatusOK {
		log.Info("Update subpaths failed", "StatusCode", resp.StatusCode, "PostURL", postURL)
	}

	return err
}

// This function will calculate the differences between old subpaths and new subpaths,
// and return three subpath arrays:
// 1. add for new subpaths
// 2. del for subpaths that need to be deleted
// 3. update for supaths that need to be updated(nextip changed)
func (r *RouteReconciler) CalcDiff(old []sdnv1.SubPath, new []sdnv1.SubPath) (add []sdnv1.SubPath, del []sdnv1.SubPath, update []sdnv1.SubPath) {
	for _, oldSubpath := range old {
		found := false
		for _, newSubpath := range new {
			if oldSubpath.Name == newSubpath.Name {
				found = true
				if oldSubpath.NextIP != newSubpath.NextIP {
					update = append(update, newSubpath)
				} else if oldSubpath.TargetIP != newSubpath.TargetIP {
					del = append(del, oldSubpath)
					add = append(add, newSubpath)
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

