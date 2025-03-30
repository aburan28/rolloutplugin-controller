package controller

// import (
// 	"context"

// 	"github.com/aburan28/rolloutplugin-controller/api/v1alpha1"
// 	appsv1 "k8s.io/api/apps/v1"
// 	corev1 "k8s.io/api/core/v1"
// 	"k8s.io/apimachinery/pkg/runtime"
// 	"k8s.io/client-go/dynamic/dynamicinformer"
// 	"k8s.io/client-go/tools/record"
// 	ctrl "sigs.k8s.io/controller-runtime"
// 	"sigs.k8s.io/controller-runtime/pkg/client"
// 	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
// 	"sigs.k8s.io/controller-runtime/pkg/controller"
// 	controllerutils "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
// )

// type WorkloadController struct {
// 	k8sclient.Client
// 	Scheme                     *runtime.Scheme
// 	Recorder                   record.EventRecorder
// 	RetryWaitSeconds           int
// 	MaxConcurrent              int
// 	kubeDynamicInformerFactory dynamicinformer.DynamicSharedInformerFactory
// }

// func (r *WorkloadController) SetupWithManager(mgr ctrl.Manager) error {
// 	builder := ctrl.NewControllerManagedBy(mgr).
// 		WithOptions(controller.Options{MaxConcurrentReconciles: r.MaxConcurrent}).
// 		For(&v1alpha1.Revision{})

// 	ownedResources := []client.Object{
// 		&corev1.Pod{},
// 		&corev1.PodTemplate{},
// 		&appsv1.ControllerRevision{},
// 		&appsv1.Deployment{},
// 		&appsv1.DaemonSet{},
// 		&appsv1.StatefulSet{},
// 		&appsv1.ReplicaSet{},
// 	}

// 	for _, resource := range ownedResources {
// 		builder.Owns(resource)
// 	}

// 	return builder.Complete(r)
// }

// func NewRevisionControler(client k8sclient.Client, scheme *runtime.Scheme, recorder record.EventRecorder, retryWaitSeconds int, maxConcurrent int) *WorkloadController {
// 	return &WorkloadController{
// 		Client:           client,
// 		Scheme:           scheme,
// 		RetryWaitSeconds: retryWaitSeconds,
// 		Recorder:         recorder,
// 		MaxConcurrent:    maxConcurrent,
// 	}
// }

// func (r *WorkloadController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
// 	// Fetch the Revision instance
// 	revision := &v1alpha1.Revision{}
// 	if err := r.Get(ctx, req.NamespacedName, revision); err != nil {
// 		return ctrl.Result{}, client.IgnoreNotFound(err)
// 	}

// 	// check if there is a RolloutPlugin that matches the revision

// 	if revision.GetDeletionTimestamp() != nil {
// 		if controllerutils.ContainsFinalizer(revision, FinalizerName) {
// 			controllerutils.RemoveFinalizer(revision, FinalizerName)
// 			if err := r.Update(ctx, revision); err != nil {
// 				return ctrl.Result{}, err
// 			}
// 		}
// 	}

// 	return ctrl.Result{}, nil
// }

// // func (r *WorkloadController) matchingRolloutPlugins(revision *v1alpha1.Revision) (v1alpha1.RolloutPluginList, err) {

// // 	r.Client.List(context.TODO(), &v1alpha1.RolloutPluginList{}, client.MatchingLabels{
// // 		"revision": revision.Name,
// // 	})
// // 	if err != nil {
// // 		return nil, err
// // 	}

// // 	return nil, nil
// // }
