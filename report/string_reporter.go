package report

import (
	"fmt"
	"io"
	"os/exec"
	"time"
)

type stringReporter struct {
	commandId   string
	commandName string
	fh          io.Writer
	out         io.Writer
	err         io.Writer
}

func (r *stringReporter) commandStart(startAt time.Time) {
	r.write(startAt, "The command has started\n")
}

func (r *stringReporter) commandSucceed(endAt time.Time, duration time.Duration) {
	r.write(endAt, "The command has finished with success in %f seconds.\n", duration.Seconds())
}

func (r *stringReporter) commandFail(endAt time.Time, duration time.Duration) {
	r.write(endAt, "The command has finished with failure in %f seconds.\n", duration.Seconds())
}

func (r *stringReporter) attemptStart(count int, pid int, startAt time.Time) {
	r.write(startAt, "The %s attempt has started. pid: %d\n", ordinalize(count), pid)
}

func (r *stringReporter) attemptSucceed(count int, endAt time.Time, duration time.Duration) {
	r.write(endAt, "The %s attempt has finished with success in %f seconds.\n", ordinalize(count), duration.Seconds())
}

func (r *stringReporter) attemptFail(count int, err *exec.ExitError, endAt time.Time, duration time.Duration) {
	r.write(endAt, "The %s attempt has failed in %f seconds.: %s.\n", ordinalize(count), duration.Seconds(), err)
}

func (r *stringReporter) attemptTimeout(count int, endAt time.Time, duration time.Duration) {
	r.write(endAt, "The %s attempt has been killed due to timeout. %f seconds has beed exceeded.\n", ordinalize(count), duration.Seconds())
}

func (r *stringReporter) attemptUnknownError(count int, err error, endAt time.Time) {
	r.write(endAt, "The %s attempt has failed with unknown error.: %s.\n", ordinalize(count), err)
}

func (r *stringReporter) stdoutLog(log string) {
	if r.out != nil {
		fmt.Fprint(r.out, log)
	}
}

func (r *stringReporter) stderrLog(log string) {
	if r.err != nil {
		fmt.Fprint(r.err, log)
	}
}

func ordinalize(count int) string {
	var ret string
	var str = fmt.Sprintf("%d", count)
	switch str[0:1] {
	case "1":
		ret = str + "st"
	case "2":
		ret = str + "nd"
	case "3":
		ret = str + "rd"
	default:
		ret = str + "th"
	}
	return ret
}

func (r *stringReporter) write(tm time.Time, format string, a ...interface{}) {
	format = tm.String() + " [" + r.commandName + "](" + r.commandId + ") " + format
	fmt.Fprintf(r.fh, format, a...)
}
