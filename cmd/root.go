package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aburan28/rolloutplugin-controller/api/v1alpha1"
	"github.com/aburan28/rolloutplugin-controller/pkg/controller"

	mgrs "github.com/aburan28/rolloutplugin-controller/pkg/manager"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func newCommand() *cobra.Command {
	var command = cobra.Command{
		Use:   "rolloutplugin-controller",
		Short: "rolloutplugin-controller",
		RunE: func(c *cobra.Command, args []string) error {

			// Retrieve configuration values via Viper
			logLevel := viper.GetString("log-level")
			istioEnabled := viper.GetBool("enable-istio")
			metricsAddr := viper.GetString("metrics-addr")
			probeBindAddr := viper.GetString("probe-addr")
			webhookAddr := viper.GetString("webhook-addr")
			leaderElectionNamespace := viper.GetString("leader-election-namespace")
			managerConfig := controller.Config{
				WebhookAddr:             webhookAddr,
				MetricsAddr:             metricsAddr,
				ProbeAddr:               probeBindAddr,
				LogLevel:                logLevel,
				LeaderElectionNamespace: leaderElectionNamespace,
				IstioEnabled:            istioEnabled,
			}
			fmt.Println(managerConfig)
			ctx := context.Background()
			level, err := zapcore.ParseLevel(logLevel)
			if err != nil {
				return fmt.Errorf("invalid log level %q: %v", logLevel, err)
			}

			ctrl.SetLogger(zap.New(func(o *zap.Options) {
				o.Level = level
			}))
			scheme := runtime.NewScheme()
			if err := v1alpha1.AddToScheme(scheme); err != nil {
				log.Fatal(err)
				os.Exit(1)
			}

			var clusterConfig *rest.Config
			kubeConfig := viper.GetString("kubeconfig")
			if kubeConfig != "" {
				clusterConfig, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
			} else {
				clusterConfig, err = rest.InClusterConfig()
			}
			if err != nil {
				log.Fatalln(err)
			}

			kubeClient, err := kubernetes.NewForConfig(clusterConfig)
			// ctx := context.Background()
			if err != nil {
				log.Fatalln(err)
			}
			dynamicClient, err := dynamic.NewForConfig(clusterConfig)
			if err != nil {
				log.Fatalln(err)
			}

			resource := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
			factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(dynamicClient, time.Minute, corev1.NamespaceAll, nil)
			informer := factory.ForResource(resource).Informer()
			stopCh := make(chan struct{})

			go informer.Run(stopCh)
			cm := mgrs.NewManager(kubeClient, dynamicClient, 8082, 8081)

			fmt.Println("Starting manager")
			cm.Start(ctx, 3, mgrs.NewLeaderElectionOptions())
			// cm.StartControllers()

			// mgr, err := ctrl.NewManager(clusterConfig, manager.Options{
			// 	Scheme:                  scheme,
			// 	LeaderElection:          true,
			// 	LeaderElectionID:        "rolloutplugin-controller",
			// 	LeaderElectionNamespace: "default",

			// 	Metrics: metricsserver.Options{
			// 		BindAddress: metricsAddr,
			// 	},
			// 	HealthProbeBindAddress: probeBindAddr,
			// })

			// if err != nil {
			// 	log.Fatal(err)
			// }
			// cntrler := controller.NewRolloutPluginController(mgr.GetClient(), mgr.GetScheme(), mgr.GetEventRecorderFor("rolloutplugin-controller"), 30, 4)
			// if istioEnabled {
			// 	err = controller.SetupIstioInformers(mgr, nil)
			// 	if err != nil {
			// 		log.Fatal(err)
			// 	}

			// }

			// if err != nil {
			// 	log.Fatal(err)
			// }
			// // Add a shutdown hook to kill plugins on manager stop
			// mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
			// 	<-ctx.Done() // Wait for stop signal
			// 	cntrler.Shutdown()
			// 	return nil
			// }))

			// if err = cntrler.SetupWithManager(mgr); err != nil {
			// 	log.Fatal(err)
			// }

			// err = mgr.Start(ctrl.SetupSignalHandler())
			// if err != nil {
			// 	log.Fatal(err)
			// }

			return nil
		},
	}

	// Define command-line flags and bind them to Viper keys.
	command.Flags().String("log-level", "info", "Log level")
	viper.BindPFlag("log-level", command.Flags().Lookup("log-level"))

	command.Flags().Bool("enable-istio", false, "Whether to enable istio informers")
	viper.BindPFlag("enable-istio", command.Flags().Lookup("enable-istio"))

	command.Flags().String("metrics-addr", ":8080", "The address the metric endpoint binds to.")
	viper.BindPFlag("metrics-addr", command.Flags().Lookup("metrics-addr"))

	command.Flags().String("probe-addr", ":8081", "The address the probe endpoint binds to.")
	viper.BindPFlag("probe-addr", command.Flags().Lookup("probe-addr"))

	command.Flags().String("webhook-addr", ":7000", "The address the webhook endpoint binds to.")
	viper.BindPFlag("webhook-addr", command.Flags().Lookup("webhook-addr"))

	command.Flags().String("leader-election-namespace", "default", "The namespace in which the leader election configmap will be created.")
	viper.BindPFlag("leader-election-namespace", command.Flags().Lookup("leader-election-namespace"))

	command.Flags().String("kubeconfig", "", "Path to a kubeconfig file")
	viper.BindPFlag("kubeconfig", command.Flags().Lookup("kubeconfig"))
	// Optional: enable reading from environment variables
	viper.AutomaticEnv()

	return &command
}

func Execute() {
	command := newCommand()
	if err := command.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
