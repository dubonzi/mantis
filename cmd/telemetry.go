package main

import (
	"context"

	"github.com/americanas-go/config"
	"github.com/americanas-go/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace/noop"
)

const (
	protocolHTTP = "http"
	protocolGRPC = "grpc"
)

func setupOtel() {
	if !config.Bool("otel.enabled") {
		return
	}

	exporterProtocol := config.String("otel.exporter.protocol")
	var err error
	var traceExporter *otlptrace.Exporter

	switch exporterProtocol {
	case protocolHTTP:
		opts := []otlptracehttp.Option{otlptracehttp.WithEndpoint(config.String("otel.exporter.endpoint"))}
		if config.Bool("otel.exporter.insecure") {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		traceExporter, err = otlptracehttp.New(context.Background(), opts...)
	case protocolGRPC:
		opts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(config.String("otel.exporter.endpoint"))}
		if config.Bool("otel.exporter.insecure") {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}
		traceExporter, err = otlptracegrpc.New(context.Background(), opts...)
	default:
		log.Warnf("opentelemetry: invalid protocol: %s", exporterProtocol)
		otel.SetTracerProvider(noop.NewTracerProvider())
		return
	}
	if err != nil {
		log.Error("error starting opentelemetry exporter", err)
		return
	}

	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithBatcher(traceExporter),
		trace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String("mantis"),
			)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
}
