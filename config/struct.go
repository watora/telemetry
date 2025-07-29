package config

type Config struct {
	AppName         string
	Version         string
	MetricsEndPoint string
	LogEndPoint     string
	UseMetrics      bool
	UseLogger       bool
	Env             string
	HostName        string
}
