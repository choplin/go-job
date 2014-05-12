package report

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/t-k/fluent-logger-golang/fluent"
)

const (
	commandStartTag        = "command_start"
	commandSucceedTag      = "command_succeed"
	commandFailTag         = "command_fail"
	attemptStartTag        = "attempt_start"
	attemptSucceedTag      = "attempt_succeed"
	attemptFailTag         = "attempt_fail"
	attemptTimeoutTag      = "attempt_timeout"
	attemptUnknownErrorTag = "attempt_unknown_error"
	stdoutTag              = "stdout"
	stderrTag              = "stderr"
)

type fluentdReporter struct {
	commandId   string
	commandName string
	hostname    string
	tagPrefix   string
	logger      *fluent.Fluent
	stdout      chan string
	stderr      chan string

	// for notification with hipchat
	buf *bytes.Buffer
	sr  *stringReporter
}

func newFluentdReporter(commandId string, commandName string, host string, port int, tagPrefix string) (*fluentdReporter, error) {
	logger, err := fluent.New(fluent.Config{
		FluentHost: host,
		FluentPort: port,
	})

	if err != nil {
		return nil, err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	sr := &stringReporter{
		commandId:   commandId,
		commandName: commandName,
		fh:          buf,
	}

	return &fluentdReporter{
		commandId:   commandId,
		commandName: commandName,
		hostname:    hostname,
		tagPrefix:   tagPrefix,
		logger:      logger,
		buf:         buf,
		sr:          sr,
	}, nil
}

func (r *fluentdReporter) createRecord(rest map[string]interface{}) map[string]interface{} {
	ret := map[string]interface{}{
		"commandId":   r.commandId,
		"commandName": r.commandName,
		"hostname":    r.hostname,
	}

	for v, k := range rest {
		ret[v] = k
	}

	return ret
}

func (r *fluentdReporter) commandStart(startAt time.Time) {
	r.sr.commandStart(startAt)
	message := r.buf.String()
	r.buf.Reset()

	record := r.createRecord(map[string]interface{}{
		"message": message,
	})
	tag := makeTag(r.tagPrefix, commandStartTag)
	r.logger.PostWithTime(tag, startAt, record)
}

func (r *fluentdReporter) commandSucceed(endAt time.Time, duration time.Duration) {
	r.sr.commandSucceed(endAt, duration)
	message := r.buf.String()
	r.buf.Reset()

	record := r.createRecord(map[string]interface{}{
		"duration": duration.Seconds(),
		"message":  message,
	})
	tag := makeTag(r.tagPrefix, commandSucceedTag)
	r.logger.PostWithTime(tag, endAt, record)
}

func (r *fluentdReporter) commandFail(endAt time.Time, duration time.Duration) {
	r.sr.commandFail(endAt, duration)
	message := r.buf.String()
	r.buf.Reset()

	record := r.createRecord(map[string]interface{}{
		"duration": duration.Seconds(),
		"message":  message,
	})
	tag := makeTag(r.tagPrefix, commandFailTag)
	r.logger.PostWithTime(tag, endAt, record)
}

func (r *fluentdReporter) attemptStart(count int, pid int, startAt time.Time) {
	r.sr.attemptStart(count, pid, startAt)
	message := r.buf.String()
	r.buf.Reset()

	record := r.createRecord(map[string]interface{}{
		"count":   count,
		"pid":     pid,
		"message": message,
	})
	tag := makeTag(r.tagPrefix, attemptStartTag)
	r.logger.PostWithTime(tag, startAt, record)
}

func (r *fluentdReporter) attemptSucceed(count int, endAt time.Time, duration time.Duration) {
	r.sr.attemptSucceed(count, endAt, duration)
	message := r.buf.String()
	r.buf.Reset()

	record := r.createRecord(map[string]interface{}{
		"count":    count,
		"duration": duration.Seconds(),
		"message":  message,
	})
	tag := makeTag(r.tagPrefix, attemptSucceedTag)
	r.logger.PostWithTime(tag, endAt, record)
}

func (r *fluentdReporter) attemptFail(count int, err *exec.ExitError, endAt time.Time, duration time.Duration) {
	r.sr.attemptFail(count, err, endAt, duration)
	message := r.buf.String()
	r.buf.Reset()

	record := r.createRecord(map[string]interface{}{
		"count":    count,
		"duration": duration.Seconds(),
		"error":    err,
		"message":  message,
	})
	tag := makeTag(r.tagPrefix, attemptFailTag)
	r.logger.PostWithTime(tag, endAt, record)
}

func (r *fluentdReporter) attemptTimeout(count int, endAt time.Time, duration time.Duration) {
	r.sr.attemptTimeout(count, endAt, duration)
	message := r.buf.String()
	r.buf.Reset()

	record := r.createRecord(map[string]interface{}{
		"count":    count,
		"duration": duration.Seconds(),
		"message":  message,
	})
	tag := makeTag(r.tagPrefix, attemptTimeoutTag)
	r.logger.PostWithTime(tag, endAt, record)
}

func (r *fluentdReporter) attemptUnknownError(count int, err error, endAt time.Time) {
	r.sr.attemptUnknownError(count, err, endAt)
	message := r.buf.String()
	r.buf.Reset()

	record := r.createRecord(map[string]interface{}{
		"count":   count,
		"error":   err,
		"message": message,
	})
	tag := makeTag(r.tagPrefix, attemptUnknownErrorTag)
	r.logger.PostWithTime(tag, endAt, record)
}

func (r *fluentdReporter) startStdoutLogger(count int) {
	r.stdout = make(chan string)
	go func() {
		for log := range r.stdout {
			record := r.createRecord(map[string]interface{}{
				"count": count,
				"log":   log,
			})
			tag := makeTag(r.tagPrefix, stdoutTag)
			r.logger.Post(tag, record)
		}
	}()
}

func (r *fluentdReporter) finishStdoutLogger() {
	close(r.stdout)
}

func (r *fluentdReporter) stdoutLog(log string) {
	r.stdout <- log
}

func (r *fluentdReporter) startStderrLogger(count int) {
	r.stderr = make(chan string)
	go func() {
		for log := range r.stderr {
			record := r.createRecord(map[string]interface{}{
				"count": count,
				"log":   log,
			})
			tag := makeTag(r.tagPrefix, stderrTag)
			r.logger.Post(tag, record)
		}
	}()
}

func (r *fluentdReporter) finishStderrLogger() {
	close(r.stderr)
}

func (r *fluentdReporter) stderrLog(log string) {
	r.stderr <- log
}

func (r *fluentdReporter) close() {
	r.logger.Close()
}

func makeTag(s ...string) string {
	return strings.Join(s, ".")
}
