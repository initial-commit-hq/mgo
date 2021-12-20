// mgo - MongoDB driver for Go

package mgo

import (
	"fmt"
	"sync"
)

// ---------------------------------------------------------------------------
// Logging integration.

// LogLogger avoid importing the log type information unnecessarily.  There's a small cost
// associated with using an interface rather than the type.  Depending on how
// often the logger is plugged in, it would be worth using the type instead.
type logLogger interface {
	Output(calldepth int, s string) error
}

var (
	globalLogger logLogger
	globalDebug  bool
	globalMutex  sync.Mutex
)

// RACE WARNING: There are known data races when logging, which are manually
// silenced when the race detector is in use. These data races won't be
// observed in typical use, because logging is supposed to be set up once when
// the application starts. Having raceDetector as a constant, the compiler
// should elide the locks altogether in actual use.

// SetLogger specify the *log.Logger object where log messages should be sent to.
func SetLogger(logger logLogger) {
	if raceDetector {
		globalMutex.Lock()
		defer globalMutex.Unlock()
	}
	globalLogger = logger
}

// SetDebug enable the delivery of debug messages to the logger.  Only meaningful
// if a logger is also set.
func SetDebug(debug bool) {
	if raceDetector {
		globalMutex.Lock()
		defer globalMutex.Unlock()
	}
	globalDebug = debug
}

func log(v ...interface{}) {
	if raceDetector {
		globalMutex.Lock()
		defer globalMutex.Unlock()
	}
	if globalLogger != nil {
		globalLogger.Output(2, fmt.Sprint(v...))
	}
}

func logln(v ...interface{}) {
	if raceDetector {
		globalMutex.Lock()
		defer globalMutex.Unlock()
	}
	if globalLogger != nil {
		globalLogger.Output(2, fmt.Sprintln(v...))
	}
}

func logf(format string, v ...interface{}) {
	if raceDetector {
		globalMutex.Lock()
		defer globalMutex.Unlock()
	}
	if globalLogger != nil {
		globalLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

func debug(v ...interface{}) {
	if raceDetector {
		globalMutex.Lock()
		defer globalMutex.Unlock()
	}
	if globalDebug && globalLogger != nil {
		globalLogger.Output(2, fmt.Sprint(v...))
	}
}

func debugln(v ...interface{}) {
	if raceDetector {
		globalMutex.Lock()
		defer globalMutex.Unlock()
	}
	if globalDebug && globalLogger != nil {
		globalLogger.Output(2, fmt.Sprintln(v...))
	}
}

func debugf(format string, v ...interface{}) {
	if raceDetector {
		globalMutex.Lock()
		defer globalMutex.Unlock()
	}
	if globalDebug && globalLogger != nil {
		globalLogger.Output(2, fmt.Sprintf(format, v...))
	}
}
