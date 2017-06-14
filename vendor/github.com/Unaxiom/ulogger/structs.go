package ulogger

import (
	"sync"

	commonStructs "github.com/Unaxiom/ulogger/structs"

	"github.com/fatih/color"
)

// timeFormat describes the output timestamp format
var timeFormat = "02-01-06 03:04:05"

type logMessage struct {
	commonStructs.LogMessage
}

type postMessage struct {
	commonStructs.PostMessage
}

// RemoteURL is the location where the log message is sent to
var RemoteURL = ""

// var RemoteURL = "http://localhost:13000/newlog"

// DebugLevel denotes a shorthand notation for setting log level to 'debug'
const DebugLevel = "debug"

// InfoLevel denotes a shorthand notation for setting log level to 'info'
const InfoLevel = "info"

// WarningLevel denotes a shorthand notation for setting log level to 'warning'
const WarningLevel = "warning"

// ErrorLevel denotes a shorthand notation for setting log level to 'error'
const ErrorLevel = "error"

// FatalLevel denotes a shorthand notation for setting log level to 'fatal'
const FatalLevel = "fatal"

// Logger is the main logging object
type Logger struct {
	OrganizationName string `json:"organization_name"`
	ApplicationName  string `json:"application_name"`
	RemoteAvailable  bool   // Stores if the struct needs to be pushed to the remote URL
	LineNumber       bool   // Stores if the file, function, and the line number need to be printed with every log

	LogLevel string // Stores the log level; values are debug, info, warning, error and fatal
	// debug --> 1
	// info --> 2
	// warning --> 3
	// error --> 4
	// fatal --> 5
	logLevelCode int // Stores the level in integer --> useful while checking if the log statement needs to be printed

	fieldsToDisplay []DisplayField
	// Customizable colors
	// Info colors
	InfoColor            *color.Color // Color of the info message
	InfoTimeColor        *color.Color // Color of the info timestamp
	InfoMessageTypeColor *color.Color // Color of the message type

	// Debug colors
	DebugColor            *color.Color // Color of the debug message
	DebugTimeColor        *color.Color // Color of the debug timestamp
	DebugMessageTypeColor *color.Color // Color of the message type

	// Warning colors
	WarningColor            *color.Color // Color of the warning message
	WarningTimeColor        *color.Color // Color of the warning timestamp
	WarningMessageTypeColor *color.Color // Color of the message type

	// Error colors
	ErrorColor            *color.Color // Color of the error message
	ErrorTimeColor        *color.Color // Color of the error timestamp
	ErrorMessageTypeColor *color.Color // Color of the message type

	// fatal colors
	FatalColor            *color.Color // Color of the fatal message
	FatalTimeColor        *color.Color // Color of the fatal timestamp
	FatalMessageTypeColor *color.Color // Color of the message type
}

// DisplayField stores the name and the value of the field that needs to be printed along with the log message
type DisplayField struct {
	Name  string
	Value interface{}
}

// runtimeParams stores the file name, the function name, and the line number of the caller function
type runtimeParams struct {
	file     string
	function string
	line     int64
}

var debugMutex sync.Mutex

type debugLogStruct struct {
	logList []logMessage
}

var debugLogs debugLogStruct

// addLog adds the log statement to the logList
func (list *debugLogStruct) addLog(log logMessage) {
	debugMutex.Lock()
	list.logList = append(list.logList, log)
	if len(list.logList) >= 5 {
		listToSend := list.logList[:]
		go postLogMessageToServer(listToSend)
		list.logList = []logMessage{}
	}
	debugMutex.Unlock()
}

var infoMutex sync.Mutex

type infoLogStruct struct {
	logList []logMessage
}

var infoLogs infoLogStruct

// addLog adds the log statement to the logList
func (list *infoLogStruct) addLog(log logMessage) {
	infoMutex.Lock()
	list.logList = append(list.logList, log)
	if len(list.logList) >= 5 {
		listToSend := list.logList[:]
		go postLogMessageToServer(listToSend)
		list.logList = []logMessage{}
	}
	infoMutex.Unlock()
}
