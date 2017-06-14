# ulogger

A custom logger that can be configured with different output colors. It can also stream logs to a remote URL. Take a look at the [documentation](https://godoc.org/github.com/Unaxiom/ulogger). It can also deep log an element using [spew](https://github.com/davecgh/go-spew).

# Usage
```
import "github.com/Unaxiom/ulogger"

func main() {
    log := ulogger.New()
    log.SetLogLevel(ulogger.DebugLevel) // Possible values are ulogger.DebugLevel (debug - least), ulogger.InfoLevel (info), ulogger.WarningLevel (warning), ulogger.ErrorLevel (error), ulogger.FatalLevel (fatal - highest), in ascending order, with ulogger.InfoLevel being the default
    log.RemoteAvailable = true // Defines whether logs need to the be streamed to the remote URL
    log.ApplicationName = "Temp Debugger" // Sets the applicaton name
    log.OrganizationName = "New org" // Sets the organization name that this build is licensed to
    log.LineNumber = false // Bool value sets if line number, function name, and the file name need to be printed to the console. Default is true
    ulogger.RemoteURL = "https://example.com" // Sets the remote URL where the log message needs to be sent via a POST request. If this is not set, and if log.RemoteAvailability is true, then the default URL is ""
}
```

# Possible log levels
```
# Debug level logs
log.Debug(args ...interface{})
log.Debugf(format string, args ...interface{})
log.Debugln(args ...interface{})

# Info level logs
log.Info(args ...interface{})
log.Infof(format string, args ...interface{})
log.Infoln(args ...interface{})

# Warning level logs
log.Warning(args ...interface{})
log.Warningf(format string, args ...interface{})
log.Warningln(args ...interface{})

# Error level logs
log.Error(args ...interface{})
log.Errorf(format string, args ...interface{})
log.Errorln(args ...interface{})

# Fatal level logs
log.Fatal(args ...interface{})
log.Fatalf(format string, args ...interface{})
log.Fatalln(args ...interface{})
```

# Deep logging
```
log.DebugDump(args ...interface{})
log.InfoDump(args ...interface{})
log.WarningDump(args ...interface{})
log.ErrorDump(args ...interface{})
log.FatalDump(args ...interface{})
```

# Modification of output colors
```
log.InfoColor
log.InfoTimeColor
log.InfoMessageTypeColor
log.DebugColor
log.DebugTimeColor
log.DebugMessageTypeColor
log.WarningColor
log.WarningTimeColor
log.WarningMessageTypeColor
log.ErrorColor
log.ErrorTimeColor
log.ErrorMessageTypeColor
log.FatalColor
log.FatalTimeColor
log.FatalMessageTypeColor
```
Each of these values could be assigned an color from the package [`github.com/fatih/color`](https://github.com/fatih/color).

# Disabling remote logging
Remote logging can be disabled by setting `log.RemoteAvailable` to `false`.

# Print function name, line number, and the file name
This can be done by setting `log.LineNumber = true`. The default instance when creating a new logger has it set to `true`. If these values aren't required, they can be hidden by setting this value to `false`.

# Send log messages to custom URL
If `log.RemoteAvailable` is set to `true`, then log messages are sent via a `POST` request to the URL. In case the URL needs to be changed, then it can be done so by updating `ulogger.RemoteURL` to the appropriate URL string. These messages would be sent via goroutines, so the execution will not be blocked.