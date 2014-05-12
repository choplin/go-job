package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/choplin/go-job/command"
	"github.com/choplin/go-job/report"
)

var (
	timeout          = flag.Duration("timeout", time.Duration(0), "timeout duration. See time.ParseDuration on Go document.")
	attempt          = flag.Int("attempt", 1, "maximum number of attempt")
	name             = flag.String("name", "", "A name for this command. A default value is a basename of the specified command path.")
	reporters        = flag.String("reporters", "console", "log reporters. you can specify multiple reporters with ',' as delimiter. available: console, fluentd, file.")
	fluentdHost      = flag.String("fluentd-host", "localhost", "fluentd host")
	fluentdPort      = flag.Int("fluentd-port", 24224, "fluentd port")
	fluentdTagPrefix = flag.String("fluentd-tag-prefix", "command", "fluentd tag prefix")
	fileDirectory    = flag.String("file-directory", "/var/log/go_job", "a base directory of file reporter. log files will be stored under ${file direcotry}/${command name}/${command id}.")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s command [args...]\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()
	args := flag.Args()

	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "command must be specified\n")
		flag.Usage()
		os.Exit(1)
	}

	reporterConfig := &report.ReporterConfig{
		Reporters:        *reporters,
		FluentdHost:      *fluentdHost,
		FluentdPort:      *fluentdPort,
		FluentdTagPrefix: *fluentdTagPrefix,
		FileDirectory:    *fileDirectory,
	}

	if *name == "" {
		*name = path.Base(args[0])
	}

	command, err := command.NewCommand(*name, timeout, *attempt, reporterConfig, args[0], args[1:]...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initalize command. %s\n", err)
		os.Exit(1)
	}

	done := command.Start()

	if success := <-done; success {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
