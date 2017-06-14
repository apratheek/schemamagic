package ulogger

import (
	"bytes"
	"fmt"
)

// Fatal displays a message and crashes the program
func (log *Logger) Fatal(args ...interface{}) {
	var rtParams runtimeParams
	if log.LineNumber {
		rtParams.file, rtParams.function, rtParams.line = fetchLocation()
	}
	if log.logLevelCode > 5 {
		if log.RemoteAvailable {
			// Create the logMessage struct here
			logStruct, _ := generateTimestamp("FATAL", rtParams)
			ch := make(chan int)
			go sendLogMessageFromWrite(logStruct, ch, args...)
			<-ch
		}
		return
	}
	ch := make(chan int)
	go func(ch chan int) {
		write(fatalPrefix, log, log.FatalColor, rtParams, ch, args...)
	}(ch)
	<-ch
}

// Fatalf displays a message and crashes the program
func (log *Logger) Fatalf(format string, args ...interface{}) {
	var rtParams runtimeParams
	if log.LineNumber {
		rtParams.file, rtParams.function, rtParams.line = fetchLocation()
	}
	if log.logLevelCode > 5 {
		if log.RemoteAvailable {
			// Create the logMessage struct here
			logStruct, _ := generateTimestamp("FATAL", rtParams)
			ch := make(chan int)
			go sendLogMessageFromWritef(logStruct, ch, format, args...)
			<-ch
		}
		return
	}
	ch := make(chan int)
	go func(ch chan int) {
		writef(fatalPrefix, log, log.FatalColor, rtParams, ch, format, args...)
	}(ch)
	<-ch
}

// Fatalln displays a message and crashes the program
func (log *Logger) Fatalln(args ...interface{}) {
	var rtParams runtimeParams
	if log.LineNumber {
		rtParams.file, rtParams.function, rtParams.line = fetchLocation()
	}
	if log.logLevelCode > 5 {
		if log.RemoteAvailable {
			// Create the logMessage struct here
			logStruct, _ := generateTimestamp("FATAL", rtParams)
			ch := make(chan int)
			go sendLogMessageFromWriteln(logStruct, ch, args...)
			<-ch
		}
		return
	}
	ch := make(chan int)
	go func(ch chan int) {
		writeln(fatalPrefix, log, log.FatalColor, rtParams, ch, args...)
	}(ch)
	<-ch
}

// FatalDump displays the dump of the variables passed using the go-spew library
func (log *Logger) FatalDump(args ...interface{}) {
	var rtParams runtimeParams
	if log.LineNumber {
		rtParams.file, rtParams.function, rtParams.line = fetchLocation()
	}
	// Don't stream this to the remote server
	ch := make(chan int)
	go func(ch chan int) {
		writeDump(fatalPrefix, log, log.FatalColor, rtParams, ch, args...)
	}(ch)
	<-ch
}

// Returns a string, along with a logMessage after prefixing the timestamp and the type of log
func fatalPrefix(log *Logger, rtParams runtimeParams) (*bytes.Buffer, logMessage) {
	buf := new(bytes.Buffer)
	logStruct, timestamp := generateTimestamp("FATAL", rtParams)
	logStruct.OrganizationName = log.OrganizationName
	logStruct.ApplicationName = log.ApplicationName
	// Print the runtime parameters
	if log.LineNumber {
		log.FatalMessageTypeColor.Fprintf(buf, "%s->%s():%d", rtParams.file, rtParams.function, rtParams.line)
		fmt.Fprint(buf, " ")
	}
	log.FatalTimeColor.Fprint(buf, timestamp.Format(timeFormat))
	fmt.Fprint(buf, " ")
	log.FatalMessageTypeColor.Fprint(buf, logStruct.MessageType)
	fmt.Fprint(buf, " ")
	return buf, logStruct
}
