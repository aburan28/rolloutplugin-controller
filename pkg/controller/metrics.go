package controller

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	registry "k8s.io/component-base/metrics/legacyregistry"
)

const (
	// MetricsPath is the endpoint to collect rollout metrics
	MetricsPath = "/metrics"
)

var (
	buildInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "build_info",
			Help: "A metric with a constant '1' value labeled by version from which Argo-Rollouts was built.",
		},
		[]string{"version", "goversion", "goarch", "commit"},
	)
)

type MetricsServer struct {
	*http.Server
	reconcileRolloutHistogram *prometheus.HistogramVec
	rolloutPluginCounter      *prometheus.CounterVec
	rolloutPluginGauge        *prometheus.GaugeVec
}

type ServerConfig struct {
	Addr string
}

func NewMetricsServer(cfg ServerConfig) *MetricsServer {
	mux := http.NewServeMux()

	reg := prometheus.NewRegistry()
	reg.MustRegister(buildInfo)
	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	mux.Handle(MetricsPath, promhttp.HandlerFor(prometheus.Gatherers{
		reg,
		registry.DefaultGatherer,
	}, promhttp.HandlerOpts{}))

	return &MetricsServer{
		Server: &http.Server{
			Addr:    cfg.Addr,
			Handler: mux,
		},
		reconcileRolloutHistogram: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "rollout_reconcile_duration_seconds",
				Help: "Duration of reconciliation loops.",
			},
			[]string{"controller"},
		),
		rolloutPluginCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "rollout_plugin_total",
				Help: "Total number of plugin operations.",
			},
			[]string{"plugin", "operation"},
		),
		rolloutPluginGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "rollout_plugin",
				Help: "Current state of plugin operations.",
			},
			[]string{"plugin", "operation"},
		),
	}
}
