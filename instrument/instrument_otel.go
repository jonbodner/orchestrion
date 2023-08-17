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

type OTelInstrumenter struct{}

func (o OTelInstrumenter) WrapHandler(handler http.Handler) http.Handler {
	//TODO implement me
	panic("implement me")
}

func (o OTelInstrumenter) WrapHTTPClient(client *http.Client) *http.Client {
	//TODO implement me
	panic("implement me")
}

func (o OTelInstrumenter) WrapHandlerFunc(handlerFunc http.HandlerFunc) http.HandlerFunc {
	//TODO implement me
	panic("implement me")
}

func (o OTelInstrumenter) Init() func() {
	return func() {}
}

func (o OTelInstrumenter) InsertHeader(r *http.Request) *http.Request {
	//TODO implement me
	return r
}

func (o OTelInstrumenter) Report(ctx context.Context, e event.Event, metadata ...any) context.Context {
	//TODO implement me
	return ctx
	// buildStackTrace := func() []uintptr {
	// 	pc := make([]uintptr, 2)
	// 	n := runtime.Callers(3, pc)
	// 	pc = pc[:n]
	// 	return pc
	// }

	// stackTrace := func(trace []uintptr) *runtime.Frames {
	// 	return runtime.CallersFrames(trace)
	// }

	// frames := stackTrace(buildStackTrace())
	// 	frame, _ := frames.Next()
	// 	file := ""
	// 	line := 0
	// 	funcName := ""
	// 	if frame.Func != nil {
	// 		file, line = frame.Func.FileLine(frame.PC)
	// 		funcName = frame.Func.Name()
	// 	}

	// in case we end up needing to walk further up, here's code to do that
	//for {
	//	frame, more := frames.Next()
	//	if frame.Func != nil {
	//		file, line := frame.Func.FileLine(frame.PC)
	//		fmt.Printf("Function %s in file %s on line %d\n", frame.Func.Name(),
	//			file, line)
	//	}
	//	if !more {
	//		break
	//	}
	//}

	// 	var s strings.Builder
	// 	s.WriteString(fmt.Sprintf(`{"time":"%s", "reportID":"%s", "event":"%s"`,
	// 		time.Now(), reportID, e))
	// 	s.WriteString(fmt.Sprintf(`, "function":"%s", "file":"%s", "line":%d`, funcName, file, line))
	// 	if len(metadata)%2 != 0 {
	// 		metadata = append(metadata, "")
	// 	}
	// 	for i := 0; i < len(metadata); i += 2 {
	// 		s.WriteString(fmt.Sprintf(`, "%s":"%s"`, metadata[i], metadata[i+1]))
	// 	}
	// 	s.WriteString("}")
	// 	fmt.Println(s.String())
}
