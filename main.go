// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2023-present Datadog, Inc.

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jonbodner/orchestrion/internal/config"
	"github.com/jonbodner/orchestrion/internal/instrument"
)

func main() {
	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprint(w, "usage: orchestrion [options] [path]\n")
		fmt.Fprint(w, "example: orchestrion -w ./\n")
		fmt.Fprint(w, "options:\n")
		flag.PrintDefaults()
	}
	var write bool
	var remove bool
	var tool bool
	var httpMode string
	var target string
	flag.BoolVar(&remove, "rm", false, "remove all instrumentation from the package")
	flag.BoolVar(&write, "w", false, "if set, overwrite the current file with the instrumented file")
	flag.BoolVar(&tool, "t", false, "if set, run in toolexec mode")
	flag.StringVar(&httpMode, "httpmode", "wrap", "set the http instrumentation mode: wrap (default) or report")
	flag.StringVar(&target, "target", "console", "set the target instrumentation type: console (default), dd, or otel")
	flag.Parse()
	if len(flag.Args()) == 0 {
		return
	}
	output := func(fullName string, out io.Reader) {
		fmt.Printf("%s:\n", fullName)
		// write the output
		txt, _ := io.ReadAll(out)
		fmt.Println(string(txt))
	}
	if write || tool {
		output = func(fullName string, out io.Reader) {
			fmt.Printf("overwriting %s:\n", fullName)
			// write the output
			txt, _ := io.ReadAll(out)
			err := os.WriteFile(fullName, txt, 0644)
			if err != nil {
				fmt.Printf("Writing file %s: %v\n", fullName, err)
			}
		}
	}
	conf := config.Config{HTTPMode: httpMode, Instrumentation: target}
	if err := conf.Validate(); err != nil {
		fmt.Printf("Config error: %v\n", err)
		os.Exit(1)
	}
	if tool {
		path := os.Args[2]
		err := runToolexecMode(path, conf, output)
		if err != nil {
			fmt.Printf("toolexec error: %v\n", err)
			os.Exit(1)
		}
		return
	}
	for _, v := range flag.Args() {
		p, err := filepath.Abs(v)
		if err != nil {
			fmt.Printf("Sanitizing path (%s) failed: %v\n", v, err)
			continue
		}
		fmt.Printf("Scanning Package %s\n", p)
		processor := instrument.InstrumentFile
		if remove {
			fmt.Printf("Removing Orchestrion instrumentation.\n")
			processor = instrument.UninstrumentFile
		}
		err = instrument.ProcessPackage(p, processor, output, conf)
		if err != nil {
			fmt.Printf("Failed to scan: %v\n", err)
			os.Exit(1)
		}
	}
}

func runToolexecMode(path string, conf config.Config, output func(fullName string, out io.Reader)) error {
	tool, args := os.Args[3], os.Args[4:]
	toolName := filepath.Base(tool)
	if len(args) > 0 && args[0] == "-V=full" {
		// We can't alter the version output.
	} else {
		fmt.Println(os.Args)
		if toolName == "compile" {
			tmpDir, err := os.MkdirTemp("", "orchestrion")
			if err != nil {
				return err
			}
			defer os.RemoveAll(tmpDir)
			newArgs := make([]string, 0, len(args))
			for _, v := range args {
				if strings.HasPrefix(v, path) && strings.HasSuffix(v, ".go") {
					fmt.Println("modifying:", v)

					otherPath, err := filepath.Abs(v)

					if err != nil {
						return fmt.Errorf("Sanitizing path (%s) failed: %v\n", v, err)
					}
					file, err := os.Open(v)
					if err != nil {
						return fmt.Errorf("error opening file: %w", err)
					}
					out, err := instrument.InstrumentFile(otherPath, file, conf)
					file.Close()
					if err != nil {
						return fmt.Errorf("error scanning file %s: %w", path, err)
					}
					newFileName := tmpDir + string(os.PathSeparator) + filepath.Base(otherPath)
					if out != nil {
						output(newFileName, out)
					}
					newArgs = append(newArgs, newFileName)
				} else {
					newArgs = append(newArgs, v)
				}
			}
			args = newArgs
		}
	}
	// Simply run the tool.
	cmd := exec.Command(tool, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
