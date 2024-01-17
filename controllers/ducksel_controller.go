/*
Copyright 2024 Duck5el.

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

	apiv1 "github.com/Duck5el/ducksel-opperator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	errors "k8s.io/apimachinery/pkg/api/errors"
	pkgRuntime "k8s.io/apimachinery/pkg/runtime"
	ctrlRuntime "sigs.k8s.io/controller-runtime"
	client "sigs.k8s.io/controller-runtime/pkg/client"
	controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	log "sigs.k8s.io/controller-runtime/pkg/log"
)

var logger = log.Log.WithName("ducksel_controller")

// DuckselReconciler reconciles a Ducksel object
type DuckselReconciler struct {
	client.Client
	Scheme *pkgRuntime.Scheme
}

//+kubebuilder:rbac:groups=api.my.domain,resources=ducksels,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=api.my.domain,resources=ducksels/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=api.my.domain,resources=ducksels/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Ducksel object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *DuckselReconciler) Reconcile(ctx context.Context, req ctrlRuntime.Request) (ctrlRuntime.Result, error) {

	log := logger

	log.Info("Reconcile called!")

	ducksel := &apiv1.Ducksel{}

	service := GetService(req.Name, req.Namespace, *ducksel)
	deployment := GetDeployment(req.Name, req.Namespace, *ducksel)

	if err := r.Get(ctx, req.NamespacedName, ducksel); err != nil {
		log.Info("Ducksel has been deleted! cleaning up...")

		err = r.Delete(ctx, service)
		if err != nil {
			log.Info("Unable to delete Service, NotFound!")
		} else {
			log.Info("Deleted service!")
		}

		err = r.Delete(ctx, deployment)
		if err != nil {
			log.Info("Unable to delete Deployment, NotFound!")
		} else {
			log.Info("Deleted Deployment!")
		}

		log.Info("Cleanup done!")

		return ctrlRuntime.Result{}, nil
	}

	log.Info("Detected change on crd ducksel with the name:" + req.Name + " In the namespace:" + req.Namespace)

	// Set Ducksel instance as the owner of the Deployment
	if err := controllerutil.SetControllerReference(ducksel, deployment, r.Scheme); err != nil {
		log.Error(err, "unable to set owner reference for Deployment!")
		return ctrlRuntime.Result{}, err
	}

	found := &appsv1.Deployment{}
	err := r.Get(ctx, client.ObjectKey{Name: deployment.Name, Namespace: deployment.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Deployment does not exist, create it
		log.Info("Creating Deployment", "Namespace", deployment.Namespace, "Name", deployment.Name)
		err = r.Create(ctx, deployment)
		if err != nil {
			log.Error(err, "unable to create Deployment!")
		} else {
			log.Info("Created Deplyoment!")
		}
	} else if err != nil {
		log.Error(err, "unable to check if Deployment exists!")
	} else {
		// Update deplyoment to apply possible changes
		err := r.Update(ctx, deployment)
		if err != nil {
			log.Error(err, "Unable to update Deployment!")
		}
	}

	if ducksel.Spec.Service.Enabled {
		log.Info("Service is set to true!")
		// Set Ducksel instance as the owner of the Service
		if err := controllerutil.SetControllerReference(ducksel, service, r.Scheme); err != nil {
			log.Error(err, "unable to set owner reference for Service!")
			return ctrlRuntime.Result{}, err
		}

		found := &corev1.Service{}
		err := r.Get(ctx, client.ObjectKey{Name: service.Name, Namespace: service.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			// Service does not exist, create it
			log.Info("Creating Service", "Namespace", service.Namespace, "Name", service.Name)
			err = r.Create(ctx, service)
			if err != nil {
				log.Error(err, "unable to create Service!")
			}
		} else if err != nil {
			log.Error(err, "unable to check if Service exists!")
			return ctrlRuntime.Result{}, err
		} else {
			log.Info("Do nothing Service already exists!")
		}
	} else {
		log.Info("Service is set to false!")
		// Check if the Service exists
		found := &corev1.Service{}
		err := r.Get(ctx, client.ObjectKey{Name: service.Name, Namespace: service.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			log.Info("Skip deletion, service is already not present!")
		} else if err != nil {
			log.Error(err, "unable to check if Service exists!")
		} else {
			// Delete the Service
			err = r.Delete(ctx, service)
			if err != nil {
				log.Error(err, "unable to delete Service!")
			} else {
				log.Info("Deleted Service", "Namespace", service.Namespace, "Name", service.Name)
			}
		}

	}
	return ctrlRuntime.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DuckselReconciler) SetupWithManager(mgr ctrlRuntime.Manager) error {
	return ctrlRuntime.NewControllerManagedBy(mgr).
		For(&apiv1.Ducksel{}).
		Complete(r)
}
