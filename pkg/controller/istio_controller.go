package controller

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"istio.io/client-go/pkg/apis/networking/v1beta1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type IstioController struct {
	k8sclient.Client
	Scheme                     *runtime.Scheme
	Recorder                   record.EventRecorder
	RetryWaitSeconds           int
	MaxConcurrent              int
	kubeDynamicInformerFactory dynamicinformer.DynamicSharedInformerFactory
}

func (r *IstioController) SetupWithManager(mgr ctrl.Manager) error {
	builder := ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{MaxConcurrentReconciles: r.MaxConcurrent}).
		WatchesRawSource(source.Kind(
			mgr.GetCache(),
			&v1beta1.DestinationRule{},
			handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, destinationRule *v1beta1.DestinationRule) []ctrl.Request {
				log := ctrl.LoggerFrom(ctx)
				log.Info("Reconciling DestinationRule", "destinationRule", destinationRule.Name)
				return []reconcile.Request{{
					NamespacedName: client.ObjectKey{
						Name:      destinationRule.Name,
						Namespace: destinationRule.Namespace,
					},
				}}
			})),
		).
		For(&v1beta1.DestinationRule{})

	return builder.Complete(r)
}

func NewIstioController(client k8sclient.Client, scheme *runtime.Scheme, recorder record.EventRecorder, retryWaitSeconds int, maxConcurrent int) *IstioController {
	return &IstioController{
		Client:           client,
		Scheme:           scheme,
		RetryWaitSeconds: retryWaitSeconds,
		Recorder:         recorder,
		MaxConcurrent:    maxConcurrent,
	}
}

func (r *IstioController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Reconciling ", req.NamespacedName)
	// Fetch the Istio object

	destinationRule := &v1beta1.DestinationRule{}
	if err := r.Get(ctx, req.NamespacedName, destinationRule); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.Info("Reconciling DestinationRule", "destinationRule", destinationRule.Name)
	return ctrl.Result{
		RequeueAfter: 5 * time.Minute,
	}, nil
}
