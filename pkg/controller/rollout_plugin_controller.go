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
	if !rolloutPlugin.Status.Initialized {
		// Initialize the plugin
		fmt.Println("Initializing plugin")
		pluginName := rolloutPlugin.Spec.Plugin.Name
		client := pluginClients.client[pluginName]
		if client == nil {
			_, err = pluginClients.startPlugin(pluginName)

		}
		// download the plugin
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("unable to find plugin (%s): %w", rolloutPlugin.Spec.Plugin.Name, err)
		}
		pluginClients.startPlugin(rolloutPlugin.Spec.Plugin.Name)
		return ctrl.Result{}, nil
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
