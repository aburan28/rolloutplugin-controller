package controller

type Config struct {
	WebhookAddr             string
	MetricsAddr             string
	ProbeAddr               string
	KubeConfig              string
	LogLevel                string
	LeaderElectionNamespace string
	IstioEnabled            bool
}

const (
	FinalizerName = "rolloutsplugin.io/finalizer"
)
