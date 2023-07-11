// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2023-present Datadog, Inc.

package instrument

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"github.com/jonbodner/orchestrion/instrument/event"
	"google.golang.org/grpc"
	sqltrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/database/sql"
	grpctrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/google.golang.org/grpc"
	httptrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"net/http"
)

type DDInstrumenter struct{}

func (D DDInstrumenter) Init() func() {
	tracer.Start()
	return tracer.Stop
}

func (D DDInstrumenter) InsertHeader(r *http.Request) *http.Request {
	span, ok := tracer.SpanFromContext(r.Context())
	if !ok {
		return r
	}
	r = r.Clone(r.Context())
	tracer.Inject(span.Context(), tracer.HTTPHeadersCarrier(r.Header))
	return r
}

func (D DDInstrumenter) Report(ctx context.Context, e event.Event, metadata ...any) context.Context {
	var span tracer.Span
	if e == event.EventStart || e == event.EventCall {
		var opts []tracer.StartSpanOption
		for i := 0; i < len(metadata); i += 2 {
			if i+1 >= len(metadata) {
				break
			}
			if k, ok := metadata[i].(string); ok {
				opts = append(opts, tracer.Tag(k, metadata[i+1]))
			}
		}
		span, ctx = tracer.StartSpanFromContext(ctx, getOpName(metadata...), opts...)
	} else if e == event.EventEnd || e == event.EventReturn {
		var ok bool
		span, ok = tracer.SpanFromContext(ctx)
		if !ok {
			fmt.Printf("Error: Received end/return event but have no corresponding span in the context.\n")
			return ctx
		}
		span.Finish()
	}
	return ctx
}

func getOpName(metadata ...any) string {
	rank := map[string]int{
		"verb":          1,
		"function-name": 2,
	}

	var (
		opname string
		oprank int = 10_000 // just a higher number than any key in the rank map.
	)
	for i := 0; i < len(metadata); i += 2 {
		if i+1 >= len(metadata) {
			break
		}
		if k, ok := metadata[i].(string); ok {
			if r, ok := rank[k]; ok && r < oprank {
				if on, ok := metadata[i+1].(string); ok {
					opname = on
					oprank = r
					continue
				}
			}
		}
	}
	return opname
}

func WrapHandler(handler http.Handler) http.Handler {
	return httptrace.WrapHandler(handler, "", "")
	// TODO: We'll reintroduce this later when we stop hard-coding dd-trace-go as above.
	//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//		r = HandleHeader(r)
	//		r = r.WithContext(Report(r.Context(), EventStart, "name", "FooHandler", "verb", r.Method))
	//		defer Report(r.Context(), EventEnd, "name", "FooHandler", "verb", r.Method)
	//		handler.ServeHTTP(w, r)
	//	})
}

func WrapHandlerFunc(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httptrace.TraceAndServe(handlerFunc, w, r, &httptrace.ServeConfig{})
	})
	// TODO: We'll reintroduce this later when we stop hard-coding dd-trace-go as above.
	//	return func(w http.ResponseWriter, r *http.Request) {
	//		r = HandleHeader(r)
	//		r = r.WithContext(Report(r.Context(), EventStart, "name", "FooHandler", "verb", r.Method))
	//		defer Report(r.Context(), EventEnd, "name", "FooHandler", "verb", r.Method)
	//		handlerFunc(w, r)
	//	}
}

func WrapHTTPClient(client *http.Client) *http.Client {
	// TODO: Stop hard-coding dd-trace-go.
	return httptrace.WrapClient(client)
}

func GRPCStreamServerInterceptor() grpc.ServerOption {
	return grpc.StreamInterceptor(grpctrace.StreamServerInterceptor())
}

func GRPCUnaryServerInterceptor() grpc.ServerOption {
	return grpc.UnaryInterceptor(grpctrace.UnaryServerInterceptor())
}

func GRPCStreamClientInterceptor() grpc.DialOption {
	return grpc.WithStreamInterceptor(grpctrace.StreamClientInterceptor())
}

func GRPCUnaryClientInterceptor() grpc.DialOption {
	return grpc.WithUnaryInterceptor(grpctrace.UnaryClientInterceptor())
}

func Open(driverName, dataSourceName string) (*sql.DB, error) {
	return sqltrace.Open(driverName, dataSourceName)
}

func OpenDB(c driver.Connector) *sql.DB {
	return sqltrace.OpenDB(c)
}
