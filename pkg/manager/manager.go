package manager

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/aburan28/rolloutplugin-controller/api/v1alpha1"
	"github.com/aburan28/rolloutplugin-controller/pkg/controller"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"

	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/kubectl/pkg/scheme"
	mgr "sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	DefaultListenAddr = ":8080"
	listenAddr        = "0.0.0.0:%d"

	DefaultHealthzPort = 8081
	DefaultMetricsPort = 8082
	// DefaultLeaderElect is the default true leader election should be enabled
	DefaultLeaderElect = true

	// DefaultLeaderElectionLeaseDuration is the default time in seconds that non-leader candidates will wait to force acquire leadership
	DefaultLeaderElectionLeaseDuration = 15 * time.Second

	// DefaultLeaderElectionRenewDeadline is the default time in seconds that the acting master will retry refreshing leadership before giving up
	DefaultLeaderElectionRenewDeadline = 10 * time.Second

	// DefaultLeaderElectionRetryPeriod is the default time in seconds that the leader election clients should wait between tries of actions
	DefaultLeaderElectionRetryPeriod = 2 * time.Second

	defaultLeaderElectionLeaseLockName = "rolloutplugin-controller-lock"
)

type Manager struct {
	wg                            *sync.WaitGroup
	kubeClientSet                 kubernetes.Interface
	metricsServer                 *controller.MetricsServer
	healthzServer                 *http.Server
	informer                      cache.SharedIndexInformer
	indexer                       cache.Indexer
	rolloutPluginWorkqueue        workqueue.TypedRateLimitingInterface[any]
	kubeInformerFactory           informers.SharedInformerFactory
	dynamicInformerFactory        dynamicinformer.DynamicSharedInformerFactory
	clusterDynamicInformerFactory dynamicinformer.DynamicSharedInformerFactory
	istioDynamicInformerFactory   dynamicinformer.DynamicSharedInformerFactory
	rolloutPluginController       controller.RolloutPluginController
	recorder                      record.EventRecorder
	Manager                       mgr.Manager

	rolloutPluginSynced   cache.InformerSynced
	deploymentSynced      cache.InformerSynced
	replicaSetSynced      cache.InformerSynced
	rolloutPluginInformer cache.SharedIndexInformer
	deploymentInformer    cache.SharedIndexInformer
	replicaSetInformer    cache.SharedIndexInformer
}

type LeaderElectionOptions struct {
	LeaderElect                 bool
	LeaderElectionNamespace     string
	LeaderElectionLeaseDuration time.Duration
	LeaderElectionRenewDeadline time.Duration
	LeaderElectionRetryPeriod   time.Duration
}

func NewLeaderElectionOptions() *LeaderElectionOptions {
	return &LeaderElectionOptions{
		LeaderElect:                 DefaultLeaderElect,
		LeaderElectionNamespace:     "default",
		LeaderElectionLeaseDuration: DefaultLeaderElectionLeaseDuration,
		LeaderElectionRenewDeadline: DefaultLeaderElectionRenewDeadline,
		LeaderElectionRetryPeriod:   DefaultLeaderElectionRetryPeriod,
	}
}

func (m *Manager) startLeading(ctx context.Context, rolloutThreadiness int) {
	defer runtime.HandleCrash()
	// Start the informer factories to begin populating the informer caches
	log.Info("Starting Controllers")
	log.Info("Waiting for controller's informer caches to sync")
	if ok := cache.WaitForCacheSync(ctx.Done(), m.rolloutPluginSynced); !ok {
		log.Fatalf("failed to wait for caches to sync, exiting")
	}
	// notice that there is no need to run Start methods in a separate goroutine. (i.e. go kubeInformerFactory.Start(stopCh)
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	m.dynamicInformerFactory.Start(ctx.Done())
	m.clusterDynamicInformerFactory.Start(ctx.Done())
	m.istioDynamicInformerFactory.Start(ctx.Done())
	m.kubeInformerFactory.Start(ctx.Done())

	go wait.Until(func() {
		m.wg.Add(1)
		m.rolloutPluginController.Run(ctx, 3)
		m.wg.Done()
	}, time.Second, ctx.Done())

}

func (m *Manager) Start(ctx context.Context, rolloutPluginThreadiness int, electOpts *LeaderElectionOptions) error {

	defer runtime.HandleCrash()
	defer m.rolloutPluginWorkqueue.ShutDown()
	go func() {
		log.Infof("Starting Healthz Server at %s", m.healthzServer.Addr)
		err := m.healthzServer.ListenAndServe()
		if err != nil {
			err = errors.Wrap(err, "Healthz Server Error")
			log.Error(err)
		}
	}()

	mux := controller.NewPProfServer()
	go func() {
		log.Println(http.ListenAndServe("127.0.0.1:8888", mux))
	}()

	go func() {
		log.Infof("Starting Metric Server at %s", m.metricsServer.Addr)
		if err := m.metricsServer.ListenAndServe(); err != nil {
			log.Error(errors.Wrap(err, "Metric Server Error"))
		}
	}()
	go m.dynamicInformerFactory.Start(ctx.Done())
	go m.clusterDynamicInformerFactory.Start(ctx.Done())
	go m.istioDynamicInformerFactory.Start(ctx.Done())

	// m.kubeInformerFactory.Start(ctx.Done())
	go func() {
		m.rolloutPluginController.Run(ctx, 3)
	}()
	// go wait.Until(func() {
	// 	m.wg.Add(1)
	// 	m.rolloutPluginController.Run(ctx, 3)
	// 	m.wg.Done()
	// }, time.Second, ctx.Done())
	<-ctx.Done()

	// Once context is canceled, gracefully shut down servers
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := m.healthzServer.Shutdown(ctxWithTimeout); err != nil {
		log.Errorf("Healthz Server Shutdown Error: %v", err)
	}
	if err := m.metricsServer.Shutdown(ctxWithTimeout); err != nil {
		log.Errorf("Metrics Server Shutdown Error: %v", err)
	}
	log.Info("Waiting for controller's informer caches to sync")
	if ok := cache.WaitForCacheSync(ctx.Done(), m.deploymentSynced, m.replicaSetSynced, m.rolloutPluginSynced); !ok {
		log.Fatalf("failed to wait for caches to sync, exiting")
	}
	m.wg.Wait()

	go wait.Until(func() { m.wg.Add(1); m.rolloutPluginController.Run(ctx, 3); m.wg.Done() }, time.Second, ctx.Done())

	return nil
}

func NewManager(kubeclientset kubernetes.Interface, dynamicClient dynamic.Interface, metricsPort int,
	healthzPort int, rolloutPluginController controller.RolloutPluginController, mgr mgr.Manager) *Manager {

	manager := &Manager{}
	manager.Manager = mgr
	runtime.Must(v1alpha1.AddToScheme(scheme.Scheme))
	recorder := record.NewBroadcaster().NewRecorder(scheme.Scheme, corev1.EventSource{Component: "rollout-plugin"})

	healthzServer := controller.NewHealthzServer(fmt.Sprintf(listenAddr, healthzPort))
	serverConfig := controller.ServerConfig{
		Addr: fmt.Sprintf(listenAddr, metricsPort),
	}
	metricsServer := controller.NewMetricsServer(serverConfig)
	rolloutPluginWorkqueue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "rolloutplugin")
	// Initialize the informer factories here
	dynamicInformerFactory := dynamicinformer.NewDynamicSharedInformerFactory(dynamicClient, time.Minute)

	clusterDynamicInformerFactory := dynamicinformer.NewDynamicSharedInformerFactory(dynamicClient, time.Minute)
	istioDynamicInformerFactory := dynamicinformer.NewDynamicSharedInformerFactory(dynamicClient, time.Minute)
	manager.rolloutPluginController = rolloutPluginController
	manager.wg = &sync.WaitGroup{}
	manager.kubeClientSet = kubeclientset
	manager.metricsServer = metricsServer
	manager.healthzServer = healthzServer
	manager.recorder = recorder
	manager.rolloutPluginWorkqueue = rolloutPluginWorkqueue
	manager.dynamicInformerFactory = dynamicInformerFactory
	manager.clusterDynamicInformerFactory = clusterDynamicInformerFactory
	manager.istioDynamicInformerFactory = istioDynamicInformerFactory

	return manager
}

func (m *Manager) StartControllers() {

}
