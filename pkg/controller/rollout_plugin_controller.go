package controller

import (
	"context"
	"fmt"
	"rolloutplugin-controller/api/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// PhasedRolloutReconciler reconciles a PhasedRollout object.
type RolloutPluginReconciler struct {
	client.Client
	Scheme           *runtime.Scheme
	Recorder         record.EventRecorder
	RetryWaitSeconds int
	MaxConcurrent    int
}

func (r *RolloutPluginReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.RolloutPlugin{}).
		Owns(&v1alpha1.RolloutPlugin{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: r.MaxConcurrent}).
		Complete(
			r,
		)
}

func (r *RolloutPluginReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	fmt.Println("Reconcile called")
	var rolloutPlugin v1alpha1.RolloutPlugin

	err := r.Client.Get(ctx, req.NamespacedName, &rolloutPlugin)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	fmt.Println("RolloutPlugin: ", rolloutPlugin)
	rolloutPlugin.Status.Initialized = true

	generation := rolloutPlugin.GetGeneration()
	rolloutPlugin.Status.ObservedGeneration = generation

	if rolloutPlugin.Status.Conditions == nil {
		rolloutPlugin.Status.Conditions = []v1alpha1.Condition{
			{
				Type:               "Initialized",
				Status:             "True",
				LastUpdateTime:     metav1.Now(),
				LastTransitionTime: metav1.Now(),
			},
		}
	}

	err = r.Client.Status().Update(ctx, &rolloutPlugin)
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *RolloutPluginReconciler) SetConditions(rolloutPlugin *v1alpha1.RolloutPlugin) {

	// Set conditions

}
