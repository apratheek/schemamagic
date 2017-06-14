package ulogger

import (
	"bytes"
	"fmt"
)

// Debug displays a debugging message useful in development environment
func (log *Logger) Debug(args ...interface{}) {
	var rtParams runtimeParams
	if log.LineNumber {
		rtParams.file, rtParams.function, rtParams.line = fetchLocation()
	}
	if log.logLevelCode > 1 {
		if log.RemoteAvailable {
			// Create the logMessage struct here
			logStruct, _ := generateTimestamp("DEBUG", rtParams)
			ch := make(chan int)
			go sendLogMessageFromWrite(logStruct, ch, args...)
			<-ch
		}
		return
	}
	ch := make(chan int)
	go func(ch chan int) {
		write(debugPrefix, log, log.DebugColor, rtParams, ch, args...)
	}(ch)
	<-ch
}

// Debugf displays a debugging message
func (log *Logger) Debugf(format string, args ...interface{}) {
	var rtParams runtimeParams
	if log.LineNumber {
		rtParams.file, rtParams.function, rtParams.line = fetchLocation()
	}
	if log.logLevelCode > 1 {
		if log.RemoteAvailable {
			// Create the logMessage struct here
			logStruct, _ := generateTimestamp("DEBUG", rtParams)
			ch := make(chan int)
			go sendLogMessageFromWritef(logStruct, ch, format, args...)
			<-ch
		}
		return
	}
	ch := make(chan int)
	go func(ch chan int) {
		writef(debugPrefix, log, log.DebugColor, rtParams, ch, format, args...)
	}(ch)
	<-ch
}

// Debugln displays a debugging message
func (log *Logger) Debugln(args ...interface{}) {
	var rtParams runtimeParams
	if log.LineNumber {
		rtParams.file, rtParams.function, rtParams.line = fetchLocation()
	}
	if log.logLevelCode > 1 {
		if log.RemoteAvailable {
			// Create the logMessage struct here
			logStruct, _ := generateTimestamp("DEBUG", rtParams)
			ch := make(chan int)
			go sendLogMessageFromWriteln(logStruct, ch, args...)
			<-ch
		}
		return
	}
	ch := make(chan int)
	go func(ch chan int) {
		writeln(debugPrefix, log, log.DebugColor, rtParams, ch, args...)
	}(ch)
	<-ch
}

// DebugDump displays the dump of the variables passed using the go-spew library
func (log *Logger) DebugDump(args ...interface{}) {
	// Don't stream this to the remote server
	var rtParams runtimeParams
	if log.LineNumber {
		rtParams.file, rtParams.function, rtParams.line = fetchLocation()
	}
	ch := make(chan int)
	go func(ch chan int) {
		writeDump(debugPrefix, log, log.DebugColor, rtParams, ch, args...)
	}(ch)
	<-ch
}

// Returns a string, along with a logMessage after prefixing the timestamp and the type of log
func debugPrefix(log *Logger, rtParams runtimeParams) (*bytes.Buffer, logMessage) {
	buf := new(bytes.Buffer)
	logStruct, timestamp := generateTimestamp("DEBUG", rtParams)
	logStruct.OrganizationName = log.OrganizationName
	logStruct.ApplicationName = log.ApplicationName
	// Print the runtime parameters
	if log.LineNumber {
		log.DebugMessageTypeColor.Fprintf(buf, "%s->%s():%d", rtParams.file, rtParams.function, rtParams.line)
		fmt.Fprint(buf, " ")
	}
	log.DebugTimeColor.Fprint(buf, timestamp.Format(timeFormat))
	fmt.Fprint(buf, " ")
	log.DebugMessageTypeColor.Fprint(buf, logStruct.MessageType)
	fmt.Fprint(buf, " ")
	return buf, logStruct
}
