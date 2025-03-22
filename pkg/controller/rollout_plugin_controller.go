package controller

import (
	"context"
	"fmt"

	goPlugin "github.com/hashicorp/go-plugin"

	"github.com/aburan28/rolloutplugin-controller/api/v1alpha1"
	"github.com/aburan28/rolloutplugin-controller/pkg/plugin/pluginclient"
	"github.com/aburan28/rolloutplugin-controller/pkg/plugin/rpc"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// PhasedRolloutReconciler reconciles a PhasedRollout object.
type RolloutPluginController struct {
	k8sclient.Client
	Scheme           *runtime.Scheme
	Recorder         record.EventRecorder
	RetryWaitSeconds int
	MaxConcurrent    int
	pluginHandler    rpc.RolloutPlugin
	pluginClients    map[string]*goPlugin.Client
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

func NewRolloutPluginController(client k8sclient.Client, scheme *runtime.Scheme, recorder record.EventRecorder, retryWaitSeconds int, maxConcurrent int) *RolloutPluginController {
	return &RolloutPluginController{
		Client:           client,
		Scheme:           scheme,
		RetryWaitSeconds: retryWaitSeconds,
		Recorder:         recorder,
		MaxConcurrent:    maxConcurrent,
		pluginHandler:    nil,
		pluginClients:    make(map[string]*goPlugin.Client),
	}
}

var pluginMap = map[string]goPlugin.Plugin{
	// "RpcRolloutPlugin": &rpc.RpcRolloutPlugin{},
	"RpcRolloutPlugin": &rpc.RpcRolloutPlugin{},
}

func (r *RolloutPluginController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	fmt.Println("Reconcile called")
	var rolloutPlugin v1alpha1.RolloutPlugin

	err := r.Client.Get(ctx, req.NamespacedName, &rolloutPlugin)
	if err != nil {
		return ctrl.Result{}, k8sclient.IgnoreNotFound(err)
	}

	if !rolloutPlugin.Status.Initialized {
		plugin, err := pluginclient.GetRolloutPlugin(rolloutPlugin.Spec.Plugin.Name)
		if err != nil {
			fmt.Println(fmt.Errorf("unable to get plugin (%s): %w", rolloutPlugin.Spec.Plugin.Name, err))
		}
		r.pluginHandler = plugin
		plugin.InitPlugin()
		rolloutPlugin.Status.Initialized = true
		conditions := []v1alpha1.Condition{
			{
				Type:               "Initialized",
				Status:             "True",
				LastUpdateTime:     metav1.Now(),
				LastTransitionTime: metav1.Now(),
			},
		}
		rolloutPlugin.Status.Conditions = conditions
		err = r.Client.Status().Update(ctx, &rolloutPlugin)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

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

	// compute steps
	if rolloutPlugin.Spec.Strategy.Type == "Canary" {
		for _, step := range rolloutPlugin.Spec.Strategy.Canary.Steps {
			// set weight
			fmt.Println("Setting weight", step)
			r.pluginHandler.SetWeight(&rolloutPlugin)
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
