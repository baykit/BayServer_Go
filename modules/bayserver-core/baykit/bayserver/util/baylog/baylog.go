package baylog

import (
	"bayserver-core/baykit/bayserver/util/exception"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const LOG_LEVEL_TRACE int = 0
const LOG_LEVEL_DEBUG int = 1
const LOG_LEVEL_INFO int = 2
const LOG_LEVEL_WARN int = 3
const LOG_LEVEL_ERROR int = 4
const LOG_LEVEL_FATAL int = 5

/** Log level */
var logLevel = LOG_LEVEL_INFO

var logLevelNames = []string{"TRACE", "DEBUG", "INFO ", "WARN ", "ERROR", "FATAL"}

func SetLogLevel(s string) {
	for i, arg := range logLevelNames {
		if strings.ToUpper(s) == arg {
			logLevel = i
			return
		}
	}
	Warn("Unknown LogLevel", s)
}

func IsDebugMode() bool {
	return logLevel <= LOG_LEVEL_DEBUG
}

func IsTraceMode() bool {
	return logLevel == LOG_LEVEL_TRACE
}

func Trace(fmt string, args ...interface{}) {
	log(LOG_LEVEL_TRACE, 3, nil, fmt, args...)
}

func TraceE(err error, fmt string, args ...interface{}) {
	log(LOG_LEVEL_TRACE, 3, err, fmt, args...)
}

func Debug(fmt string, args ...interface{}) {
	log(LOG_LEVEL_DEBUG, 3, nil, fmt, args...)
}

func DebugE(err error, fmt string, args ...interface{}) {
	log(LOG_LEVEL_DEBUG, 3, err, fmt, args...)
}

func Info(fmt string, args ...interface{}) {
	log(LOG_LEVEL_INFO, 3, nil, fmt, args...)
}

func Warn(fmt string, args ...interface{}) {
	log(LOG_LEVEL_WARN, 3, nil, fmt, args...)
}

func WarnE(err error, fmt string, args ...interface{}) {
	log(LOG_LEVEL_WARN, 3, err, fmt, args...)
}

func Error(fmt string, args ...interface{}) {
	log(LOG_LEVEL_ERROR, 3, nil, fmt, args...)
}

func ErrorE(err error, fmt string, args ...interface{}) {
	log(LOG_LEVEL_ERROR, 3, err, fmt, args...)
}

func Fatal(fmt string, args ...interface{}) {
	log(LOG_LEVEL_FATAL, 3, nil, fmt, args...)
}

func FatalE(err error, fmt string, args ...interface{}) {
	log(LOG_LEVEL_FATAL, 3, err, fmt, args...)
}

func log(lvl int, stackIdx int, err error, format string, args ...interface{}) {

	if lvl < logLevel {
		return
	}

	pos := getPos()

	dateStr := time.Now().Format("2006-01-02 15:04:05 MST")

	var msg string
	if len(args) == 0 {
		msg = fmt.Sprintf("%s", format)
	} else {
		msg = fmt.Sprintf(format, args...)
	}
	fmt.Printf("[%s] %s. %s (%s)\n", dateStr, logLevelNames[lvl], msg, pos)

	if err != nil {
		fmt.Printf("%s\n", err.Error())
		if IsDebugMode() || lvl == LOG_LEVEL_FATAL {
			if err, ok := err.(exception.Exception); ok {
				stackTraceString := ""
				for {
					frame, more := err.Frames().Next()
					stackTraceString += frame.Function + "\n\t" + frame.File + ":" + strconv.Itoa(frame.Line) + "\n"
					if !more {
						break
					}
				}
				fmt.Println(stackTraceString)
			}
		}
	}
}

func getPos() string {
	buf := make([]uintptr, 10)
	n := runtime.Callers(1, buf[:])
	frames := runtime.CallersFrames(buf[:n])
	var frame runtime.Frame
	for i := 0; i < 4; i++ {
		frame, _ = frames.Next()
	}
	return fmt.Sprintf("%s:%d", frame.File, frame.Line)
}
