package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	appauth "github.com/slok/simple-ingress-external-auth/internal/app/auth"
	httpauthenticate "github.com/slok/simple-ingress-external-auth/internal/http/authenticate"
	"github.com/slok/simple-ingress-external-auth/internal/info"
	"github.com/slok/simple-ingress-external-auth/internal/log"
	loglogrus "github.com/slok/simple-ingress-external-auth/internal/log/logrus"
	metrics "github.com/slok/simple-ingress-external-auth/internal/metrics/prometheus"
	"github.com/slok/simple-ingress-external-auth/internal/storage/memory"
)

// Run runs the main application.
func Run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	// Ensure our context will end if any of the func uses as the main context.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Load command flags and arguments.
	cmdCfg, err := NewCmdConfig(args)
	if err != nil {
		return fmt.Errorf("could not load command configuration: %w", err)
	}

	// Set up logger.
	logrusLog := logrus.New()
	logrusLog.Out = stderr // By default logger goes to stderr (so it can split stdout prints).
	logrusLogEntry := logrus.NewEntry(logrusLog)
	if cmdCfg.Debug {
		logrusLogEntry.Logger.SetLevel(logrus.DebugLevel)
	}

	logger := loglogrus.NewLogrus(logrusLogEntry).WithValues(log.Kv{
		"app":     "simple-ingress-external-auth",
		"version": info.Version,
	})

	// Set up metrics with default metrics recorder.
	metricsRecorder := metrics.NewRecorder(prometheus.DefaultRegisterer)

	// Prepare our main runner.
	var g run.Group

	// Serving app HTTP server.
	{
		logger := logger.WithValues(log.Kv{"addr": cmdCfg.ListenAddress})

		configData := cmdCfg.TokenConfigData
		if cmdCfg.TokenConfigFile != "" {
			data, err := os.ReadFile(cmdCfg.TokenConfigFile)
			if err != nil {
				return fmt.Errorf("could not read token config file: %w", err)
			}
			configData = string(data)
		}

		// Create dependencies.
		repo, err := memory.NewTokenRepository(logger, configData)
		if err != nil {
			return fmt.Errorf("could not create memory token repository: %w", err)
		}

		appSvc := appauth.NewService(logger, metricsRecorder, repo)

		// Create server.
		handler := httpauthenticate.New(logger, metricsRecorder, appSvc)
		mux := http.NewServeMux()
		mux.Handle(cmdCfg.AuthenticationPath, handler)

		server := &http.Server{
			Addr:    cmdCfg.ListenAddress,
			Handler: mux,
		}

		g.Add(
			func() error {
				logger.Infof("HTTP server listening for requests")
				return server.ListenAndServe()
			},
			func(_ error) {
				logger.Infof("HTTP server shutdown, draining connections...")
				ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()
				err := server.Shutdown(ctx)
				if err != nil {
					logger.Errorf("error shutting down server: %w", err)
				}

				logger.Infof("Connections drained")
			},
		)
	}

	// Serving internal HTTP server.
	{
		logger := logger.WithValues(log.Kv{
			"addr":         cmdCfg.InternalListenAddr,
			"metrics":      cmdCfg.MetricsPath,
			"health-check": cmdCfg.HealthCheckPath,
			"pprof":        cmdCfg.PprofPath,
		})
		mux := http.NewServeMux()

		// Metrics.
		mux.Handle(cmdCfg.MetricsPath, promhttp.Handler())

		// Pprof.
		mux.HandleFunc(cmdCfg.PprofPath+"/", pprof.Index)
		mux.HandleFunc(cmdCfg.PprofPath+"/cmdline", pprof.Cmdline)
		mux.HandleFunc(cmdCfg.PprofPath+"/profile", pprof.Profile)
		mux.HandleFunc(cmdCfg.PprofPath+"/symbol", pprof.Symbol)
		mux.HandleFunc(cmdCfg.PprofPath+"/trace", pprof.Trace)

		// Health check.
		mux.Handle(cmdCfg.HealthCheckPath, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`{"status":"ok"}`)) }))

		// Create server.
		server := &http.Server{
			Addr:    cmdCfg.InternalListenAddr,
			Handler: mux,
		}

		g.Add(
			func() error {
				logger.Infof("http server listening for requests")
				return server.ListenAndServe()
			},
			func(_ error) {
				logger.Infof("http server shutdown, draining connections...")

				ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()
				err := server.Shutdown(ctx)
				if err != nil {
					logger.Errorf("error shutting down server: %w", err)
				}

				logger.Infof("connections drained")
			},
		)
	}

	// OS signals.
	{
		sigC := make(chan os.Signal, 1)
		exitC := make(chan struct{})
		signal.Notify(sigC, syscall.SIGTERM, syscall.SIGINT)

		g.Add(
			func() error {
				select {
				case s := <-sigC:
					logger.Infof("signal %s received", s)
					return nil
				case <-exitC:
					return nil
				}
			},
			func(_ error) {
				close(exitC)
			},
		)
	}

	err = g.Run()
	if err != nil {
		return err
	}

	return nil
}

func main() {
	ctx := context.Background()

	err := Run(ctx, os.Args, os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
