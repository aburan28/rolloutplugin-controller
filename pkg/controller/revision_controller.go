package controller

import (
	"context"
	"time"

	"github.com/aburan28/rolloutplugin-controller/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type RevisionController struct {
	k8sclient.Client
	Scheme                     *runtime.Scheme
	Recorder                   record.EventRecorder
	RetryWaitSeconds           int
	MaxConcurrent              int
	kubeDynamicInformerFactory dynamicinformer.DynamicSharedInformerFactory
}

func (r *RevisionController) SetupWithManager(mgr ctrl.Manager) error {
	builder := ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{MaxConcurrentReconciles: r.MaxConcurrent}).
		For(&v1alpha1.Revision{}).
		WatchesRawSource(source.Kind(
			mgr.GetCache(),
			&appsv1.ControllerRevision{},
			handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, controllerRevision *appsv1.ControllerRevision) []ctrl.Request {
				log := ctrl.LoggerFrom(ctx)
				log.Info("Reconciling ControllerRevision", "controllerRevision", controllerRevision.Name)
				return []ctrl.Request{{
					NamespacedName: client.ObjectKey{
						Name:      controllerRevision.Name,
						Namespace: controllerRevision.Namespace,
					},
				}}
			})),
		)
		// WatchesRawSource(source.Kind(
		// 	mgr.GetCache(),
		// 	&appsv1.StatefulSet{},
		// 	handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, statefulSet *appsv1.StatefulSet) []ctrl.Request {
		// 		log := ctrl.LoggerFrom(ctx)
		// 		log.Info("Reconciling StatefulSet", "statefulset", statefulSet.Name)
		// 		return []reconcile.Request{{
		// 			NamespacedName: client.ObjectKey{
		// 				Name:      statefulSet.Name,
		// 				Namespace: statefulSet.Namespace,
		// 			},
		// 		}}
		// 	})),
		// ).
		// WatchesRawSource(source.Kind(
		// 	mgr.GetCache(),
		// 	&appsv1.ReplicaSet{},
		// 	handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, replicaSet *appsv1.ReplicaSet) []ctrl.Request {
		// 		log := ctrl.LoggerFrom(ctx)
		// 		log.Info("Reconciling ReplicaSet", "replicaset", replicaSet.Name)
		// 		return []reconcile.Request{{
		// 			NamespacedName: client.ObjectKey{
		// 				Name:      replicaSet.Name,
		// 				Namespace: replicaSet.Namespace,
		// 			},
		// 		}}
		// 	})),
		// ).
		// WatchesRawSource(source.Kind(
		// 	mgr.GetCache(),
		// 	&appsv1.DaemonSet{},
		// 	handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, daemonSet *appsv1.DaemonSet) []ctrl.Request {
		// 		log := ctrl.LoggerFrom(ctx)
		// 		log.Info("Reconciling DaemonSet", "daemonset", daemonSet.Name)
		// 		return []reconcile.Request{{
		// 			NamespacedName: client.ObjectKey{
		// 				Name:      daemonSet.Name,
		// 				Namespace: daemonSet.Namespace,
		// 			},
		// 		}}
		// 	})),
		// ).
		// WatchesRawSource(source.Kind(
		// 	mgr.GetCache(),
		// 	&appsv1.Deployment{},
		// 	handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, deployment *appsv1.Deployment) []ctrl.Request {
		// 		log := ctrl.LoggerFrom(ctx)
		// 		log.Info("Reconciling Deployment", "deployment", deployment.Name)
		// 		return []reconcile.Request{{
		// 			NamespacedName: client.ObjectKey{
		// 				Name:      deployment.Name,
		// 				Namespace: deployment.Namespace,
		// 			},
		// 		}}
		// 	})),
		// )

	return builder.Complete(r)
}

func NewRevisionControler(client k8sclient.Client, scheme *runtime.Scheme, recorder record.EventRecorder, retryWaitSeconds int, maxConcurrent int) *RevisionController {
	return &RevisionController{
		Client:           client,
		Scheme:           scheme,
		RetryWaitSeconds: retryWaitSeconds,
		Recorder:         recorder,
		MaxConcurrent:    maxConcurrent,
	}
}

func (r *RevisionController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Reconciling Revision", "revision", req.NamespacedName)
	// Fetch the Revision instance

	var revision v1alpha1.Revision
	if err := r.Get(ctx, req.NamespacedName, &revision); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// check if there is a RolloutPlugin that matches the revision
	return ctrl.Result{
		RequeueAfter: 5 * time.Minute,
	}, nil
}

// func (r *RevisionController) matchingRolloutPlugins(revision *v1alpha1.Revision) (v1alpha1.RolloutPluginList, err) {

// 	r.Client.List(context.TODO(), &v1alpha1.RolloutPluginList{}, client.MatchingLabels{
// 		"revision": revision.Name,
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	return nil, nil
// }
