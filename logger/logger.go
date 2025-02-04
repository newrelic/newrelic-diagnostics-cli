package logger

import (
	"fmt"
	"log"
	"time"

	"github.com/newrelic/newrelic-diagnostics-cli/config"
)

// API - an empty struct to hang the logger methods off of
type API interface {
	FixedPrefix(int, string, string)
	Info(...interface{})
	Infof(string, ...interface{})
	Debug(...interface{})
	Debugf(string, ...interface{})
	Dump(...interface{})
}

type packageMethods struct{}

// Log is a struct that exposes all the public functions of the logger package
var Log = packageMethods{}

// FixedPrefix is for printing a line with a fixed width prefix followed by other text
func FixedPrefix(length int, prefix string, text string) {
	format := fmt.Sprintf("%%-%ds%%s\n", length) //produces something like "%-10s%s\n"
	fmt.Printf(format, prefix, text)
}

// FixedPrefix - alias via an empty struct to the original implementation
func (l packageMethods) FixedPrefix(length int, prefix string, text string) {
	FixedPrefix(length, prefix, text)
}

// Info is wrapper for Println. No verbosity check -- it always logs.
func Info(s ...interface{}) {
	fmt.Println(s...)
}

// Info - alias via an empty struct to the original implementation
func (l packageMethods) Info(s ...interface{}) {
	Info(s...)
}

// Infof is wrapper for Printf. No verbosity check -- it always logs.
func Infof(format string, s ...interface{}) {
	fmt.Printf(format, s...)
}

// Fatal is wrapper for log.Fatal(). Prints message followed by a call to os.Exit(1).
func Fatal(s ...interface{}) {
	// log.Fatal("Exception occurred!") or log.Fatal(err)
	log.Fatal(s...)
}

// Fatalf is wrapper for log.Fatalf. No verbosity check --it always logs followed by a call to os.Exit(1).
func Fatalf(format string, s ...interface{}) {
	log.Fatalf(format, s...)
}

// Infof - alias via an empty struct to the original implementation
func (l packageMethods) Infof(format string, s ...interface{}) {
	Infof(format, s...)
}

// Dump is wrapper for Printf with a preset formatting to display variable types. No verbosity check -- it always logs.
func Dump(s ...interface{}) {
	for _, d := range s {
		fmt.Printf("%#v", d)
	}
}

// Dump - alias via an empty struct to the original implementation
func (l packageMethods) Dump(s ...interface{}) {
	Dump(s)
}

// Debug is wrapper for Println. Only logs if LogLevel is set to Debug verbosity
func Debug(s ...interface{}) {
	if config.LogLevel == 1 {
		// This adds the timestamp and DEBUG statement to the log message
		// The single line format defines a slice of empty interfaces defined a new interface with the timestamp and [DEBUG] string as the first element,
		// then expands s into a single interface to add to the slice of interfaces and then finally expands the inline slice to a single interface :)
		fmt.Println(append([]interface{}{interface{}(getTimestamp() + " [DEBUG]")}, s...)...)
	}
}

// Debug - alias via an empty struct to the original implementation
func (l packageMethods) Debug(s ...interface{}) {
	Debug(s...)
}

// Debugf is wrapper for Printf. Only logs if LogLevel is set to Debug verbosity
func Debugf(format string, s ...interface{}) {
	if config.LogLevel == 1 {
		fmt.Printf(getTimestamp()+" [DEBUG] "+format, s...)
	}
}

// Debugf - alias via an empty struct to the original implementation
func (l packageMethods) Debugf(format string, s ...interface{}) {
	Debugf(format, s...)
}

// getTimestamp provides current time in UTC
func getTimestamp() string {

	//time constants for formatting, found here: https://golang.org/pkg/time/#pkg-constants
	return time.Now().UTC().Format(time.RFC3339)
}
