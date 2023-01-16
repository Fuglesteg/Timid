package verboseLog

import "log"

var Verbosity int = 6

// Log result if verbosity level high enough
func Vlogf(level int, format string, v ...interface{}) {
	if level <= Verbosity {
		log.Printf(format, v...)
	}
}

// Handle errors
func Checkreport(level int, err error) bool {
	if err == nil {
		return false
	}
	Vlogf(level, "Error: %s", err.Error())
	return true
}
