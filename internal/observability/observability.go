package observability

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/otlptranslator"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

const TracerName = "avatar-service"

type Config struct {
	ServiceName      string
	ServiceVersion   string
	Environment      string
	TracingEnabled   bool
	LogsEnabled      bool
	MetricsEnabled   bool
	OTLPEndpoint     string
	TraceSampleRatio float64
}

type Runtime struct {
	Shutdown       func(context.Context) error
	LogProvider    *log.LoggerProvider
	Kit            Kit
	MetricsHandler http.Handler
}

type metricsRuntime struct {
	provider *sdkmetric.MeterProvider
	recorder Recorder
	handler  http.Handler
}

func Init(ctx context.Context, cfg Config) (Runtime, error) {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	traceProvider := sdktrace.NewTracerProvider()
	meterProvider := sdkmetric.NewMeterProvider()
	otel.SetTracerProvider(traceProvider)
	otel.SetMeterProvider(meterProvider)

	if !cfg.TracingEnabled && !cfg.LogsEnabled && !cfg.MetricsEnabled {
		return disabledRuntime(traceProvider), nil
	}

	res, err := newResource(ctx, cfg)
	if err != nil {
		return Runtime{}, err
	}

	if cfg.TracingEnabled {
		traceProvider, err = newTraceProvider(ctx, cfg, res)
		if err != nil {
			return Runtime{}, err
		}
		otel.SetTracerProvider(traceProvider)
	}

	var logProvider *log.LoggerProvider
	if cfg.LogsEnabled {
		logProvider, err = newLogProvider(ctx, cfg, res)
		if err != nil {
			return Runtime{}, err
		}
	}

	var metrics Recorder
	var metricsHandler http.Handler
	if cfg.MetricsEnabled {
		started, err := newMetricsRuntime(res)
		if err != nil {
			return Runtime{}, err
		}
		meterProvider = started.provider
		metrics = started.recorder
		metricsHandler = started.handler
		otel.SetMeterProvider(meterProvider)
	}

	return Runtime{
		Shutdown:       shutdownRuntime(traceProvider, logProvider, meterProvider),
		LogProvider:    logProvider,
		Kit:            NewOTelKit(traceProvider.Tracer(TracerName), metrics),
		MetricsHandler: metricsHandler,
	}, nil
}

func disabledRuntime(traceProvider *sdktrace.TracerProvider) Runtime {
	return Runtime{
		Shutdown: func(context.Context) error { return nil },
		Kit:      NewOTelKit(traceProvider.Tracer(TracerName), NopRecorder),
	}
}

func newResource(ctx context.Context, cfg Config) (*resource.Resource, error) {
	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithContainer(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create otel resource: %w", err)
	}
	return res, nil
}

func newTraceProvider(ctx context.Context, cfg Config, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("create otlp trace exporter: %w", err)
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.TraceSampleRatio))),
	), nil
}

func newLogProvider(ctx context.Context, cfg Config, res *resource.Resource) (*log.LoggerProvider, error) {
	exporter, err := otlploggrpc.New(ctx,
		otlploggrpc.WithEndpoint(cfg.OTLPEndpoint),
		otlploggrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("create otlp log exporter: %w", err)
	}

	return log.NewLoggerProvider(
		log.WithResource(res),
		log.WithProcessor(log.NewBatchProcessor(exporter)),
	), nil
}

func newMetricsRuntime(res *resource.Resource) (metricsRuntime, error) {
	registry := prometheus.NewRegistry()
	exporter, err := otelprom.New(
		otelprom.WithRegisterer(registry),
		otelprom.WithTranslationStrategy(otlptranslator.UnderscoreEscapingWithoutSuffixes),
	)
	if err != nil {
		return metricsRuntime{}, fmt.Errorf("create prometheus exporter: %w", err)
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(exporter),
	)
	recorder, err := NewMetrics(provider.Meter(TracerName))
	if err != nil {
		return metricsRuntime{}, fmt.Errorf("init metrics instruments: %w", err)
	}

	return metricsRuntime{
		provider: provider,
		recorder: recorder,
		handler:  promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
	}, nil
}

func shutdownRuntime(
	traceProvider *sdktrace.TracerProvider,
	logProvider *log.LoggerProvider,
	meterProvider *sdkmetric.MeterProvider,
) func(context.Context) error {
	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		var shutdownErr error
		shutdownErr = errors.Join(shutdownErr, traceProvider.Shutdown(ctx))
		if logProvider != nil {
			shutdownErr = errors.Join(shutdownErr, logProvider.Shutdown(ctx))
		}
		return errors.Join(shutdownErr, meterProvider.Shutdown(ctx))
	}
}
