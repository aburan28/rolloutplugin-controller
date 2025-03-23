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
	pluginHandler    rpc.RolloutPlugin
	pluginClients    map[string]*goPlugin.Client
	plugins          map[string]rpc.RolloutPlugin
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
		// pluginClients:    make(map[string]*goPlugin.Client),

	}
}

var pluginMap = map[string]goPlugin.Plugin{
	"RpcRolloutPlugin": &rpc.RpcRolloutPlugin{},
}

func (r *RolloutPluginController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Reconciling RolloutPlugin")
	var rolloutPlugin v1alpha1.RolloutPlugin

	err := r.Client.Get(ctx, req.NamespacedName, &rolloutPlugin)
	if err != nil {
		return ctrl.Result{}, k8sclient.IgnoreNotFound(err)
	}

	// check observed generation
	// if rolloutPlugin.Status.ObservedGeneration == rolloutPlugin.GetGeneration() {
	// 	return ctrl.Result{}, nil
	// }
	var debug bool
	debug = true
	if !rolloutPlugin.Status.Initialized || debug {
		err = utils.CheckIfExists(rolloutPlugin.Spec.Plugin.Name)
		downloaded := false
		if err == nil {
			downloaded = true
		} else {
			fmt.Println(fmt.Errorf("unable to find plugin (%s): %w", rolloutPlugin.Spec.Plugin.Name, err))
		}

		if !downloaded {
			err := pluginclient.DownloadPlugin(ctx, rolloutPlugin.Spec.Plugin.Url, rolloutPlugin.Spec.Plugin.Name)
			if err != nil {
				fmt.Println(fmt.Errorf("unable to download plugin (%s): %w", rolloutPlugin.Spec.Plugin.Name, err))
			}
		}

		if rolloutPlugin.Spec.Plugin.Verify {
			err = pluginclient.VerifyPlugin(rolloutPlugin.Spec.Plugin.Name, rolloutPlugin.Spec.Plugin.Sha256)
			if err != nil {
				fmt.Println(fmt.Errorf("unable to verify plugin (%s): %w", rolloutPlugin.Spec.Plugin.Name, err))
			}
		} else {
			log.Info("Skipping plugin verification")
		}

		// plugin, err := pluginclient.GetRolloutPlugin(rolloutPlugin.Spec.Plugin.Name)
		// if err != nil {
		// 	fmt.Println(fmt.Errorf("unable to get plugin (%s): %w", rolloutPlugin.Spec.Plugin.Name, err))
		// }
		if r.plugins == nil {
			r.plugins = make(map[string]rpc.RolloutPlugin)
			r.pluginClients = make(map[string]*goPlugin.Client)
			t := pluginclient.NewRolloutPlugin()
			t.StartPlugin(rolloutPlugin.Spec.Plugin.Name)
			r.plugins[rolloutPlugin.Spec.Plugin.Name] = t.Plugin[rolloutPlugin.Spec.Plugin.Name]
			r.pluginClients[rolloutPlugin.Spec.Plugin.Name] = t.Client[rolloutPlugin.Spec.Plugin.Name]
			log.Info("Starting plugin")
		}
		log.Info("we have the plugins")
		fmt.Println(r.plugins, r.pluginClients)
		fmt.Println("pluginPath: ", rolloutPlugin.Spec.Plugin.Name)
		// log.Info("Starting plugin")
		log.Info("Initializing plugin")
		// rpcError := plugin.InitPlugin()

		// if rpcError.HasError() {
		// 	fmt.Println(fmt.Errorf("unable to initialize plugin (%s): %w", rolloutPlugin.Spec.Plugin.Name, rpcError))
		// }

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
		rolloutPlugin.Status.Initialized = true
		generation := rolloutPlugin.GetGeneration()
		rolloutPlugin.Status.ObservedGeneration = generation
		err = r.Client.Status().Update(ctx, &rolloutPlugin)
		if err != nil {
			return ctrl.Result{
				RequeueAfter: 5 * time.Second,
			}, err
		}
		return ctrl.Result{
			RequeueAfter: 5 * time.Second,
		}, nil
	}

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

	if rolloutPlugin.Spec.Strategy.Type == "Canary" {
		for k, _ := range rolloutPlugin.Spec.Strategy.Canary.Steps {
			log.Info("Setting weight")
			log.Info(fmt.Sprintf("Step: %d", k))
			err = r.pluginHandler.SetWeight(&rolloutPlugin)
			if err != nil {
				log.Error(err, "unable to set weight")
			}
			// defer func() {
			// 	go r.pluginHandler.SetWeight(&rolloutPlugin)
			// }()
		}
	}

	err = r.Client.Status().Update(ctx, &rolloutPlugin)
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *RolloutPluginController) SetConditions(rolloutPlugin *v1alpha1.RolloutPlugin) {

}

func (r *RolloutPluginController) Run(ctx context.Context) error {

	return nil
}
