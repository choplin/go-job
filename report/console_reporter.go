package report

import (
	"os"
)

type consoleReporter struct {
	stringReporter
}

func newConsoleReporter(commandId string, commandName string) *consoleReporter {
	return &consoleReporter{
		stringReporter{commandId, commandName, os.Stdout, os.Stdout, os.Stderr},
	}
}

func (r *consoleReporter) startStdoutLogger(count int) {
	// do nothing
}
func (r *consoleReporter) finishStdoutLogger() {
	// do nothing
}
func (r *consoleReporter) startStderrLogger(count int) {
	// do nothing
}
func (r *consoleReporter) finishStderrLogger() {
	// do nothing
}
func (r *consoleReporter) close() {
	// do nothing
}
