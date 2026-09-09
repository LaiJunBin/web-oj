package metrics

import (
	"net/http"
	"net/http/pprof"

	"judge-backend/internal/config"
	sharedtypes "judge-backend/internal/app/shared/types"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// StartMetricsServer starts a standalone HTTP server on port 9091 that exposes
// Prometheus metrics. Started as a plain goroutine because fx lifecycle OnStart
// hooks are never reached in this codebase — SetupHandler blocks in gin's
// r.Run() before fx can fire OnStart hooks.
// pprof endpoints are only registered when env != "production".
func StartMetricsServer(cfg *config.Config, logger sharedtypes.Logger) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	if !cfg.IsProduction() {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
		mux.HandleFunc("/debug/pprof/goroutine", pprof.Handler("goroutine").ServeHTTP)
		mux.HandleFunc("/debug/pprof/heap", pprof.Handler("heap").ServeHTTP)
		mux.HandleFunc("/debug/pprof/threadcreate", pprof.Handler("threadcreate").ServeHTTP)
		mux.HandleFunc("/debug/pprof/block", pprof.Handler("block").ServeHTTP)
		mux.HandleFunc("/debug/pprof/mutex", pprof.Handler("mutex").ServeHTTP)
	}

	// Bind to the configured host (127.0.0.1 by default) rather than all
	// interfaces, so /metrics and the non-production pprof endpoints are not
	// exposed on the network unless a deployment opts in.
	addr := cfg.Server.MetricsHost + ":9091"
	go func() {
		if cfg.IsProduction() {
			logger.Infof("[Metrics] Metrics server listening on %s/metrics", addr)
		} else {
			logger.Infof("[Metrics] Metrics server listening on %s/metrics and %s/debug/pprof/", addr, addr)
		}
		// Existing posture, now at least visible: on error the goroutine ends
		// and the metrics endpoint is gone for the life of the process — no
		// restart is attempted.
		if err := http.ListenAndServe(addr, mux); err != nil {
			logger.Errorf("[Metrics] Metrics server error: %v", err)
		}
	}()
}
