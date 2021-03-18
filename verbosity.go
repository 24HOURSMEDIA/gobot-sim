package gobot_sim

import "fmt"

const (
	VERBOSITY_QUIET = iota
	// log info and errors
	VERBOSITY_V
	// log info, errors and warnings
	VERBOSITY_VV
	// log info, errors, warnings and debug
	VERBOSITY_VVV
)

type VerbosityLogger struct {
	Verbosity int
	Prefix    string
}

func (logger VerbosityLogger) Debug(format string, a ...interface{}) {
	if logger.Verbosity >= VERBOSITY_VVV {
		fmt.Println("GOBOTSIM[debug]:	" + fmt.Sprintf(format, a...))
	}
}
func (logger VerbosityLogger) Warning(format string, a ...interface{}) {
	if logger.Verbosity >= VERBOSITY_VV {
		fmt.Println("GOBOTSIM[warning]:	" + fmt.Sprintf(format, a...))
	}
}
func (logger VerbosityLogger) Info(format string, a ...interface{}) {
	if logger.Verbosity >= VERBOSITY_V {
		fmt.Println("GOBOTSIM[info]:	" + fmt.Sprintf(format, a...))
	}
}
func (logger VerbosityLogger) Error(format string, a ...interface{}) {
	if logger.Verbosity >= VERBOSITY_V {
		fmt.Println("GOBOTSIM[error]:	" + fmt.Sprintf(format, a...))
	}
}
