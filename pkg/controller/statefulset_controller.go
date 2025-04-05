package controller

// package controller

// import (
// 	"context"
// 	"time"

// 	corev1 "k8s.io/api/core/v1"
// 	"k8s.io/apimachinery/pkg/runtime"
// 	"k8s.io/client-go/dynamic/dynamicinformer"
// 	"k8s.io/client-go/tools/record"
// 	ctrl "sigs.k8s.io/controller-runtime"
// 	"sigs.k8s.io/controller-runtime/pkg/client"
// 	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
// 	"sigs.k8s.io/controller-runtime/pkg/controller"
// 	"sigs.k8s.io/controller-runtime/pkg/handler"
// 	"sigs.k8s.io/controller-runtime/pkg/reconcile"
// 	"sigs.k8s.io/controller-runtime/pkg/source"
// )

// type PodTemplateController struct {
// 	k8sclient.Client
// 	Scheme                     *runtime.Scheme
// 	Recorder                   record.EventRecorder
// 	RetryWaitSeconds           int
// 	MaxConcurrent              int
// 	kubeDynamicInformerFactory dynamicinformer.DynamicSharedInformerFactory
// }

// func (r *PodTemplateController) SetupWithManager(mgr ctrl.Manager) error {
// 	builder := ctrl.NewControllerManagedBy(mgr).
// 		WithOptions(controller.Options{MaxConcurrentReconciles: r.MaxConcurrent}).
// 		WatchesRawSource(source.Kind(
// 			mgr.GetCache(),
// 			&corev1.PodTemplate{},
// 			handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, podTemplate *corev1.PodTemplate) []ctrl.Request {
// 				log := ctrl.LoggerFrom(ctx)
// 				log.Info("Reconciling PodTemplate", "podTemplate", podTemplate.Name)
// 				return []reconcile.Request{{
// 					NamespacedName: client.ObjectKey{
// 						Name:      podTemplate.Name,
// 						Namespace: podTemplate.Namespace,
// 					},
// 				}}
// 			})),
// 		).
// 		For(&corev1.PodTemplate{})

// 	return builder.Complete(r)
// }

// func NewPodTemplateControler(client k8sclient.Client, scheme *runtime.Scheme, recorder record.EventRecorder, retryWaitSeconds int, maxConcurrent int) *PodTemplateController {
// 	return &PodTemplateController{
// 		Client:           client,
// 		Scheme:           scheme,
// 		RetryWaitSeconds: retryWaitSeconds,
// 		Recorder:         recorder,
// 		MaxConcurrent:    maxConcurrent,
// 	}
// }

// func (r *PodTemplateController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
// 	log := ctrl.LoggerFrom(ctx)
// 	log.Info("Reconciling PodTemplate", "podtemplate", req.NamespacedName)
// 	// Fetch the Revision instance

// 	var podTemplate corev1.PodTemplate
// 	if err := r.Get(ctx, req.NamespacedName, &podTemplate); err != nil {
// 		return ctrl.Result{}, client.IgnoreNotFound(err)
// 	}

// 	// check if there is a RolloutPlugin that matches the revision
// 	return ctrl.Result{
// 		RequeueAfter: 5 * time.Minute,
// 	}, nil
// }
