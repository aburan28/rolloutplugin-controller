package cmd

import (
	"fmt"
	"log"
	"os"
	"rolloutplugin-controller/api/v1alpha1"
	"rolloutplugin-controller/pkg/controller"

	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func newCommand() *cobra.Command {
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
			mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), manager.Options{
				Scheme:                  scheme,
				LeaderElection:          true,
				LeaderElectionID:        "rolloutplugin-controller",
				LeaderElectionNamespace: "default",
			})

			if err != nil {
				log.Fatal(err)
			}
			if err = (&controller.RolloutPluginController{
				Client:           mgr.GetClient(),
				Scheme:           mgr.GetScheme(),
				Recorder:         mgr.GetEventRecorderFor("rolloutplugin-controller"),
				RetryWaitSeconds: 30,
				MaxConcurrent:    4,
			}).SetupWithManager(mgr); err != nil {
				log.Fatal(err)
				os.Exit(1)
			}
			err = mgr.Start(ctrl.SetupSignalHandler())
			if err != nil {
				log.Fatal(err)
			}

			return nil
		},
	}
	return &command
}

func Execute() {
	command := newCommand()
	if err := command.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
