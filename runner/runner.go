package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/amad/smoker/core"
	"github.com/amad/smoker/runner/internal/report"
)

// NewRunner creates and returns a new Runner.
func NewRunner(workers int, timeout time.Duration, stopOnFailure bool, stdout io.StringWriter, stderr io.StringWriter) *Runner {
	reports := []core.TestResult{}

	ctx, cancelFunc := context.WithCancel(context.Background())

	return &Runner{
		workers:       workers,
		timeout:       timeout,
		stopOnFailure: stopOnFailure,
		stdout:        stdout,
		stderr:        stderr,
		reports:       reports,
		ctx:           ctx,
		cancelFunc:    cancelFunc,
	}
}

// Runner is a type to manage and run tests and provide logs.
type Runner struct {
	workers        int
	timeout        time.Duration
	stopOnFailure  bool
	stdout, stderr io.StringWriter
	reports        []core.TestResult
	ctx            context.Context
	cancelFunc     context.CancelFunc
}

// Run smoke test on a testsuite and provides results.
func (r *Runner) Run(requester core.Requester, testsuite *core.Testsuite) (bool, error) {
	if len(testsuite.Tests) < 1 {
		return false, errors.New("no testcase found in this testsuite")
	}

	r.printfOut("Tests:   %d total", len(testsuite.Tests))
	r.printfOut("Workers: %d total", r.workers)
	r.printfOut("Timeout: %s", r.timeout.String())
	r.printfOut("Stop on failure: %t\n", r.stopOnFailure)

	start := time.Now()

	var wg sync.WaitGroup
	reportsChan := make(chan core.TestResult, len(testsuite.Tests))
	poolChan := make(chan struct{}, r.getPoolsize(testsuite))

	r.printfOut("Waiting for results\n")
	go r.reportWriter(&wg, reportsChan)

	for i, tc := range testsuite.Tests {
		poolChan <- struct{}{}

		if r.isClosing() {
			break
		}

		wg.Add(2) // delta=2 to sync worker and reportWriter.
		go r.worker(&wg, requester, i+1, tc, poolChan, reportsChan)
	}

	wg.Wait()

	close(reportsChan)
	close(poolChan)
	defer r.cancelFunc()

	r.printfOut("\nElapsed: %.2fs", time.Since(start).Seconds())

	for _, rp := range r.reports {
		if !rp.Passed() {
			return false, nil
		}
	}

	return true, nil
}

func (r *Runner) worker(wg *sync.WaitGroup, requester core.Requester, idx int, tc core.TestCase, pool <-chan struct{}, reportsChan chan<- core.TestResult) {
	defer wg.Done()

	s := time.Now()
	res, err := requester.Request(tc)

	reportsChan <- &report.TestReport{
		Index:    idx,
		Name:     tc.Name,
		Status:   res,
		Err:      err,
		Duration: time.Since(s),
	}

	if !res {
		r.shouldStopOnFailure()
	}

	<-pool
}

// Stop pauses off the runner.
// used for signal handling or when stop on failure is enabled.
func (r *Runner) Stop() {
	r.cancelFunc()
}

func (r *Runner) isClosing() bool {
	if err := r.ctx.Err(); err != nil {
		return true
	}

	return false
}

func (r *Runner) shouldStopOnFailure() {
	if r.stopOnFailure {
		r.Stop()
	}
}

func (r *Runner) reportWriter(wg *sync.WaitGroup, reportsChan <-chan core.TestResult) {
	for report := range reportsChan {
		r.reports = append(r.reports, report)

		if report.Passed() {
			r.printfOut(report.String())
		} else {
			r.printfErrOut(report.String())
		}

		wg.Done()
	}
}

func (r *Runner) getPoolsize(ts *core.Testsuite) int {
	if r.workers >= len(ts.Tests) {
		return len(ts.Tests)
	}

	return r.workers
}

func (r *Runner) printfOut(msg string, params ...interface{}) {
	_, err := r.stdout.WriteString(fmt.Sprintf(msg, params...) + "\n")
	if err != nil {
		log.Fatal(err)
	}
}

func (r *Runner) printfErrOut(msg string, params ...interface{}) {
	_, err := r.stderr.WriteString(fmt.Sprintf(msg, params...) + "\n")
	if err != nil {
		log.Fatal(err)
	}
}
