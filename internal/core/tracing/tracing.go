package tracing

import (
	"context"
	"go.uber.org/zap"
	"io"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type Tracer struct {
	tp  *trace.TracerProvider
	log *zap.Logger
}

func NewTracer(appName string, log *zap.Logger) (*Tracer, error) {
	var tp *trace.TracerProvider

	exp, err := stdouttrace.New(
		stdouttrace.WithWriter(io.Discard),
		stdouttrace.WithPrettyPrint(),
		stdouttrace.WithoutTimestamps(),
	)
	if err != nil {
		return nil, err
	}

	tp = trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(newResource(appName)),
	)
	/*s.OnShutdown(func() {
		log.Info("shutting down tracing provider")

		if err := tp.Shutdown(ctx); err != nil {
			log.Error("failed to shutdown tracing provider")
		}
	})*/
	otel.SetTracerProvider(tp)

	return &Tracer{tp, log}, nil
}

func (t *Tracer) OnTracerShutdown() func(ctx context.Context) {

	return func(ctx context.Context) {
		t.log.Info("shutting down tracing provider")

		if err := t.tp.Shutdown(ctx); err != nil {
			t.log.Error("failed to shutdown tracing provider")
		}
	}
	/*t.log.Info("shutting down tracing provider")

	if err := t.tp.Shutdown(ctx); err != nil {
		t.log.Error("failed to shutdown tracing provider")
	}*/
}

func newResource(appName string) *resource.Resource {
	r, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(appName),
		),
	)
	return r
}

/*
// EnableTracing enables tracing.
func (t *Tracer) EnableTracing(ctx context.Context, appName string, s OnShutdowner) error {
	exp, err := stdouttrace.New(
		stdouttrace.WithWriter(io.Discard),
		stdouttrace.WithPrettyPrint(),
		stdouttrace.WithoutTimestamps(),
	)
	if err != nil {
		return err
	}

	t.tp = trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(newResource(appName)),
	)
	s.OnShutdown(func() {
		t.log.Info("shutting down tracing provider")

		if err := tp.Shutdown(ctx); err != nil {
			logging.From(ctx).Error("failed to shutdown tracing provider")
		}
	})
	otel.SetTracerProvider(tp)

	return nil
}*/
