// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2023-present Datadog, Inc.

package instrument

import (
	"context"
	"github.com/jonbodner/orchestrion/instrument/event"
	"net/http"
)

// if a function meets the handlerfunc type, insert code to:
// get the header from the request and look for the trace id
// if it's there but not in the context, add it to the context, add the context back to the request
// if it's not there and there's no traceid in the context, generate a guid, add it to the context, put the context back into the request
// output an "event" with a start message that has the method name, verb, id
// add a defer that outputs an event with an end message that has method name, verb, id
// can do this by having a function call that takes in the request and returns a request
/*
convert this:
func doThing(w http.ResponseWriter, r *http.Request) {
	// stuff here
}

to this:
func doThing(w http.ResponseWriter, r *http.Request) {
	//dd:startinstrument
	r = HandleHeader(r)
	Report(r.Context(), EventStart, "name", "doThing", "verb", r.Method)
	defer Report(r.Context(), EventEnd, "name", "doThing", "verb", r.Method)
	//dd:endinstrument
	// stuff here
}

Will need to properly capture the name of r from the function signature


For a client:
If you see a NewRequestWithContext or NewRequest call:
after the call,
- see if there's a traceid in the context
- if not add one and make a new context and request
- insert the header with the traceid

convert this:
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "localhost:8080", strings.NewReader(os.Args[1]))
	if err != nil {
		panic(err)
	}
	resp, err := client.Do(req)

to this:
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "localhost:8080", strings.NewReader(os.Args[1]))
	//dd:startinstrument
	if req != nil {
		req = InsertHeader(req)
		Report(req.Context(), EventCall, "url", req.URL, "method", req.Method)
		defer Report(req.Context(), EventReturn, "url", req.URL, "method", req.Method)
	}
	//dd:endinstrument
	if err != nil {
		panic(err)
	}
	resp, err := client.Do(req)

Will need to properly capture the name of req from the return values of the NewRequest/NewRequestWithContext call

Once we have this working for these simple cases, can work on harder ones!
*/

type Instrumenter interface {
	Init() func()
	InsertHeader(r *http.Request) *http.Request
	Report(ctx context.Context, e event.Event, metadata ...any) context.Context
	WrapHandlerFunc(handlerFunc http.HandlerFunc) http.HandlerFunc
	WrapHTTPClient(client *http.Client) *http.Client
	WrapHandler(handler http.Handler) http.Handler
}

type Key string

const (
	DD      Key = "dd"
	Console Key = "console"
	OTel    Key = "otel"
)

var instrumenters = map[Key]Instrumenter{
	DD:      DDInstrumenter{},
	Console: ConsoleInstrumenter{},
	OTel:    OTelInstrumenter{},
}

var instrumenter = instrumenters[DD]

func SetInstrumenter(key Key) {
	instrumenter = instrumenters[key]
	if instrumenter == nil {
		panic("unknown key: " + key)
	}
}

func InsertHeader(r *http.Request) *http.Request {
	return instrumenter.InsertHeader(r)
}

func Report(ctx context.Context, e event.Event, metadata ...any) context.Context {
	return instrumenter.Report(ctx, e, metadata)
}

func WrapHandlerFunc(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return instrumenter.WrapHandlerFunc(handlerFunc)
}

func WrapHandler(handler http.Handler) http.Handler {
	return instrumenter.WrapHandler(handler)
}

func WrapHTTPClient(client *http.Client) *http.Client {
	return instrumenter.WrapHTTPClient(client)
}

func Init(target string) func() {
	SetInstrumenter(Key(target))
	return instrumenter.Init()
}

const (
	EventStart    = event.EventStart
	EventEnd      = event.EventEnd
	EventCall     = event.EventCall
	EventReturn   = event.EventReturn
	EventDBCall   = event.EventDBCall
	EventDBReturn = event.EventDBReturn
)
