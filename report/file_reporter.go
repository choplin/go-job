package report

import (
	"fmt"
	"os"
	"path"
)

type fileReporter struct {
	stringReporter
	directory string
}

func newFileReporter(commandId string, commandName string, directory string) (*fileReporter, error) {
	logPath := fmt.Sprintf("%s/%s/%s/command.log", directory, commandName, commandId)

	err := os.MkdirAll(path.Dir(logPath), 0755)
	if err != nil {
		return nil, err
	}

	fh, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE, 0644)

	if err != nil {
		return nil, err
	}

	fr := &fileReporter{
		stringReporter{
			commandId:   commandId,
			commandName: commandName,
			fh:          fh,
		}, directory}

	return fr, nil
}

func (r *fileReporter) startStdoutLogger(count int) {
	path := fmt.Sprintf("%s/%s/%s/stdout.log.%d", r.directory, r.commandName, r.commandId, count)
	fh, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err == nil {
		r.out = fh
	}
}

func (r *fileReporter) finishStdoutLogger() {
	if r.out != nil {
		if fh, ok := r.out.(*os.File); ok {
			fh.Close()
		}
	}
}

func (r *fileReporter) startStderrLogger(count int) {
	path := fmt.Sprintf("%s/%s/%s/stderr.log.%d", r.directory, r.commandName, r.commandId, count)
	fh, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err == nil {
		r.err = fh
	}
}

func (r *fileReporter) finishStderrLogger() {
	if r.err != nil {
		if fh, ok := r.err.(*os.File); ok {
			fh.Close()
		}
	}
}

func (r *fileReporter) close() {
	if fh, ok := r.fh.(*os.File); ok {
		fh.Close()
	}
}
