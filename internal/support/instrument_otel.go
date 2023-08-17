// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2023-present Datadog, Inc.

package support

import (
	"context"
	"github.com/jonbodner/orchestrion/instrument/event"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/contrib/propagators/autoprop"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/trace"
	"log"
	"net/http"
	"os"
)

type OTelInstrumenter struct {
	Tracer   trace.Tracer
	RootSpan trace.Span
}

func (o *OTelInstrumenter) WrapHandler(handler http.Handler) http.Handler {
	return otelhttp.NewHandler(handler, "")
}

func (o *OTelInstrumenter) WrapHTTPClient(client *http.Client) *http.Client {
	if client.Transport == nil {
		client.Transport = http.DefaultTransport
	}
	client.Transport = otelhttp.NewTransport(client.Transport)
	return client
}

func (o *OTelInstrumenter) WrapHandlerFunc(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		otelhttp.NewHandler(handlerFunc, "").ServeHTTP(rw, r)
	}
}

func (o *OTelInstrumenter) Init() func() {

	tp, err := tracerProvider("http://localhost:14268/api/traces")
	if err != nil {
		log.Fatal(err)
	}

	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)

	otel.SetTextMapPropagator(autoprop.NewTextMapPropagator())

	o.Tracer = otel.Tracer("")
	return func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Fatal(err)
		}
	}
}

// tracerProvider returns an OpenTelemetry TracerProvider configured to use
// the Jaeger exporter that will send spans to the provided url. The returned
// TracerProvider will also use a Resource configured with all the information
// about the application.
func tracerProvider(url string) (*tracesdk.TracerProvider, error) {
	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(os.Args[0]),
			attribute.String("environment", "demo"),
			attribute.Int64("ID", 1),
		)),
	)
	return tp, nil
}

func (o *OTelInstrumenter) InsertHeader(r *http.Request) *http.Request {
	//TODO implement me
	return r
}

func (o *OTelInstrumenter) Report(ctx context.Context, e event.Event, metadata ...any) context.Context {
	var span trace.Span
	switch e {
	case event.EventStart:
		ctx, span = o.Tracer.Start(ctx, metadata[1].(string))
		for i := 0; i < len(metadata); i += 2 {
			span.SetAttributes(attribute.String(metadata[i].(string), metadata[i+1].(string)))
		}
	case event.EventEnd:
		span = trace.SpanFromContext(ctx)
		defer span.End()
	}
	return ctx
}
