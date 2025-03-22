package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aburan28/rolloutplugin-controller/api/v1alpha1"
	"github.com/aburan28/rolloutplugin-controller/pkg/controller"
	mgrs "github.com/aburan28/rolloutplugin-controller/pkg/manager"

	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func newCommand() *cobra.Command {
	var (
		metricsAddr   string
		probeBindAddr string
		webhookAddr   string
	)
	var command = cobra.Command{
		Use:   "rolloutplugin-controller",
		Short: "rolloutplugin-controller",
		RunE: func(c *cobra.Command, args []string) error {
			ctrl.SetLogger(
				zap.New(func(o *zap.Options) {
					o.Level = zapcore.Level(-5)
				}),
			)
			scheme := runtime.NewScheme()
			if err := v1alpha1.AddToScheme(scheme); err != nil {
				log.Fatal(err)
				os.Exit(1)
			}

			kubeClient, _ := kubernetes.NewForConfig(ctrl.GetConfigOrDie())
			ctx := context.Background()

			cm := mgrs.NewManager(kubeClient, 8082, 8081)
			cm.Start(ctx, 3, mgrs.NewLeaderElectionOptions())

			mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), manager.Options{
				Scheme:                  scheme,
				LeaderElection:          true,
				LeaderElectionID:        "rolloutplugin-controller",
				LeaderElectionNamespace: "default",
				Metrics: metricsserver.Options{
					BindAddress: metricsAddr,
				},
				HealthProbeBindAddress: probeBindAddr,
			})

			if err != nil {
				log.Fatal(err)
			}
			cntrler := controller.NewRolloutPluginController(mgr.GetClient(), mgr.GetScheme(), mgr.GetEventRecorderFor("rolloutplugin-controller"), 30, 4)
			if err = cntrler.SetupWithManager(mgr); err != nil {
				log.Fatal(err)
			}

			err = mgr.Start(ctrl.SetupSignalHandler())
			if err != nil {
				log.Fatal(err)
			}

			return nil
		},
	}

	command.Flags().StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	command.Flags().StringVar(&probeBindAddr, "probe-addr", ":8081", "The address the probe endpoint binds to.")
	command.Flags().StringVar(&webhookAddr, "webhook-addr", ":7000", "The address the webhook endpoint binds to.")
	return &command
}

func Execute() {
	command := newCommand()
	if err := command.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
