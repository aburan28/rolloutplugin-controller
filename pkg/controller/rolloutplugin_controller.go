package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/aburan28/rolloutplugin-controller/api/v1alpha1"
	"github.com/aburan28/rolloutplugin-controller/pkg/plugin/pluginclient"
	"github.com/aburan28/rolloutplugin-controller/pkg/plugin/rpc"
	utils "github.com/aburan28/rolloutplugin-controller/pkg/utils/plugin"
	goPlugin "github.com/hashicorp/go-plugin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	controllerutils "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var pluginMap = map[string]goPlugin.Plugin{
	"RpcRolloutPlugin": &rpc.RpcRolloutPlugin{},
}

// PhasedRolloutReconciler reconciles a PhasedRollout object.
type RolloutPluginController struct {
	k8sclient.Client
	Scheme                 *runtime.Scheme
	Recorder               record.EventRecorder
	RetryWaitSeconds       int
	MaxConcurrent          int
	rolloutPluginWorkqueue workqueue.TypedRateLimitingInterface[any]

	pluginClients map[string]*goPlugin.Client
	plugins       map[string]rpc.RolloutPlugin
}

func (r *RolloutPluginController) SetupWithManager(mgr ctrl.Manager) error {
	builder := ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{MaxConcurrentReconciles: r.MaxConcurrent}).
		For(&v1alpha1.RolloutPlugin{})

	// ownedResources := []client.Object{
	// 	&corev1.Pod{},
	// 	&corev1.PodTemplate{},
	// 	&appsv1.ControllerRevision{},
	// 	&corev1.Service{},
	// 	&appsv1.Deployment{},
	// 	&appsv1.DaemonSet{},
	// 	&appsv1.StatefulSet{},
	// 	&appsv1.ReplicaSet{},
	// 	&networkingv1alpha3.VirtualService{},
	// 	&networkingv1alpha3.DestinationRule{},
	// }

	// for _, resource := range ownedResources {
	// 	builder.Owns(resource)
	// }

	return builder.Complete(r)

}

func NewRolloutPluginController(client k8sclient.Client, scheme *runtime.Scheme, recorder record.EventRecorder, retryWaitSeconds int, maxConcurrent int) *RolloutPluginController {
	return &RolloutPluginController{
		Client:           client,
		Scheme:           scheme,
		RetryWaitSeconds: retryWaitSeconds,
		Recorder:         recorder,
		MaxConcurrent:    maxConcurrent,
	}
}

func (r *RolloutPluginController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx).WithValues("rolloutplugin", req.NamespacedName)

	// Retrieve the RolloutPlugin CR
	var rolloutPlugin v1alpha1.RolloutPlugin
	if err := r.Client.Get(ctx, req.NamespacedName, &rolloutPlugin); err != nil {
		return ctrl.Result{}, k8sclient.IgnoreNotFound(err)
	}
	var err error

	if deletionTimestamp := rolloutPlugin.GetDeletionTimestamp(); deletionTimestamp != nil {
		log.Info("Deleting plugin", "plugin", rolloutPlugin.Spec.Plugin.Name)
		if controllerutils.ContainsFinalizer(&rolloutPlugin, FinalizerName) {
			go r.pluginClients[rolloutPlugin.Spec.Plugin.Name].Kill()
			// Remove the finalizer from the CR

			controllerutils.RemoveFinalizer(&rolloutPlugin, FinalizerName)

			if err := r.Client.Update(ctx, &rolloutPlugin); err != nil {
				log.Error(err, "Failed to remove finalizer from rolloutPlugin", "plugin", rolloutPlugin.Spec.Plugin.Name)
				return ctrl.Result{}, err
			}
		}
	}

	// Initialize plugin if not already present in the controller maps
	if r.plugins == nil || r.pluginClients == nil {
		err = r.checkPluginExists(ctx, &rolloutPlugin)
		if err != nil {
			log.Error(err, "Failed to check plugin existence", "plugin", rolloutPlugin.Spec.Plugin.Name)
			return ctrl.Result{}, err
		}
		if rolloutPlugin.Spec.Plugin.Verify {
			if err := pluginclient.VerifyPlugin(rolloutPlugin.Spec.Plugin.Name, rolloutPlugin.Spec.Plugin.Sha256); err != nil {
				log.Error(err, "Plugin verification failed", "plugin", rolloutPlugin.Spec.Plugin.Name)
				return ctrl.Result{}, err
			}
		} else {
			log.Info("Skipping plugin verification", "plugin", rolloutPlugin.Spec.Plugin.Name)
		}
		err := r.initializePlugin(ctx, &rolloutPlugin)
		if err != nil {
			log.Error(err, "Failed to initialize plugin", "plugin", rolloutPlugin.Spec.Plugin.Name)
			return ctrl.Result{}, err
		}
		rpcErr := r.plugins[rolloutPlugin.Spec.Plugin.Name].Sync(&rolloutPlugin)
		if rpcErr.HasError() {
			err := fmt.Errorf("unable to sync plugin %s: %v", rolloutPlugin.Spec.Plugin.Name, rpcErr)
			log.Error(err, "Failed to sync plugin")
			return ctrl.Result{}, err
		}
	}

	controllerutils.AddFinalizer(&rolloutPlugin, FinalizerName)
	if err := r.Client.Update(ctx, &rolloutPlugin); err != nil {
		log.Error(err, "Failed to add finalizer to rolloutPlugin", "plugin", rolloutPlugin.Spec.Plugin.Name)
		return ctrl.Result{}, err
	}

	// Set the observed generation
	observedGeneration := rolloutPlugin.Status.ObservedGeneration
	if observedGeneration != rolloutPlugin.GetGeneration() {
		log.Info("Updating observed generation", "plugin", rolloutPlugin.Spec.Plugin.Name)
		rolloutPlugin.Status.ObservedGeneration = rolloutPlugin.GetGeneration()
	}

	currentRevision := rolloutPlugin.Status.CurrentRevision
	if currentRevision != rolloutPlugin.Status.UpdatedRevision || rolloutPlugin.Status.RolloutInProgress {
		log.Info("Updating current revision", "plugin", rolloutPlugin.Spec.Plugin.Name)
		err = r.processRolloutSteps(ctx, &rolloutPlugin)
		if err != nil {
			log.Error(err, "Failed to process rollout steps", "plugin", rolloutPlugin.Spec.Plugin.Name)
			return ctrl.Result{}, err
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

func (r *RolloutPluginController) checkPluginExists(ctx context.Context, rolloutPlugin *v1alpha1.RolloutPlugin) error {
	log := ctrl.LoggerFrom(ctx).WithValues("rolloutplugin", rolloutPlugin.Name)

	if err := utils.CheckIfExists(rolloutPlugin.Spec.Plugin.Name); err != nil {
		log.Info("Plugin not found locally, attempting to download", "plugin", rolloutPlugin.Spec.Plugin.Name)
		if err := pluginclient.DownloadPlugin(ctx, rolloutPlugin.Spec.Plugin.Url, rolloutPlugin.Spec.Plugin.Name); err != nil {
			log.Error(err, "Failed to download plugin", "plugin", rolloutPlugin.Spec.Plugin.Name)
			return err
		}
	}
	log.Info("Plugin found locally", "plugin", rolloutPlugin.Spec.Plugin.Name)
	// Check if the plugin is already running
	return nil
}

func (r *RolloutPluginController) initializePlugin(ctx context.Context, rolloutPlugin *v1alpha1.RolloutPlugin) error {
	log := ctrl.LoggerFrom(ctx).WithValues("rolloutplugin", rolloutPlugin.Name)
	r.plugins = make(map[string]rpc.RolloutPlugin)
	r.pluginClients = make(map[string]*goPlugin.Client)

	// Start the plugin and store its instance and client
	t := pluginclient.NewRolloutPlugin()
	pInstance, err := t.StartPlugin(rolloutPlugin.Spec.Plugin.Name)
	r.pluginClients[rolloutPlugin.Spec.Plugin.Name] = t.Client[rolloutPlugin.Spec.Plugin.Name]
	r.plugins[rolloutPlugin.Spec.Plugin.Name] = pInstance
	if err != nil {

		condition := v1alpha1.Condition{
			Type:    v1alpha1.RolloutPluginConditionTypeInitialized,
			Status:  v1alpha1.RolloutPluginConditionStatusFalse,
			Reason:  "PluginInitializationFailed",
			Message: "Plugin initialization failed",
		}
		rolloutPlugin.Status.RolloutInProgress = false
		r.SetConditions(rolloutPlugin, condition)
		log.Error(err, "Failed to initialize plugin", "plugin", rolloutPlugin.Spec.Plugin.Name)
		return err
	}
	// Mark as initialized and update conditions
	rolloutPlugin.Status.Initialized = true
	condition := v1alpha1.Condition{
		Type:    v1alpha1.RolloutPluginConditionTypeInitialized,
		Status:  v1alpha1.RolloutPluginConditionStatusUnknown,
		Reason:  "PluginInitialized",
		Message: "Plugin initialized successfully",
	}

	r.SetConditions(rolloutPlugin, condition)
	log.Info("Plugin initialized successfully", "plugin", rolloutPlugin.Spec.Plugin.Name)

	return nil
}

func (r *RolloutPluginController) processRolloutSteps(ctx context.Context, rolloutPlugin *v1alpha1.RolloutPlugin) error {

	log := ctrl.LoggerFrom(ctx).WithValues("rolloutplugin", rolloutPlugin.Name)
	if rolloutPlugin.Spec.Strategy.Type == "Canary" {
		var curStepIndex int32

		if rolloutPlugin.Status.UpdatedRevision == rolloutPlugin.Status.PreviousRevision {
			// Set the rollout in progress status
			rolloutPlugin.Status.RolloutInProgress = false

			log.Info("Rollout is not in progress")
			if err := r.Client.Status().Update(ctx, rolloutPlugin); err != nil {
				log.Error(err, "Failed to update rolloutPlugin status")
				return err
			}
			return nil
		}

		// Check if the rollout is in progress
		if rolloutPlugin.Status.CurrentStepIndex == 0 && rolloutPlugin.Status.PreviousRevision != rolloutPlugin.Status.UpdatedRevision {
			// Set the rollout in progress status
			rolloutPlugin.Status.RolloutInProgress = true
			// Initialize the current step index if not set
			rolloutPlugin.Status.CurrentStepIndex = 1
			curStepIndex = 1
			log.Info("Initializing current step index", "step", rolloutPlugin.Status.CurrentStepIndex)
		}

		if rolloutPlugin.Status.CurrentStepComplete {
			curStepIndex += 1
			rolloutPlugin.Status.CurrentStepIndex = curStepIndex
			rolloutPlugin.Status.CurrentStepComplete = false
			log.Info("Current step complete, moving to next step", "step", rolloutPlugin.Status.CurrentStepIndex)
		}

		log.Info("Executing rollout steps for plugin", "plugin", rolloutPlugin.Spec.Plugin.Name)
		curStep := rolloutPlugin.Spec.Strategy.Canary.Steps[rolloutPlugin.Status.CurrentStepIndex-1]

		if curStep.Pause != nil {
			log.Info("Pausing rollout for plugin", "plugin", rolloutPlugin.Spec.Plugin.Name)
			rolloutPlugin.Status.Paused = true
			var d time.Duration
			rolloutPlugin.Status.ResumeTime = &metav1.Time{Time: time.Now().Add(d)}
		}

		if curStep.SetWeight != nil {
			rpcErr := r.plugins[rolloutPlugin.Spec.Plugin.Name].SetWeight(rolloutPlugin)
			if rpcErr.HasError() {
				err := fmt.Errorf("unable to set weight for plugin %s: %v", rolloutPlugin.Spec.Plugin.Name, rpcErr)
				log.Error(err, "Failed to set weight")
				// Optionally return error or continue based on your desired behavior
				return err
			}
			log.Info("Setting weight for plugin", "plugin", rolloutPlugin.Spec.Plugin.Name)
			rolloutPlugin.Status.CurrentStepComplete = true
			rolloutPlugin.Status.Paused = false
			rolloutPlugin.Status.ResumeTime = nil
			rolloutPlugin.Status.CurrentStepIndex = curStepIndex
		}
	}

	return nil
}
func (r *RolloutPluginController) SetConditions(rolloutPlugin *v1alpha1.RolloutPlugin, condition v1alpha1.Condition) {
	condition.LastUpdateTime = metav1.Now()
	condition.LastTransitionTime = metav1.Now()
	rolloutPlugin.Status.Conditions = append(rolloutPlugin.Status.Conditions, condition)
	if err := r.Client.Status().Update(context.Background(), rolloutPlugin); err != nil {
		ctrl.LoggerFrom(context.Background()).Error(err, "Failed to update rolloutPlugin status")
	}
}

func (r *RolloutPluginController) Shutdown() {
	for _, client := range r.pluginClients {
		if client != nil {
			client.Kill()
		}
	}
}

func (r *RolloutPluginController) Run(ctx context.Context, threadiness int) error {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Starting Rollout workers")

	return nil
}

// func (r *RolloutPluginController) worker(ctx context.Context) {
// 	log := ctrl.LoggerFrom(ctx)
// 	log.Info("Starting rollout plugin worker")
// 	for {
// 		obj, shutdown := r.rolloutPluginWorkqueue.Get()
// 		if shutdown {
// 			return
// 		}
// 		_, err := r.Reconcile(ctx, obj.(ctrl.Request))
// 		if err != nil {
// 			log.Error(err, "Error reconciling rollout plugin")
// 			r.rolloutPluginWorkqueue.AddRateLimited(obj)
// 		}
// 		r.rolloutPluginWorkqueue.Forget(obj)
// 	}
// }
