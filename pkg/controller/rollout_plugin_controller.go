package controller

import (
	"context"
	"fmt"

	"github.com/aburan28/rolloutplugin-controller/api/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// PhasedRolloutReconciler reconciles a PhasedRollout object.
type RolloutPluginController struct {
	client.Client
	Scheme           *runtime.Scheme
	Recorder         record.EventRecorder
	RetryWaitSeconds int
	MaxConcurrent    int
}

func (r *RolloutPluginController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.RolloutPlugin{}).
		Owns(&v1alpha1.RolloutPlugin{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: r.MaxConcurrent}).
		Complete(
			r,
		)
}

func NewRolloutPluginController(client client.Client, scheme *runtime.Scheme, recorder record.EventRecorder, retryWaitSeconds int, maxConcurrent int) *RolloutPluginController {
	return &RolloutPluginController{
		Client:           client,
		Scheme:           scheme,
		RetryWaitSeconds: retryWaitSeconds,
		Recorder:         recorder,
		MaxConcurrent:    maxConcurrent,
	}
}

func (r *RolloutPluginController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	fmt.Println("Reconcile called")
	var rolloutPlugin v1alpha1.RolloutPlugin

	err := r.Client.Get(ctx, req.NamespacedName, &rolloutPlugin)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if !rolloutPlugin.Status.Initialized {
		// Initialize the plugin
		fmt.Println("Initializing plugin")
		// pluginName := rolloutPlugin.Spec.Plugin.Name
		// client := pluginClients.client[pluginName]
		// if client == nil {
		// 	_, err = pluginClients.startPlugin(pluginName)

		// }
		// // download the plugin
		// if err != nil {
		// 	return ctrl.Result{}, fmt.Errorf("unable to find plugin (%s): %w", rolloutPlugin.Spec.Plugin.Name, err)
		// }
		// pluginClients.startPlugin(rolloutPlugin.Spec.Plugin.Name)
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

func (r *RolloutPluginController) SetConditions(rolloutPlugin *v1alpha1.RolloutPlugin) {

	// Set conditions

}

func (r *RolloutPluginController) Run(ctx context.Context) error {

	return nil
}
