/*
Copyright 2024.

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

package controller

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	saasv1alpha1 "github.com/agastyadev-a/quick-bytes-k8s-controller.git/api/v1alpha1"
)

// ToDoAppReconciler reconciles a ToDoApp object
type ToDoAppReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=saas.thoughtworks.com,resources=todoapps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=saas.thoughtworks.com,resources=todoapps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=saas.thoughtworks.com,resources=todoapps/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ToDoApp object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *ToDoAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	var todo saasv1alpha1.ToDoApp

	log.Log.Info(req.NamespacedName.String())

	if err := r.Get(ctx, req.NamespacedName, &todo); err == nil {
		// If TodoApp is not found, create a deployment
		if apierrors.IsNotFound(err) {
			log.Log.Error(err, "unable to fetch ToDoApp")
			return ctrl.Result{}, err
		}
		clientk8s := k8sClient()
		deployment := createToDoAppDeployment(todo.Spec.InstallationName, todo.Spec.ImageVersion, todo.Spec.PostgresURI, req.Namespace, todo.Spec.ContainerPort)
		deployTodoApp(clientk8s, deployment)

	} else {
		// If TodoApp is no longer present in namespace, delete the deployment
		log.Log.Info("Todo not found in " + req.NamespacedName.String())
		clientk8s := k8sClient()
		deployment := getDeployment(clientk8s, req.Namespace)
		log.Log.Info("Todo deployment " + deployment.String())
		deleteDeployment(clientk8s, &deployment)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ToDoAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&saasv1alpha1.ToDoApp{}).Owns(&appsv1.Deployment{}).Complete(r)
}
