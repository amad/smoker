package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amad/smoker/cmdoptions"
	"github.com/amad/smoker/loader"
	"github.com/amad/smoker/requester"
	"github.com/amad/smoker/runner"
	"github.com/amad/smoker/version"
)

func main() {
	var err error

	flags, err := cmdoptions.InstallFlags(version.String(), os.Stdout)
	exitIfError(err)

	testsuite, err := loader.LoadTestsuite(flags.TestsuiteFile)
	exitIfError(err)

	runner := runner.NewRunner(flags.Workers, flags.Timeout, flags.StopOnFailure, os.Stdout, os.Stderr)
	requester := requester.NewRequester(flags.Timeout, fmt.Sprintf("smoker/%s", version.String()))

	sigsChan := make(chan os.Signal)
	signal.Notify(sigsChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigsChan

		runner.Stop()

		exitWithError(errors.New("Interrupted"))
	}()

	ok, err := runner.Run(requester, testsuite)
	exitIfError(err)

	fmt.Println("Done")

	if !ok {
		os.Exit(1)
	}
}

func exitIfError(err error) {
	if err == nil {
		return
	}

	exitWithError(err)
}

func exitWithError(e error) {
	os.Stderr.WriteString(fmt.Sprintf("ERROR: %s\n", e.Error()))
	os.Exit(1)
}
