// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2023-present Datadog, Inc.

package support

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jonbodner/orchestrion/instrument/event"
	"net/http"
	"os"
	"time"
)

type ConsoleInstrumenter struct{}

type field int

const (
	_                 field = iota
	traceIDField      field = iota
	parentSpanIDField field = iota
	spanIDField       field = iota
)

func addFieldToContext(ctx context.Context, f field, v string) context.Context {
	return context.WithValue(ctx, f, v)
}

func getFieldFromContext(ctx context.Context, f field) string {
	val, _ := ctx.Value(f).(string)
	return val
}

const (
	traceHeader      = "X-Trace-ID"
	parentSpanHeader = "X-Parent-Span-ID"
)

func makeID() string {
	return uuid.NewString()
}

func (c ConsoleInstrumenter) WrapHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// check for incoming trace id and parent span id in request header
		traceID := r.Header.Get(traceHeader)
		parentSpanID := r.Header.Get(parentSpanHeader)

		// if not there, create traceID
		if traceID == "" {
			traceID = makeID()
		}

		// make span id
		spanID := makeID()
		// put trace id, parent span id, span id in context
		ctx = addFieldToContext(ctx, traceIDField, traceID)
		ctx = addFieldToContext(ctx, parentSpanIDField, parentSpanID)
		ctx = addFieldToContext(ctx, spanIDField, spanID)

		r = r.WithContext(ctx)
		// print out the values
		fmt.Fprintf(os.Stderr, "%s: %s server trace_id=%q, parent_span_id=%q, span_id=%q\n", time.Now().UTC().Format(time.RFC3339Nano), event.EventStart, traceID, parentSpanID, spanID)
		// defer printing out that we're done
		defer fmt.Fprintf(os.Stderr, "%s: %s server trace_id=%q, parent_span_id=%q, span_id=%q\n", time.Now().UTC().Format(time.RFC3339Nano), event.EventEnd, traceID, parentSpanID, spanID)
		handler.ServeHTTP(rw, r)
	})
}

type ConsoleRoundTripper struct {
	internalRoundTripper http.RoundTripper
}

func (c *ConsoleRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	ctx := getOrBuildIDs(r.Context())
	traceID := getFieldFromContext(ctx, traceIDField)
	parentSpanID := getFieldFromContext(ctx, parentSpanIDField)
	spanID := getFieldFromContext(ctx, spanIDField)
	r = r.Clone(ctx)
	r.Header.Add(traceHeader, traceID)
	// current span becomes the parent
	r.Header.Add(parentSpanHeader, spanID)
	// print out the values
	fmt.Fprintf(os.Stderr, "%s: %s client trace_id=%q, parent_span_id=%q, span_id=%q\n", time.Now().UTC().Format(time.RFC3339Nano), event.EventStart, traceID, parentSpanID, spanID)
	// defer printing out that we're done
	defer fmt.Fprintf(os.Stderr, "%s: %s client trace_id=%q, parent_span_id=%q, span_id=%q\n", time.Now().UTC().Format(time.RFC3339Nano), event.EventEnd, traceID, parentSpanID, spanID)
	return c.internalRoundTripper.RoundTrip(r)
}

func getOrBuildIDs(ctx context.Context) context.Context {
	// pull ids from context
	traceID := getFieldFromContext(ctx, traceIDField)
	// if not there, create and add
	if traceID == "" {
		traceID = makeID()
		ctx = context.WithValue(ctx, traceIDField, traceID)
	}
	// the current span id becomes the parent span id
	parentSpanID := getFieldFromContext(ctx, spanIDField)
	// if not there, do nothing
	// add it
	ctx = context.WithValue(ctx, parentSpanIDField, parentSpanID)

	// make a new span id and add
	spanID := makeID()
	ctx = context.WithValue(ctx, spanIDField, spanID)

	return ctx
}

func (c ConsoleInstrumenter) WrapHTTPClient(client *http.Client) *http.Client {
	if client.Transport == nil {
		client.Transport = http.DefaultTransport
	}

	return &http.Client{
		Transport:     &ConsoleRoundTripper{client.Transport},
		CheckRedirect: client.CheckRedirect,
		Jar:           client.Jar,
		Timeout:       client.Timeout,
	}
}

func (c ConsoleInstrumenter) WrapHandlerFunc(handlerFunc http.HandlerFunc) http.HandlerFunc {
	// I know it's a HandlerFunc
	return c.WrapHandler(handlerFunc).(http.HandlerFunc)
}

func (c ConsoleInstrumenter) Init() func() {
	return func() {}
}

func (c ConsoleInstrumenter) InsertHeader(r *http.Request) *http.Request {
	// unneeded
	return r
}

func (c ConsoleInstrumenter) Report(ctx context.Context, e event.Event, metadata ...any) context.Context {
	if e == event.EventStart {
		ctx = getOrBuildIDs(ctx)
	}
	traceID := getFieldFromContext(ctx, traceIDField)
	parentSpanID := getFieldFromContext(ctx, parentSpanIDField)
	spanID := getFieldFromContext(ctx, spanIDField)
	// print out the values
	fmt.Fprintf(os.Stderr, "%s: %s report trace_id=%q, parent_span_id=%q, span_id=%q", time.Now().UTC().Format(time.RFC3339Nano), e, traceID, parentSpanID, spanID)
	for i := 0; i < len(metadata); i += 2 {
		fmt.Fprintf(os.Stderr, " %v=%v", metadata[i], metadata[i+1])
	}
	fmt.Fprintln(os.Stderr)
	return ctx
}
