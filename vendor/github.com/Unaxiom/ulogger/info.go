package ulogger

import (
	"bytes"
	"fmt"
)

// Info displays a non-fatal log message
func (log *Logger) Info(args ...interface{}) {
	var rtParams runtimeParams
	if log.LineNumber {
		rtParams.file, rtParams.function, rtParams.line = fetchLocation()
	}
	if log.logLevelCode > 2 {
		if log.RemoteAvailable {
			// Create the logMessage struct here
			logStruct, _ := generateTimestamp("INFO", rtParams)
			ch := make(chan int)
			go sendLogMessageFromWrite(logStruct, ch, args...)
			<-ch
		}
		return
	}
	ch := make(chan int)
	go func(ch chan int) {
		write(infoPrefix, log, log.InfoColor, rtParams, ch, args...)
	}(ch)
	<-ch
}

// Infof displays a non-fatal log message according to the format string
func (log *Logger) Infof(format string, args ...interface{}) {
	var rtParams runtimeParams
	if log.LineNumber {
		rtParams.file, rtParams.function, rtParams.line = fetchLocation()
	}
	if log.logLevelCode > 2 {
		if log.RemoteAvailable {
			// Create the logMessage struct here
			logStruct, _ := generateTimestamp("INFO", rtParams)
			ch := make(chan int)
			go sendLogMessageFromWritef(logStruct, ch, format, args...)
			<-ch
		}
		return
	}
	ch := make(chan int)
	go func(ch chan int) {
		writef(infoPrefix, log, log.InfoColor, rtParams, ch, format, args...)
	}(ch)
	<-ch
}

// Infoln displays a non-fatal log message
func (log *Logger) Infoln(args ...interface{}) {
	var rtParams runtimeParams
	if log.LineNumber {
		rtParams.file, rtParams.function, rtParams.line = fetchLocation()
	}
	if log.logLevelCode > 2 {
		if log.RemoteAvailable {
			// Create the logMessage struct here
			logStruct, _ := generateTimestamp("INFO", rtParams)
			ch := make(chan int)
			go sendLogMessageFromWriteln(logStruct, ch, args...)
			<-ch
		}
		return
	}
	ch := make(chan int)
	go func(ch chan int) {
		writeln(infoPrefix, log, log.InfoColor, rtParams, ch, args...)
	}(ch)
	<-ch
}

// InfoDump displays the dump of the variables passed using the go-spew library
func (log *Logger) InfoDump(args ...interface{}) {
	var rtParams runtimeParams
	if log.LineNumber {
		rtParams.file, rtParams.function, rtParams.line = fetchLocation()
	}
	// Don't stream this to the remote server
	ch := make(chan int)
	go func(ch chan int) {
		writeDump(infoPrefix, log, log.InfoColor, rtParams, ch, args...)
	}(ch)
	<-ch
}

// Returns a string, along with a logMessage after prefixing the timestamp and the type of log
func infoPrefix(log *Logger, rtParams runtimeParams) (*bytes.Buffer, logMessage) {
	buf := new(bytes.Buffer)
	logStruct, timestamp := generateTimestamp("INFO", rtParams)
	logStruct.OrganizationName = log.OrganizationName
	logStruct.ApplicationName = log.ApplicationName
	// Print the runtime parameters
	if log.LineNumber {
		log.InfoMessageTypeColor.Fprintf(buf, "%s->%s():%d", rtParams.file, rtParams.function, rtParams.line)
		fmt.Fprint(buf, " ")
	}
	log.InfoTimeColor.Fprint(buf, timestamp.Format(timeFormat))
	fmt.Fprint(buf, " ")
	log.InfoMessageTypeColor.Fprint(buf, logStruct.MessageType)
	fmt.Fprint(buf, " ")
	// Attach the display fields here
	// attachDisplayFields(buf, log.InfoColor, log.fieldsToDisplay)
	return buf, logStruct
}
