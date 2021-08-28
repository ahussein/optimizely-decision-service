package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"contrib.go.opencensus.io/exporter/jaeger"
	pb "github.com/ahussein/optimizely-decision-service/cmd/grpc/proto"
	"github.com/kelseyhightower/envconfig"
	optly "github.com/optimizely/go-sdk"
	"github.com/optimizely/go-sdk/pkg/client"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
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
	Port            string        `default:":50051" envconfig:"GRPC_PORT" required:"true"`
	Env             string        `default:"development" envconfig:"ENV" required:"true"`
	DeploymentName  string        `default:"optimizely-decision-service" envconfig:"DEPLOYMENT_NAME" required:"true"`
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

func initOPtimizelyClient(sdkKey string) (*client.OptimizelyClient, error) {
	return optly.Client(sdkKey)
}

// server is used to implement an interface for optimizely experiments/features
type server struct {
	pb.ExperimentServer
	OptlyClient *client.OptimizelyClient
	logger      *zap.Logger
}

// Activate returns the key of the variation the user is bucketed into and queues up an impression event to be sent to
// the Optimizely log endpoint for results processing.
func (s *server) Activate(ctx context.Context, in *pb.ActivateRequest) (*pb.Variation, error) {
	ek := in.ExperimentKey
	var userAttributes map[string]interface{}
	attr, err := in.Attributes.MarshalJSON()
	if err != nil {
		s.logger.Error("reading user attributes", zap.String("experiment_key", ek), zap.Any("attributes", in.Attributes), zap.Error(err))
		return nil, err
	}
	err = json.Unmarshal(attr, &userAttributes)
	if err != nil {
		s.logger.Error("loading user attributes", zap.String("experiment_key", ek), zap.Any("attributes", in.Attributes), zap.Error(err))
		return nil, err
	}
	uc := entities.UserContext{
		ID:         in.UserId,
		Attributes: userAttributes,
	}
	variation, err := s.OptlyClient.Activate(ek, uc)
	if err != nil {
		s.logger.Error("activating user", zap.String("experiment_key", ek), zap.Any("user", uc), zap.Error(err))
		return nil, err
	}
	return &pb.Variation{
		Variation: variation,
	}, nil
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

	// initialize optimizely client
	optlyClient, err := initOPtimizelyClient(cfg.OptimizelyConfig.SDKKey)
	if err != nil {
		logger.Fatal("initializing optimizely client", zap.Error(err))
	}
	logger.Info("initialized optimizley client", zap.Any("revision", optlyClient.GetOptimizelyConfig().Revision))

	lis, err := net.Listen("tcp", cfg.AppConfig.Port)
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}
	s := grpc.NewServer()
	pb.RegisterExperimentServer(s, &server{
		OptlyClient: optlyClient,
		logger:      logger,
	})
	if err := s.Serve(lis); err != nil {
		logger.Fatal("failed to server", zap.Error(err))
	}
}
