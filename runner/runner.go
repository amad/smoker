package runner

import (
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
	closeChan := make(chan struct{}, workers)
	reports := []core.TestResult{}

	return &Runner{
		workers:       workers,
		timeout:       timeout,
		stopOnFailure: stopOnFailure,
		closeChan:     closeChan,
		stdout:        stdout,
		stderr:        stderr,
		reports:       reports,
	}
}

// Runner is a type to manage and run tests and provide logs.
type Runner struct {
	workers        int
	timeout        time.Duration
	stopOnFailure  bool
	closeChan      chan struct{}
	stdout, stderr io.StringWriter
	reports        []core.TestResult
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
		if r.isClosing() {
			break
		}

		poolChan <- struct{}{}
		wg.Add(2) // delta=2 to sync worker and reportWriter.

		go r.worker(&wg, requester, i+1, tc, poolChan, reportsChan)
	}

	wg.Wait()

	close(r.closeChan)
	close(reportsChan)
	close(poolChan)

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
	r.closeChan <- struct{}{}
}

func (r *Runner) isClosing() bool {
	select {
	case <-r.closeChan:
		return true
	default:
		return false
	}
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
