package main

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/cli/cli/version"
	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type shutdownFunc func(context.Context) error

const (
	// semconvServiceName is the well-known key for the service name.
	//
	// It's hardcoded here to avoid versioning issues from the semconv OTel package.
	semconvServiceName = "service.name"
	// semconvServiceVersion is the well-known key for the service version.
	//
	// It's hardcoded here to avoid versioning issues from the semconv OTel package.
	semconvServiceVersion = "service.version"
)

// initializeTracing configures OTel span exports if enabled via environment
// variables.
//
// See https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/
func initializeTracing(defaultServiceName string) (shutdownFunc, error) {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	exporter, err := autoexport.NewSpanExporter(context.Background())
	if err != nil {
		return nil, fmt.Errorf("creating span exporter: %w", err)
	}

	if autoexport.IsNoneSpanExporter(exporter) {
		// tracing is not configured, so there's no need to continue setup
		return nil, nil
	}

	attrs := []attribute.KeyValue{
		attribute.String(semconvServiceVersion, version.Version),
		attribute.String("build.git_commit", version.GitCommit),
		attribute.String("build.time", version.BuildTime),
		attribute.String("build.platform", version.PlatformName),
	}
	if v := os.Getenv("OTEL_SERVICE_NAME"); v == "" {
		// If unspecified, the default OTel detector will return a service name
		// in the format `unknown_service:docker` (for a binary name of "docker"),
		// so a service name is always explicitly provided to avoid that.
		attrs = append(attrs, attribute.String(semconvServiceName, defaultServiceName))
	}

	res, err := resource.Merge(resource.Default(), resource.NewSchemaless(attrs...))
	if err != nil {
		return nil, fmt.Errorf("merging resources: %w", err)
	}

	sp := sdktrace.NewBatchSpanProcessor(exporter)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(sp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	return tp.Shutdown, nil
}
