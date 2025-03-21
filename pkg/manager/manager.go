package manager

import (
	"context"
	"fmt"
	"net/http"
	"rolloutplugin-controller/api/v1alpha1"
	"rolloutplugin-controller/pkg/controller"
	"sync"
	"time"

	"github.com/argoproj/argo-rollouts/utils/record"
	"github.com/pkg/errors"

	"github.com/argoproj/argo-cd/server/metrics"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/kubectl/pkg/scheme"
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
	rolloutPluginSynced           cache.InformerSynced
	kubeClientSet                 kubernetes.Interface
	metricsServer                 *metrics.MetricsServer
	healthzServer                 *http.Server
	informer                      cache.SharedIndexInformer
	indexer                       cache.Indexer
	rolloutPluginWorkqueue        workqueue.TypedRateLimitingInterface[string]
	dynamicInformerFactory        dynamicinformer.DynamicSharedInformerFactory
	clusterDynamicInformerFactory dynamicinformer.DynamicSharedInformerFactory
	istioDynamicInformerFactory   dynamicinformer.DynamicSharedInformerFactory
	rolloutPluginController       controller.RolloutPluginController
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

func (m *Manager) Start(ctx context.Context, rolloutPluginThreadiness int, electOpts *LeaderElectionOptions) error {

	defer runtime.HandleCrash()
	defer m.rolloutPluginWorkqueue.ShutDown()
	go func() {
		log.Infof("Starting Healthz Server at %s", c.healthzServer.Addr)
		err := c.healthzServer.ListenAndServe()
		if err != nil {
			err = errors.Wrap(err, "Healthz Server Error")
			log.Error(err)
		}
	}()

	go func() {
		log.Infof("Starting Metric Server at %s", c.metricsServer.Addr)
		if err := c.metricsServer.ListenAndServe(); err != nil {
			log.Error(errors.Wrap(err, "Metric Server Error"))
		}
	}()

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second) // give max of 10 seconds for http servers to shut down
	defer cancel()
	m.healthzServer.Shutdown(ctxWithTimeout)
	m.metricsServer.Shutdown(ctxWithTimeout)

	m.wg.Wait()

	return nil
}
func DefaultArgoRolloutsRateLimiter() workqueue.RateLimiter {
	return workqueue.NewItemExponentialFailureRateLimiter(time.Millisecond, 10*time.Second)
}

func NewManager(kubeclientset kubernetes.Interface, metricsPort int,
	healthzPort int) *Manager {

	runtime.Must(v1alpha1.AddToScheme(scheme.Scheme))

	recorder := record.NewEventRecorder(kubeclientset, metrics.MetricRolloutEventsTotal, metrics.MetricNotificationFailedTotal, metrics.MetricNotificationSuccessTotal, metrics.MetricNotificationSend, nil)
	healthzServer := controller.NewHealthzServer(fmt.Sprintf(listenAddr, healthzPort))
	metricsServer := metrics.NewMetricsServer(fmt.Sprintf(listenAddr, metricsPort))

	rolloutWorkqueue := workqueue.NewNamedRateLimitingQueue(DefaultArgoRolloutsRateLimiter(), "RolloutPlugin")

	return &Manager{
		wg:                      &sync.WaitGroup{},
		kubeClientSet:           kubeclientset,
		metricsServer:           metricsServer,
		healthzServer:           healthzServer,
		rolloutPluginWorkqueue:  rolloutWorkqueue,
		rolloutPluginController: controller.NewRolloutPluginController(kubeclientset, scheme.Scheme, recorder, rolloutWorkqueue),
	}
}
