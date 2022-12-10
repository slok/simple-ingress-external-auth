package main

import (
	"fmt"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/slok/simple-ingress-external-auth/internal/info"
)

// CmdConfig represents the configuration of the command.
type CmdConfig struct {
	Debug              bool
	ListenAddress      string
	AuthenticationPath string
	TokenConfigData    string
	TokenConfigFile    string
	InternalListenAddr string
	MetricsPath        string
	HealthCheckPath    string
	PprofPath          string
	ClientIDHeader     string
}

// NewCmdConfig returns a new command configuration.
func NewCmdConfig(args []string) (*CmdConfig, error) {
	c := &CmdConfig{}
	app := kingpin.New("simple-ingress-external-auth", "Simple external authentication service for Kubernetes ingresses.")
	app.DefaultEnvars()
	app.Version(info.Version)

	// General.
	app.Flag("debug", "Enable debug mode.").BoolVar(&c.Debug)

	// App.
	app.Flag("listen-address", "The address where the HTTP API server will be listening.").Default(":8080").StringVar(&c.ListenAddress)
	app.Flag("authentication-path", "The path user for authenticating then tokens.").Default("/auth").StringVar(&c.AuthenticationPath)
	app.Flag("token-config-data", "The raw data token configuration.").StringVar(&c.TokenConfigData)
	app.Flag("token-config-file", "The raw data token configuration file (can't be used with token-config-data).").StringVar(&c.TokenConfigFile)
	app.Flag("client-id-header", "Return the client id as a custom header").Default("X-Ext-Auth-Client-Id").StringVar(&c.ClientIDHeader)

	// Internal.
	app.Flag("internal-listen-address", "The address where the HTTP internal data (metrics, pprof...) server will be listening.").Default(":8081").StringVar(&c.InternalListenAddr)
	app.Flag("metrics-path", "the path where Prometehus metrics will be served.").Default("/metrics").StringVar(&c.MetricsPath)
	app.Flag("health-check-path", "the path where the health check will be served.").Default("/status").StringVar(&c.HealthCheckPath)
	app.Flag("pprof-path", "the path where the pprof handlers will be served.").Default("/debug/pprof").StringVar(&c.PprofPath)

	_, err := app.Parse(args[1:])
	if err != nil {
		return nil, err
	}

	// Check.
	if c.TokenConfigFile == "" && c.TokenConfigData == "" {
		return nil, fmt.Errorf("one of token config file or token config data is required")
	}

	if c.TokenConfigFile != "" && c.TokenConfigData != "" {
		return nil, fmt.Errorf("token config file and token config data can't be used at the same time")
	}

	return c, nil
}
