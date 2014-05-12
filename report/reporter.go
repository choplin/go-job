package report

import (
	"os/exec"
	"time"
)

type reporter interface {
	commandStart(startAt time.Time)
	commandSucceed(endAt time.Time, duration time.Duration)
	commandFail(endAt time.Time, duration time.Duration)

	attemptStart(count int, pid int, startAt time.Time)
	attemptSucceed(count int, startAt time.Time, duration time.Duration)
	attemptFail(count int, err *exec.ExitError, endAt time.Time, duration time.Duration)
	attemptTimeout(count int, endAt time.Time, duration time.Duration)
	attemptUnknownError(count int, err error, endAt time.Time)

	startStdoutLogger(count int)
	finishStdoutLogger()
	stdoutLog(log string)

	startStderrLogger(count int)
	finishStderrLogger()
	stderrLog(log string)

	close()
}
