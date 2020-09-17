// Copyright (C) MongoDB, Inc. 2014-present.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

// Package log provides a utility to log timestamped messages to an io.Writer.
package log

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Tool Logger verbosity constants
const (
	Error = iota
	Debug
	Trace
	Warn
	Info
)

const (
	ToolTimeFormat = "2006-01-02T15:04:05.000-0700"
)

var errAbbreviations = []string{"ERR", "DEB", "TRC", "WRN", "INF"}

//// Tool Logger Definition

type ToolLogger struct {
	mutex     *sync.Mutex
	writer    io.Writer
	format    string
	verbosity int
}

type VerbosityLevel interface {
	Level() int
	IsQuiet() bool
}

func (tl *ToolLogger) SetVerbosity(level VerbosityLevel) {
	if level == nil {
		tl.verbosity = 0
		return
	}

	if level.IsQuiet() {
		tl.verbosity = -1
	} else {
		tl.verbosity = level.Level()
	}
}

func (tl *ToolLogger) SetWriter(writer io.Writer) {
	tl.writer = writer
}

func (tl *ToolLogger) SetDateFormat(dateFormat string) {
	tl.format = dateFormat
}

func (tl *ToolLogger) Logvf(minVerb int, printLogType bool, format string, a ...interface{}) {
	if minVerb < 0 {
		panic("cannot set a minimum log verbosity that is less than 0")
	}

	logLevel := minVerb
	if minVerb == Error || minVerb == Info || minVerb == Warn {
		logLevel = 0
	}
	if logLevel <= tl.verbosity {
		tl.mutex.Lock()
		defer tl.mutex.Unlock()
		if printLogType {
			tl.log(fmt.Sprintf(errAbbreviations[minVerb] + " " + format, a...))
		} else {
			tl.log(fmt.Sprintf(format, a...))
		}
	}
}

func (tl *ToolLogger) Logv(minVerb int, printLogType bool, msg string) {
	if minVerb < 0 {
		panic("cannot set a minimum log verbosity that is less than 0")
	}

	logLevel := minVerb
	if minVerb == Error || minVerb == Info || minVerb == Warn {
		logLevel = 0
	}
	if logLevel <= tl.verbosity {
		tl.mutex.Lock()
		defer tl.mutex.Unlock()
		if printLogType {
			tl.log(errAbbreviations[minVerb] + " " + msg)
		} else {
			tl.log(msg)
		}

	}
}

func (tl *ToolLogger) log(msg string) {
	fmt.Fprintf(tl.writer, "%v\t%v\n", time.Now().Format(tl.format), msg)
}

func NewToolLogger(verbosity VerbosityLevel) *ToolLogger {
	tl := &ToolLogger{
		mutex:  &sync.Mutex{},
		writer: os.Stderr, // default to stderr
		format: ToolTimeFormat,
	}
	tl.SetVerbosity(verbosity)
	return tl
}

//// Log Writer Interface

// toolLogWriter is an io.Writer wrapping a tool logger. It is a private
// type meant for creation with the ToolLogger.Writer(...) method.
type toolLogWriter struct {
	logger       *ToolLogger
	minVerbosity int
}

func (tlw *toolLogWriter) Write(message []byte) (int, error) {
	tlw.logger.Logv(tlw.minVerbosity, false, string(message))
	return len(message), nil
}

// Writer returns an io.Writer that writes to the logger with
// the given verbosity
func (tl *ToolLogger) Writer(minVerb int) io.Writer {
	return &toolLogWriter{tl, minVerb}
}

//// Global Logging

var globalToolLogger *ToolLogger

func init() {
	if globalToolLogger == nil {
		// initialize tool logger with verbosity level = 0
		globalToolLogger = NewToolLogger(nil)
	}
}

// IsInVerbosity returns true if the current verbosity level setting is
// greater than or equal to the given level.
func IsInVerbosity(minVerb int) bool {
	return minVerb <= globalToolLogger.verbosity
}

func Logvf(minVerb int, printLogType bool, format string, a ...interface{}) {
	globalToolLogger.Logvf(minVerb, printLogType, format, a...)
}

func Logv(minVerb int, printLogType bool, msg string) {
	globalToolLogger.Logv(minVerb, printLogType, msg)
}

func SetVerbosity(verbosity VerbosityLevel) {
	globalToolLogger.SetVerbosity(verbosity)
}

func SetWriter(writer io.Writer) {
	globalToolLogger.SetWriter(writer)
}

func SetDateFormat(dateFormat string) {
	globalToolLogger.SetDateFormat(dateFormat)
}

func Writer(minVerb int) io.Writer {
	return globalToolLogger.Writer(minVerb)
}
