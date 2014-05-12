package report

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type ReporterList []reporter

func NewReporterList(commandId, commandName string, config *ReporterConfig) (ReporterList, error) {
	list := make([]reporter, 0)

	for _, s := range strings.Split(config.Reporters, ",") {
		switch s {
		case "console":
			list = append(list, newConsoleReporter(commandId, commandName))
		case "fluentd":
			r, err := newFluentdReporter(commandId, commandName, config.FluentdHost, config.FluentdPort, config.FluentdTagPrefix)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to initialize fluentd reporter. %s\n", err)
			} else {
				list = append(list, r)
			}
		case "file":
			r, err := newFileReporter(commandId, commandName, config.FileDirectory)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to initialize file reporter. %s\n", err)
			} else {
				list = append(list, r)
			}
		default:
			return list, fmt.Errorf("unknown reporter option: %s", s)
		}
	}
	return list, nil
}

func (list *ReporterList) CommandStart(startAt time.Time) {
	f := func(r reporter) {
		r.commandStart(startAt)
	}
	list.doForEachReporter(f)
}

func (list *ReporterList) CommandSucceed(endAt time.Time, duration time.Duration) {
	f := func(r reporter) {
		r.commandSucceed(endAt, duration)
	}
	list.doForEachReporter(f)
}

func (list *ReporterList) CommandFail(endAt time.Time, duration time.Duration) {
	f := func(r reporter) {
		r.commandFail(endAt, duration)
	}
	list.doForEachReporter(f)
}

func (list *ReporterList) AttemptStart(count int, pid int, startAt time.Time) {
	f := func(r reporter) {
		r.attemptStart(count, pid, startAt)
	}
	list.doForEachReporter(f)
}

func (list *ReporterList) AttemptSucceed(count int, endAt time.Time, duration time.Duration) {
	f := func(r reporter) {
		r.attemptSucceed(count, endAt, duration)
	}
	list.doForEachReporter(f)
}

func (list *ReporterList) AttemptFail(count int, err *exec.ExitError, endAt time.Time, duration time.Duration) {
	f := func(r reporter) {
		r.attemptFail(count, err, endAt, duration)
	}
	list.doForEachReporter(f)
}

func (list *ReporterList) AttemptTimeout(count int, endAt time.Time, duration time.Duration) {
	f := func(r reporter) {
		r.attemptTimeout(count, endAt, duration)
	}
	list.doForEachReporter(f)
}

func (list *ReporterList) AttemptUnknownError(count int, err error, endAt time.Time) {
	f := func(r reporter) {
		r.attemptUnknownError(count, err, endAt)
	}
	list.doForEachReporter(f)
}

func (list *ReporterList) StartStdoutLogger(count int) {
	f := func(r reporter) {
		r.startStdoutLogger(count)
	}
	list.doForEachReporter(f)
}

func (list *ReporterList) FinishStdoutLogger() {
	f := func(r reporter) {
		r.finishStdoutLogger()
	}
	list.doForEachReporter(f)
}

func (list *ReporterList) StdoutLog(log string) {
	f := func(r reporter) {
		r.stdoutLog(log)
	}
	list.doForEachReporter(f)
}

func (list *ReporterList) StartStderrLogger(count int) {
	f := func(r reporter) {
		r.startStderrLogger(count)
	}
	list.doForEachReporter(f)
}

func (list *ReporterList) FinishStderrLogger() {
	f := func(r reporter) {
		r.finishStderrLogger()
	}
	list.doForEachReporter(f)
}

func (list *ReporterList) StderrLog(log string) {
	f := func(r reporter) {
		r.stderrLog(log)
	}
	list.doForEachReporter(f)
}

func (list *ReporterList) Close() {
	for _, r := range []reporter(*list) {
		r.close()
	}
}

func (list *ReporterList) doForEachReporter(f func(r reporter)) {
	var wg sync.WaitGroup

	for _, r := range []reporter(*list) {
		wg.Add(1)
		go func(r reporter) {
			defer wg.Done()
			f(r)
		}(r)
	}
	wg.Wait()
}
