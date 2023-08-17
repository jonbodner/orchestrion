// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2023-present Datadog, Inc.

package instrument

import (
	"context"
	"github.com/jonbodner/orchestrion/instrument/event"
	"io"
	"testing"

	"github.com/jonbodner/orchestrion/internal/config"
	"github.com/jonbodner/orchestrion/internal/instrument"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func TestScanPackageDST(t *testing.T) {
	output := func(fullName string, out io.Reader) {
		io.ReadAll(out)
	}
	instrument.ProcessPackage("./samples", instrument.InstrumentFile, output, config.Default)
}

func TestReport(t *testing.T) {
	t.Run("start", func(t *testing.T) {
		ctx := context.Background()
		ctx = Report(ctx, event.EventStart)
		if _, ok := tracer.SpanFromContext(ctx); !ok {
			t.Errorf("Expected Report of StartEvent to generate a new ID.")
		}
	})

	t.Run("call", func(t *testing.T) {
		ctx := context.Background()
		ctx = Report(ctx, event.EventCall)
		if _, ok := tracer.SpanFromContext(ctx); !ok {
			t.Errorf("Expected Report of CallEvent to generate a new ID.")
		}
	})

	t.Run("end", func(t *testing.T) {
		ctx := context.Background()
		ctx = Report(ctx, event.EventEnd)
		if _, ok := tracer.SpanFromContext(ctx); ok {
			t.Errorf("Expected Report of EndEvent not to generate a new ID.")
		}
	})

	t.Run("return", func(t *testing.T) {
		ctx := context.Background()
		ctx = Report(ctx, event.EventReturn)
		if _, ok := tracer.SpanFromContext(ctx); ok {
			t.Errorf("Expected Report of ReturnEvent not to generate a new ID.")
		}
	})
}
