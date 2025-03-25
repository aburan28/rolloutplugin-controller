package controller

type Config struct {
	WebhookAddr             string
	MetricsAddr             string
	ProbeAddr               string
	KubeConfig              string
	LogLevel                string
	LeaderElectionNamespace string
}
