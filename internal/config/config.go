// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2023-present Datadog, Inc.

package config

import (
	"fmt"
	"strings"
)

// Config holds the instrumentation config
type Config struct {
	// HTTPMode controls the technique used for HTTP instrumentation
	// The possible values are "wrap", "report"
	HTTPMode string
	// Instrumentation specifies which output format is used
	// The possible values are "console", "dd", or "otel"
	Instrumentation string
}

var Default = Config{HTTPMode: "wrap", Instrumentation: "console"}

func (c *Config) Validate() error {
	c.HTTPMode = strings.ToLower(c.HTTPMode)
	switch c.HTTPMode {
	case "wrap", "report":
		// do nothing
	default:
		return fmt.Errorf("invalid httpmode %q, the supported values are wrap or report", c.HTTPMode)
	}
	c.Instrumentation = strings.ToLower(c.Instrumentation)
	switch c.Instrumentation {
	case "console", "dd", "otel":
		// do nothing
	default:
		return fmt.Errorf("invalid target %q, the supported values are console, dd, or otel", c.Instrumentation)
	}
	return nil
}
