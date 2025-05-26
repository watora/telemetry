package config

type Config struct {
	Init            bool
	AppName         string
	Version         string
	MetricsEndPoint string
	LogEndPoint     string
	UseMetrics      bool
	UseLogger       bool
	Env             string
	HostName        string
}
