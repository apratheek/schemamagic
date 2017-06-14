package ulogger

import (
	"bytes"
	"fmt"
)

// Warning displays a warning message
func (log *Logger) Warning(args ...interface{}) {
	var rtParams runtimeParams
	if log.LineNumber {
		rtParams.file, rtParams.function, rtParams.line = fetchLocation()
	}
	if log.logLevelCode > 3 {
		if log.RemoteAvailable {
			// Create the logMessage struct here
			logStruct, _ := generateTimestamp("WARNING", rtParams)
			ch := make(chan int)
			go sendLogMessageFromWrite(logStruct, ch, args...)
			<-ch
		}
		return
	}
	ch := make(chan int)
	go func(ch chan int) {
		write(warningPrefix, log, log.WarningColor, rtParams, ch, args...)
	}(ch)
	<-ch
}

// Warningf displays a warning message
func (log *Logger) Warningf(format string, args ...interface{}) {
	var rtParams runtimeParams
	if log.LineNumber {
		rtParams.file, rtParams.function, rtParams.line = fetchLocation()
	}
	if log.logLevelCode > 3 {
		if log.RemoteAvailable {
			// Create the logMessage struct here
			logStruct, _ := generateTimestamp("WARNING", rtParams)
			ch := make(chan int)
			go sendLogMessageFromWritef(logStruct, ch, format, args...)
			<-ch
		}
		return
	}
	ch := make(chan int)
	go func(ch chan int) {
		writef(warningPrefix, log, log.WarningColor, rtParams, ch, format, args...)
	}(ch)
	<-ch
}

// Warningln displays a warning message
func (log *Logger) Warningln(args ...interface{}) {
	var rtParams runtimeParams
	if log.LineNumber {
		rtParams.file, rtParams.function, rtParams.line = fetchLocation()
	}
	if log.logLevelCode > 3 {
		if log.RemoteAvailable {
			// Create the logMessage struct here
			logStruct, _ := generateTimestamp("WARNING", rtParams)
			ch := make(chan int)
			go sendLogMessageFromWriteln(logStruct, ch, args...)
			<-ch
		}
		return
	}
	ch := make(chan int)
	go func(ch chan int) {
		writeln(warningPrefix, log, log.WarningColor, rtParams, ch, args...)
	}(ch)
	<-ch
}

// WarningDump displays the dump of the variables passed using the go-spew library
func (log *Logger) WarningDump(args ...interface{}) {
	var rtParams runtimeParams
	if log.LineNumber {
		rtParams.file, rtParams.function, rtParams.line = fetchLocation()
	}
	// Don't stream this to the remote server
	ch := make(chan int)
	go func(ch chan int) {
		writeDump(warningPrefix, log, log.WarningColor, rtParams, ch, args...)
	}(ch)
	<-ch
}

// Returns a string, along with a logMessage after prefixing the timestamp and the type of log
func warningPrefix(log *Logger, rtParams runtimeParams) (*bytes.Buffer, logMessage) {
	buf := new(bytes.Buffer)
	logStruct, timestamp := generateTimestamp("WARNING", rtParams)
	logStruct.OrganizationName = log.OrganizationName
	logStruct.ApplicationName = log.ApplicationName
	// Print the runtime parameters
	if log.LineNumber {
		log.WarningMessageTypeColor.Fprintf(buf, "%s->%s():%d", rtParams.file, rtParams.function, rtParams.line)
		fmt.Fprint(buf, " ")
	}
	log.WarningTimeColor.Fprint(buf, timestamp.Format(timeFormat))
	fmt.Fprint(buf, " ")
	log.WarningMessageTypeColor.Fprint(buf, logStruct.MessageType)
	fmt.Fprint(buf, " ")
	return buf, logStruct
}
