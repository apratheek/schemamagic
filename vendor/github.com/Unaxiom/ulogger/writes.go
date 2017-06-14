package ulogger

import (
	"bytes"
	"fmt"
	"strings"

	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/fatih/color"
)

type prefixerSignature func(log *Logger, rtParams runtimeParams) (*bytes.Buffer, logMessage)

// write is applicable for all simple logs
func write(prefixFunc prefixerSignature, log *Logger, clr *color.Color, rtParams runtimeParams, ch chan int, args ...interface{}) {
	// Create the log that needs to be displayed on stdout
	buf, logStruct := prefixFunc(log, rtParams)
	clr.Fprint(buf, args...)
	clr.Print(buf.String())
	go func() {
		// Create the log message that needs to be sent to the server, only if the RemoteAvailable flag is set
		if !log.RemoteAvailable {
			ch <- 1
			if logStruct.MessageType == "FATAL" {
				os.Exit(1)
			}
			return
		}
		go sendLogMessageFromWrite(logStruct, ch, args...)
	}()
}

// sendLogMessageFromWrite sends log message to server from write()
func sendLogMessageFromWrite(logStruct logMessage, ch chan int, args ...interface{}) {
	buf := new(bytes.Buffer)
	fmt.Fprint(buf, args...)
	logStruct.MessageContent = strings.TrimSpace(buf.String())
	go func() {
		// Send the actual message here
		pushLogMessageToQueue(logStruct)
		ch <- 1
		if logStruct.MessageType == "FATAL" {
			os.Exit(1)
		}
	}()
}

// writef is applicable for all logs that need to be formatted
func writef(prefixFunc prefixerSignature, log *Logger, clr *color.Color, rtParams runtimeParams, ch chan int, format string, args ...interface{}) {
	// Create the log that needs to be displayed on stdout
	buf, logStruct := prefixFunc(log, rtParams)
	clr.Fprintf(buf, format, args...)
	clr.Print(buf.String())
	go func() {
		// Create the log message that needs to be sent to the server, only if the RemoteAvailable flag is set
		if !log.RemoteAvailable {
			ch <- 1
			if logStruct.MessageType == "FATAL" {
				os.Exit(1)
			}
			return
		}
		go sendLogMessageFromWritef(logStruct, ch, format, args...)
	}()
}

// writeDump is applicable for all simple logs
func writeDump(prefixFunc prefixerSignature, log *Logger, clr *color.Color, rtParams runtimeParams, ch chan int, args ...interface{}) {
	// Create the log that needs to be displayed on stdout
	buf, _ := prefixFunc(log, rtParams)
	// clr.Fprint(buf, args...)
	spew.Fdump(buf, args...)
	clr.Print(buf.String())
	ch <- 1
}

// sendLogMessageFromWritef sends log message to server from writef()
func sendLogMessageFromWritef(logStruct logMessage, ch chan int, format string, args ...interface{}) {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, format, args...)
	logStruct.MessageContent = strings.TrimSpace(buf.String())
	go func() {
		// Send the actual message here
		pushLogMessageToQueue(logStruct)
		ch <- 1
		if logStruct.MessageType == "FATAL" {
			os.Exit(1)
		}
	}()
}

// writeln is applicable for all logs ending with 'ln'
func writeln(prefixFunc prefixerSignature, log *Logger, clr *color.Color, rtParams runtimeParams, ch chan int, args ...interface{}) {
	// Create the log that needs to be displayed on stdout
	buf, logStruct := prefixFunc(log, rtParams)
	clr.Fprint(buf, args...)
	clr.Println(buf.String())
	go func() {
		// Create the log message that needs to be sent to the server, only if the RemoteAvailable flag is set
		if !log.RemoteAvailable {
			ch <- 1
			if logStruct.MessageType == "FATAL" {
				os.Exit(1)
			}
			return
		}
		go sendLogMessageFromWriteln(logStruct, ch, args...)
	}()
}

// sendLogMessageFromWriteln sends log message to server from writeln()
func sendLogMessageFromWriteln(logStruct logMessage, ch chan int, args ...interface{}) {
	buf := new(bytes.Buffer)
	fmt.Fprintln(buf, args...)
	logStruct.MessageContent = strings.TrimSpace(buf.String())
	go func() {
		// Send the actual message here
		pushLogMessageToQueue(logStruct)
		ch <- 1
		if logStruct.MessageType == "FATAL" {
			os.Exit(1)
		}
	}()
}
