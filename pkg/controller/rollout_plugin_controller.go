package controller

import (
	"context"
	"fmt"
	"time"

	goPlugin "github.com/hashicorp/go-plugin"

	"github.com/aburan28/rolloutplugin-controller/api/v1alpha1"
	"github.com/aburan28/rolloutplugin-controller/pkg/plugin/pluginclient"
	"github.com/aburan28/rolloutplugin-controller/pkg/plugin/rpc"
	utils "github.com/aburan28/rolloutplugin-controller/pkg/utils/plugin"
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

	pluginClients map[string]*goPlugin.Client
	plugins       map[string]rpc.RolloutPlugin
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
		// pluginClients:    make(map[string]*goPlugin.Client),

	}
}

var pluginMap = map[string]goPlugin.Plugin{
	"RpcRolloutPlugin": &rpc.RpcRolloutPlugin{},
}

func (r *RolloutPluginController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Reconciling RolloutPlugin")

	// Retrieve the RolloutPlugin CR
	var rolloutPlugin v1alpha1.RolloutPlugin
	if err := r.Client.Get(ctx, req.NamespacedName, &rolloutPlugin); err != nil {
		return ctrl.Result{}, k8sclient.IgnoreNotFound(err)
	}

	// Initialize the plugin if it hasn't been initialized yet
	// if !rolloutPlugin.Status.Initialized {
	// Check if plugin binary exists locally
	if err := utils.CheckIfExists(rolloutPlugin.Spec.Plugin.Name); err != nil {
		log.Info("Plugin not found locally, attempting to download", "plugin", rolloutPlugin.Spec.Plugin.Name)
		if err := pluginclient.DownloadPlugin(ctx, rolloutPlugin.Spec.Plugin.Url, rolloutPlugin.Spec.Plugin.Name); err != nil {
			log.Error(err, "Failed to download plugin", "plugin", rolloutPlugin.Spec.Plugin.Name)
			return ctrl.Result{}, err
		}
	}

	// Optionally verify the plugin binary if verification is enabled
	if rolloutPlugin.Spec.Plugin.Verify {
		if err := pluginclient.VerifyPlugin(rolloutPlugin.Spec.Plugin.Name, rolloutPlugin.Spec.Plugin.Sha256); err != nil {
			log.Error(err, "Plugin verification failed", "plugin", rolloutPlugin.Spec.Plugin.Name)
			return ctrl.Result{}, err
		}
	} else {
		log.Info("Skipping plugin verification", "plugin", rolloutPlugin.Spec.Plugin.Name)
	}

	// Initialize plugin if not already present in the controller maps
	if r.plugins == nil || r.pluginClients == nil {
		r.plugins = make(map[string]rpc.RolloutPlugin)
		r.pluginClients = make(map[string]*goPlugin.Client)

		// Start the plugin and store its instance and client
		t := pluginclient.NewRolloutPlugin()
		pInstance, err := t.StartPlugin(rolloutPlugin.Spec.Plugin.Name)
		r.pluginClients[rolloutPlugin.Spec.Plugin.Name] = t.Client[rolloutPlugin.Spec.Plugin.Name]
		r.plugins[rolloutPlugin.Spec.Plugin.Name] = pInstance
		if err != nil {
			log.Error(err, "Unable to start plugin", "plugin", rolloutPlugin.Spec.Plugin.Name)
			return ctrl.Result{}, err
		}
	}

	// Mark as initialized and update conditions
	rolloutPlugin.Status.Initialized = true
	rolloutPlugin.Status.Conditions = []v1alpha1.Condition{{
		Type:               "Initialized",
		Status:             "True",
		LastUpdateTime:     metav1.Now(),
		LastTransitionTime: metav1.Now(),
	}}

	// Update status after initialization
	rolloutPlugin.Status.ObservedGeneration = rolloutPlugin.GetGeneration()
	if err := r.Client.Status().Update(ctx, &rolloutPlugin); err != nil {
		log.Error(err, "Failed to update rolloutPlugin status after initialization")
		return ctrl.Result{}, err
	}
	// }

	// Set the observed generation
	rolloutPlugin.Status.ObservedGeneration = rolloutPlugin.GetGeneration()
	log.Info("Setting observed generation", "generation", rolloutPlugin.GetGeneration())

	// Execute the rollout steps if the strategy is Canary
	if rolloutPlugin.Spec.Strategy.Type == "Canary" {

		// Iterate through each step and update weight accordingly
		for i := range rolloutPlugin.Spec.Strategy.Canary.Steps {
			log.Info("Setting weight", "step", i)
			rpcErr := r.plugins[rolloutPlugin.Spec.Plugin.Name].SetWeight(&rolloutPlugin)
			// rpcErr := pluginInstance.SetWeight(&rolloutPlugin)
			// pluginInstance.SetMirrorRoute(&rolloutPlugin)
			if rpcErr.HasError() {
				err := fmt.Errorf("unable to set weight for plugin %s: %v", rolloutPlugin.Spec.Plugin.Name, rpcErr)
				log.Error(err, "Failed to set weight")
				// Optionally return error or continue based on your desired behavior
				return ctrl.Result{}, err
			}
			log.Info("Weight set successfully for step", "step", i)
		}
	}

	// Update the rolloutPlugin status at the end of reconcile
	if err := r.Client.Status().Update(ctx, &rolloutPlugin); err != nil {
		log.Error(err, "Failed to update rolloutPlugin status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{
		RequeueAfter: 5 * time.Second,
	}, nil
}

// func (r *RolloutPluginController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
// 	log := ctrl.LoggerFrom(ctx)
// 	log.Info("Reconciling RolloutPlugin")
// 	var rolloutPlugin v1alpha1.RolloutPlugin

// 	err := r.Client.Get(ctx, req.NamespacedName, &rolloutPlugin)
// 	if err != nil {
// 		return ctrl.Result{}, k8sclient.IgnoreNotFound(err)
// 	}

// 	// check observed generation
// 	// if rolloutPlugin.Status.ObservedGeneration == rolloutPlugin.GetGeneration() {
// 	// 	return ctrl.Result{}, nil
// 	// }
// 	var debug bool
// 	debug = true
// 	if !rolloutPlugin.Status.Initialized || debug {
// 		err = utils.CheckIfExists(rolloutPlugin.Spec.Plugin.Name)
// 		downloaded := false
// 		if err == nil {
// 			downloaded = true
// 		} else {
// 			fmt.Println(fmt.Errorf("unable to find plugin (%s): %w", rolloutPlugin.Spec.Plugin.Name, err))
// 		}

// 		if !downloaded {
// 			err := pluginclient.DownloadPlugin(ctx, rolloutPlugin.Spec.Plugin.Url, rolloutPlugin.Spec.Plugin.Name)
// 			if err != nil {
// 				fmt.Println(fmt.Errorf("unable to download plugin (%s): %w", rolloutPlugin.Spec.Plugin.Name, err))
// 			}
// 		}

// 		if rolloutPlugin.Spec.Plugin.Verify {
// 			err = pluginclient.VerifyPlugin(rolloutPlugin.Spec.Plugin.Name, rolloutPlugin.Spec.Plugin.Sha256)
// 			if err != nil {
// 				fmt.Println(fmt.Errorf("unable to verify plugin (%s): %w", rolloutPlugin.Spec.Plugin.Name, err))
// 			}
// 		} else {
// 			log.Info("Skipping plugin verification")
// 		}

// 		// plugin, err := pluginclient.GetRolloutPlugin(rolloutPlugin.Spec.Plugin.Name)
// 		// if err != nil {
// 		// 	fmt.Println(fmt.Errorf("unable to get plugin (%s): %w", rolloutPlugin.Spec.Plugin.Name, err))
// 		// }
// 		var pluginInstance rpc.RolloutPlugin
// 		if r.plugins == nil {
// 			r.plugins = make(map[string]rpc.RolloutPlugin)
// 			r.pluginClients = make(map[string]*goPlugin.Client)
// 			t := pluginclient.NewRolloutPlugin()
// 			pluginInstance, err = t.StartPlugin(rolloutPlugin.Spec.Plugin.Name)
// 			if err != nil {
// 				fmt.Println(fmt.Errorf("unable to get plugin (%s): %w", rolloutPlugin.Spec.Plugin.Name, err))
// 			}
// 			r.plugins[rolloutPlugin.Spec.Plugin.Name] = pluginInstance
// 			r.plugins[rolloutPlugin.Spec.Plugin.Name] = t.Plugin[rolloutPlugin.Spec.Plugin.Name]
// 			r.pluginClients[rolloutPlugin.Spec.Plugin.Name] = t.Client[rolloutPlugin.Spec.Plugin.Name]
// 			log.Info("Starting plugin")
// 			fmt.Println(r.plugins, r.pluginClients)
// 		}
// 		log.Info("we have the plugins")
// 		fmt.Println(r.plugins, r.pluginClients)
// 		fmt.Println("pluginPath: ", rolloutPlugin.Spec.Plugin.Name)
// 		// log.Info("Starting plugin")
// 		log.Info("Initializing plugin")
// 		// rpcError := plugin.InitPlugin()

// 		// if rpcError.HasError() {
// 		// 	fmt.Println(fmt.Errorf("unable to initialize plugin (%s): %w", rolloutPlugin.Spec.Plugin.Name, rpcError))
// 		// }

// 		// rolloutPlugin.Status.Initialized = true
// 		// conditions := []v1alpha1.Condition{
// 		// 	{
// 		// 		Type:               "Initialized",
// 		// 		Status:             "True",
// 		// 		LastUpdateTime:     metav1.Now(),
// 		// 		LastTransitionTime: metav1.Now(),
// 		// 	},
// 		// }
// 		// rolloutPlugin.Status.Conditions = conditions
// 		// rolloutPlugin.Status.Initialized = true
// 		// generation := rolloutPlugin.GetGeneration()
// 		// rolloutPlugin.Status.ObservedGeneration = generation
// 		// err = r.Client.Status().Update(ctx, &rolloutPlugin)
// 		// r.pluginClients[rolloutPlugin.Spec.Plugin.Name].Client()
// 	}

// 	log.Info("Checking if we need to start the steps")

// 	generation := rolloutPlugin.GetGeneration()
// 	rolloutPlugin.Status.ObservedGeneration = generation
// 	log.Info("Setting generation")
// 	if rolloutPlugin.Status.Conditions == nil {
// 		rolloutPlugin.Status.Conditions = []v1alpha1.Condition{
// 			{
// 				Type:               "Initialized",
// 				Status:             "True",
// 				LastUpdateTime:     metav1.Now(),
// 				LastTransitionTime: metav1.Now(),
// 			},
// 		}
// 	}
// 	log.Info("lets start the steps")
// 	if rolloutPlugin.Spec.Strategy.Type == "Canary" {
// 		for k, _ := range rolloutPlugin.Spec.Strategy.Canary.Steps {
// 			log.Info("Setting weight")
// 			log.Info(fmt.Sprintf("Step: %d", k))
// 			if r.plugins[rolloutPlugin.Spec.Plugin.Name] == nil {
// 				log.Info("Plugin is nil")
// 			}
// 			rpcErr := r.plugins[rolloutPlugin.Spec.Plugin.Name].SetWeight(&rolloutPlugin)
// 			// rpcError := plugin.SetWeight(rolloutPlugin.Spec.Strategy.Canary.Steps[k].Weight)
// 			log.Info("Setting weight post clients")
// 			if rpcErr.HasError() {
// 				fmt.Println(fmt.Errorf("unable to set weight (%s): %w", rolloutPlugin.Spec.Plugin.Name, rpcErr))
// 			}
// 			// defer func() {
// 			// 	go r.pluginHandler.SetWeight(&rolloutPlugin)
// 			// }()
// 		}
// 	}

// 	err = r.Client.Status().Update(ctx, &rolloutPlugin)
// 	if err != nil {
// 		return ctrl.Result{}, err
// 	}
// 	return ctrl.Result{}, nil
// }

func (r *RolloutPluginController) SetConditions(rolloutPlugin *v1alpha1.RolloutPlugin) {

}

func (r *RolloutPluginController) Run(ctx context.Context) error {

	return nil
}

func (r *RolloutPluginController) Shutdown() {
	for _, client := range r.pluginClients {

		client.Kill()
	}
}
