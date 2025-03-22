package main

import (
	"github.com/aburan28/rolloutplugin-controller/cmd"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var log = logf.Log.WithName("rolloutplugin-controller")

func init() {
	logf.SetLogger(zap.New())
}

func main() {
	cmd.Execute()
	// cfg, err := config.GetConfig()
	// if err != nil {
	// 	log.Error(err, "unable to get kubeconfig")
	// 	os.Exit(1)
	// }
	// scheme := runtime.NewScheme()
	// if err := v1alpha1.AddToScheme(scheme); err != nil {
	// 	log.Error(err, "unable to add scheme")
	// 	os.Exit(1)
	// }
	// mgr, err := ctrl.NewManager(cfg, ctrl.Options{
	// 	Scheme: scheme,
	// 	// Metrics:                metricsServerOptions,
	// 	// WebhookServer:          webhookServer,
	// 	// HealthProbeBindAddress: probeAddr,
	// 	// LeaderElection:         enableLeaderElection,
	// 	// LeaderElectionID:       "",
	// 	// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
	// 	// when the Manager ends. This requires the binary to immediately end when the
	// 	// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
	// 	// speeds up voluntary leader transitions as the new leader don't have to wait
	// 	// LeaseDuration time first.
	// 	//
	// 	// In the default scaffold provided, the program ends immediately after
	// 	// the manager stops, so would be fine to enable this option. However,
	// 	// if you are doing or is intended to do any operation such as perform cleanups
	// 	// after the manager stops then its usage might be unsafe.
	// 	// LeaderElectionReleaseOnCancel: true,
	// })
	// if err != nil {
	// 	log.Error(err, "unable to start manager")
	// 	os.Exit(1)
	// }

	// if err = (&controller.RolloutPluginReconciler{
	// 	Client:           mgr.GetClient(),
	// 	Scheme:           mgr.GetScheme(),
	// 	Recorder:         mgr.GetEventRecorderFor("rolloutplugin-controller"),
	// 	RetryWaitSeconds: 30,
	// }).SetupWithManager(mgr); err != nil {
	// 	log.Error(err, "unable to create controller", "controller", "PhasedRollout")
	// 	os.Exit(1)
	// }

	// if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
	// 	log.Error(err, "problem running manager")
	// 	os.Exit(1)
	// }
}
