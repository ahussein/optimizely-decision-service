package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"contrib.go.opencensus.io/exporter/jaeger"
	"github.com/ahussein/optimizely-decision-service/cmd/api/handlers"
	"github.com/ahussein/optimizely-decision-service/internal/tracer"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/kelseyhightower/envconfig"
	optly "github.com/optimizely/go-sdk"
	"github.com/optimizely/go-sdk/pkg/client"
	"github.com/pkg/errors"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
)

type config struct {
	AppConfig
	JaegerConfig
	OptimizelyConfig
}

//OptimizelyConfig hold optimizley configurations
type OptimizelyConfig struct {
	SDKKey string `envconfig:"OPTIMIZELY_SDK_KEY" required:"true"`
}

// AppConfig holds application configurations
type AppConfig struct {
	Port            int           `default:"80" envconfig:"PORT" required:"true"`
	Env             string        `default:"development" envconfig:"ENV" required:"true"`
	DeploymentName  string        `envconfig:"DEPLOYMENT_NAME" required:"true"`
	ShutdownTimeout time.Duration `default:"5s" envconfig:"SHUTDOWN_TIMEOUT"`
	// HealthcheckTimeout time.Duration `default:"2s" envconfig:"HEALTHCHECK_TIMEOUT"`
}

// JaegerConfig holds Jaeger specific configurations
type JaegerConfig struct {
	Host         string  `envconfig:"JAEGER_AGENT_HOST"`
	Port         int     `envconfig:"JAEGER_AGENT_PORT"`
	SamplerParam float64 `envconfig:"JAEGER_SAMPLER_PARAM"`
	SamplerType  string  `envconfig:"JAEGER_SAMPLER_TYPE"`
}

func getLogger(env string) (*zap.Logger, error) {
	switch env {
	case "live":
		return zap.NewProduction()
	case "staging":
		cfg := zap.NewProductionConfig()
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		return cfg.Build()
	case "testing":
		return zap.NewNop(), nil
	default:
		return zap.NewDevelopment()
	}
}

// func initTracer() (trace.Provider, error) {
// 	// Create stdout exporter to be able to retrieve
// 	// the collected spans.
// 	exporter, err := stdout.NewExporter(stdout.Options{PrettyPrint: true})
// 	if err != nil {
// 		return nil, err
// 	}

// 	// For the demonstration, use sdktrace.AlwaysSample sampler to sample all traces.
// 	// In a production application, use sdktrace.ProbabilitySampler with a desired probability.
// 	tp, err := sdktrace.NewProvider(sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
// 		sdktrace.WithSyncer(exporter))
// 	if err != nil {
// 		return nil, err
// 	}
// 	global.SetTraceProvider(tp)
// 	return tp, nil
// }

// WithTracable adds tracing capabilities to the http handler
// func WithTracable(spanName string, t trace.Tracer) func(next http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		fn := func(w http.ResponseWriter, r *http.Request) {
// 			attrs, entries, spanCtx := httptrace.Extract(r.Context(), r)
// 			ctx := r.Context()
// 			if spanCtx.IsValid() {
// 				ctx = trace.ContextWithRemoteSpanContext(ctx, spanCtx)
// 			}

// 			// Apply the correlation context tags to the request
// 			// context.
// 			r = r.WithContext(correlation.ContextWithMap(ctx, correlation.NewMap(correlation.MapUpdate{
// 				MultiKV: entries,
// 			})))
// 			// Start the server-side span, passing the remote
// 			// child span context explicitly.
// 			_, span := t.Start(
// 				r.Context(),
// 				spanName,
// 				trace.WithAttributes(attrs...),
// 			)
// 			defer span.End()
// 			next.ServeHTTP(w, r)
// 		}
// 		return http.HandlerFunc(fn)
// 	}
// }

func initRoutes(cfg config, logger *zap.Logger) (*chi.Mux, error) {
	r := chi.NewRouter()
	r.Use(middleware.RedirectSlashes)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Mount("/debug", middleware.Profiler())

	r.Get("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Welcome to Optimizely Decision Service")
	}))
	tracedRoute(r, http.MethodPost, "/projects/{project_id}/experiment/activation", handlers.CreateExperimentActivationHandler(logger))
	return r, nil
}

func tracedRoute(r chi.Router, method, pattern string, h http.Handler) {
	spanName := fmt.Sprintf("%s %s", method, pattern)
	r.Method(method, pattern, tracer.TracingMiddleware(spanName)(h))
}

func initTracer(cfg config) error {
	// add tracing
	jaegerExporter, err := jaeger.NewExporter(jaeger.Options{
		AgentEndpoint: fmt.Sprintf("%s:%d", cfg.JaegerConfig.Host, cfg.JaegerConfig.Port),
		Process: jaeger.Process{
			ServiceName: cfg.AppConfig.DeploymentName,
		},
	})
	if err != nil {
		return errors.Wrap(err, "failed to create jeager exporter")
	}
	trace.RegisterExporter(jaegerExporter)
	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.ProbabilitySampler(cfg.JaegerConfig.SamplerParam),
	})
	return nil
}

func registerMonitoringViews() error {
	// add default views since ochttp.DefaultClientViews is deprecated
	allViews := []*view.View{
		ochttp.ServerRequestCountView,
		ochttp.ServerRequestBytesView,
		ochttp.ServerResponseBytesView,
		ochttp.ServerLatencyView,
		ochttp.ServerRequestCountByMethod,
		ochttp.ServerResponseCountByStatusCode,
		ochttp.ClientReceivedBytesDistribution,
		ochttp.ClientSentBytesDistribution,
		ochttp.ClientRoundtripLatencyDistribution,
	}

	return view.Register(allViews...)
}

func initOPtimizelyClient(sdkKey string) (*client.OptimizelyClient, error) {
	return optly.Client(sdkKey)
}

func main() {
	// load configurations
	var cfg config
	err := envconfig.Process("", &cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to load configurations %q", err))
	}

	logger, err := getLogger(cfg.AppConfig.Env)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger %q", err))
	}

	// Initialize tracing
	err = initTracer(cfg)
	if err != nil {
		logger.Error("failed to initialize tracer", zap.Error(err))
	}
	if err := registerMonitoringViews(); err != nil {
		logger.Error("register view stats", zap.Error(err))
	}

	// initialize optimizely client
	optlyClient, err := initOPtimizelyClient(cfg.OptimizelyConfig.SDKKey)
	if err != nil {
		logger.Fatal("initializing optimizely client", zap.Error(err))
	}
	logger.Info("initialized optimizley client", zap.Any("revision", optlyClient.GetOptimizelyConfig().Revision))

	// initialize routes
	r, err := initRoutes(cfg, logger)
	if err != nil {
		logger.Fatal("set up routes", zap.Error(err))
	}

	// Start API Service
	_, cancel := context.WithCancel(context.Background())
	addr := fmt.Sprintf(":%d", cfg.AppConfig.Port)
	server := http.Server{
		Addr: addr,
		Handler: &ochttp.Handler{
			Handler: r,
		},
	}
	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Start the service listening for requests.
	go func() {
		logger.Info("starting optimizely-decision-service-api", zap.String("address", addr))
		serverErrors <- server.ListenAndServe()
	}()

	// =========================================================================
	// Shutdown

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)

	// =========================================================================
	// Stop API Service

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		logger.Fatal("start service", zap.Error(err))

	case <-osSignals:
		logger.Info("shutting down...")
		cancel()

		// Create context for Shutdown call.
		ctx, cancel := context.WithTimeout(context.Background(), cfg.AppConfig.ShutdownTimeout)
		defer cancel()

		// Asking listener to shutdown and load shed.
		if err := server.Shutdown(ctx); err != nil {
			logger.Error("graceful shutdown", zap.Int("timeout", int(cfg.ShutdownTimeout)), zap.Error(err))
			if err := server.Close(); err != nil {
				logger.Fatal("stop http server", zap.Error(err))
			}
		}
	}

}
