package observability

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
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
	Shutdown    func(context.Context) error
	LogProvider *log.LoggerProvider
}

func Init(ctx context.Context, cfg Config) (Runtime, error) {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	if !cfg.TracingEnabled && !cfg.LogsEnabled && !cfg.MetricsEnabled {
		otel.SetTracerProvider(sdktrace.NewTracerProvider())
		otel.SetMeterProvider(sdkmetric.NewMeterProvider())
		return Runtime{Shutdown: func(context.Context) error { return nil }}, nil
	}

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
		return Runtime{}, fmt.Errorf("create otel resource: %w", err)
	}

	var (
		traceProvider *sdktrace.TracerProvider
		logProvider   *log.LoggerProvider
		meterProvider *sdkmetric.MeterProvider
	)

	if cfg.TracingEnabled {
		traceExporter, err := otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
			otlptracegrpc.WithInsecure(),
		)
		if err != nil {
			return Runtime{}, fmt.Errorf("create otlp trace exporter: %w", err)
		}

		traceProvider = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(traceExporter),
			sdktrace.WithResource(res),
			sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.TraceSampleRatio))),
		)
		otel.SetTracerProvider(traceProvider)
	} else {
		otel.SetTracerProvider(sdktrace.NewTracerProvider())
	}

	if cfg.LogsEnabled {
		logExporter, err := otlploggrpc.New(ctx,
			otlploggrpc.WithEndpoint(cfg.OTLPEndpoint),
			otlploggrpc.WithInsecure(),
		)
		if err != nil {
			return Runtime{}, fmt.Errorf("create otlp log exporter: %w", err)
		}

		logProvider = log.NewLoggerProvider(
			log.WithResource(res),
			log.WithProcessor(log.NewBatchProcessor(logExporter)),
		)
	}

	if cfg.MetricsEnabled {
		metricExporter, err := otlpmetricgrpc.New(ctx,
			otlpmetricgrpc.WithEndpoint(cfg.OTLPEndpoint),
			otlpmetricgrpc.WithInsecure(),
		)
		if err != nil {
			return Runtime{}, fmt.Errorf("create otlp metric exporter: %w", err)
		}

		meterProvider = sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(res),
			sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		)
		otel.SetMeterProvider(meterProvider)

		if err := InitMetrics(); err != nil {
			return Runtime{}, fmt.Errorf("init metrics instruments: %w", err)
		}
	} else {
		otel.SetMeterProvider(sdkmetric.NewMeterProvider())
	}

	shutdown := func(shutdownCtx context.Context) error {
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 5*time.Second)
		defer cancel()

		var shutdownErr error
		if traceProvider != nil {
			shutdownErr = errors.Join(shutdownErr, traceProvider.Shutdown(shutdownCtx))
		}
		if logProvider != nil {
			shutdownErr = errors.Join(shutdownErr, logProvider.Shutdown(shutdownCtx))
		}
		if meterProvider != nil {
			shutdownErr = errors.Join(shutdownErr, meterProvider.Shutdown(shutdownCtx))
		}
		return shutdownErr
	}

	return Runtime{
		Shutdown:    shutdown,
		LogProvider: logProvider,
	}, nil
}
