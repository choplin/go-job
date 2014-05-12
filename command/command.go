package command

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os/exec"
	"time"

	"github.com/choplin/go-job/report"
)

type attemptTimeoutError struct {
	endAt    time.Time
	duration time.Duration
}

func (err *attemptTimeoutError) Error() string {
	return "timeout"
}

type attemptExitError struct {
	endAt    time.Time
	duration time.Duration
	err      *exec.ExitError
}

func (err *attemptExitError) Error() string {
	return err.err.Error()
}

type Command struct {
	id         string
	name       string
	commandStr string
	args       []string
	timeout    *time.Duration
	maxAttempt int
	reporters  report.ReporterList
}

func NewCommand(name string, timeout *time.Duration, maxAttempt int, reporterConfig *report.ReporterConfig, commandStr string, args ...string) (*Command, error) {
	id := generateId()

	reporters, err := report.NewReporterList(id, name, reporterConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize reporter: %s", err)
	}

	return &Command{
		id,
		name,
		commandStr,
		args,
		timeout,
		maxAttempt,
		reporters}, nil
}

func (c *Command) Start() chan bool {
	done := make(chan bool)
	go func() {
		success := false
		startAt := time.Now()
		c.reporters.CommandStart(startAt)

		for attemptCount := 1; attemptCount <= c.maxAttempt; attemptCount++ {
			if err := c.attempt(attemptCount); err != nil {

				switch e := err.(type) {
				case *attemptExitError:
					c.reporters.AttemptFail(attemptCount, e.err, e.endAt, e.duration)
				case *attemptTimeoutError:
					c.reporters.AttemptTimeout(attemptCount, e.endAt, e.duration)
				default:
					c.reporters.AttemptUnknownError(attemptCount, e, time.Now())
				}
			} else {
				success = true
				break
			}
		}
		endAt := time.Now()
		if success {
			c.reporters.CommandSucceed(endAt, endAt.Sub(startAt))
		} else {
			c.reporters.CommandFail(endAt, endAt.Sub(startAt))
		}
		done <- success
	}()
	return done
}

func (c *Command) attempt(count int) error {
	startAt := time.Now()
	cmd := exec.Command(c.commandStr, c.args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create a stdout pipe. %s", err)
	}
	waitStdout := make(chan bool)
	go func() {
		c.reporters.StartStdoutLogger(count)
		defer c.reporters.FinishStdoutLogger()

		buf := make([]byte, 4096)
		var err error
		var n int
		for err == nil {
			n, err = stdout.Read(buf)
			if n > 0 {
				c.reporters.StdoutLog(string(buf[0:n]))
			}
		}
		close(waitStdout)
	}()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create a stderr pipe. %s", err)
	}
	waitStderr := make(chan bool)
	go func() {
		c.reporters.StartStderrLogger(count)
		defer c.reporters.FinishStderrLogger()

		buf := make([]byte, 4096)
		var err error
		var n int
		for err == nil {
			n, err = stderr.Read(buf)
			if n > 0 {
				c.reporters.StderrLog(string(buf[0:n]))
			}
		}
		close(waitStderr)
	}()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process(%s %). %s", cmd.Path, cmd.Args, err)
	}

	pid := cmd.Process.Pid
	c.reporters.AttemptStart(count, pid, startAt)

	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()

	var timer <-chan time.Time
	if *c.timeout == 0 {
		timer = nil
	} else {
		timer = time.After(*c.timeout)
	}
	select {
	case <-timer:
		if err := cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process %d. %s", pid, err)
		}
		<-done
		<-waitStdout
		<-waitStderr
		endAt := time.Now()
		return &attemptTimeoutError{endAt, endAt.Sub(startAt)}
	case err := <-done:
		<-waitStdout
		<-waitStderr
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				endAt := time.Now()
				return &attemptExitError{endAt, endAt.Sub(startAt), exitErr}
			} else {
				fmt.Errorf("process exited with unknown error. %s", err)
			}
			return fmt.Errorf("exit with failure")
		}
	}
	endAt := time.Now()
	c.reporters.AttemptSucceed(count, endAt, endAt.Sub(startAt))
	return nil
}

func (c *Command) Close() {
	c.reporters.Close()
}

func generateId() string {
	randBytes := make([]byte, 4)
	rand.Read(randBytes)

	t := time.Now()

	return fmt.Sprintf("%d%02d%02d-%02d%02d%02d-%s", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), hex.EncodeToString(randBytes))
}
