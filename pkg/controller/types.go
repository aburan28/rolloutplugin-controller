package controller

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
)

type ControllerConfig struct {
}

type Controller interface {
	Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error)
	SetupWithManager(mgr ctrl.Manager) error
}
