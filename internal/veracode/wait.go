package veracode

import (
	"fmt"
	"time"
)

// WaitOptions tunes WaitForLatestScanResults.
//
// The wait is always bounded by Timeout, so a stuck scan fails the command
// instead of hanging a CI pipeline forever. Progress goes to Logf (the CLI
// wires it to stderr, keeping stdout clean for scripting).
type WaitOptions struct {
	Interval time.Duration
	Timeout  time.Duration
	Logf     func(format string, args ...any)
	Sleep    func(time.Duration)
}

func (o *WaitOptions) withDefaults() (logf func(string, ...any), sleep func(time.Duration)) {
	logf = o.Logf
	if logf == nil {
		logf = func(string, ...any) {}
	}
	sleep = o.Sleep
	if sleep == nil {
		sleep = time.Sleep
	}
	if o.Interval <= 0 {
		o.Interval = 30 * time.Second
	}
	if o.Timeout <= 0 {
		o.Timeout = 15 * time.Minute
	}
	return logf, sleep
}

// WaitForLatestScanResults returns the application, waiting first when a scan
// is currently in progress so the caller gets the fresh results URL.
//
// Behavior:
//   - No builds, or latest build already published: returns immediately.
//   - Latest build failed/cancelled/invalid: returns an error immediately
//     (waiting would never produce new results).
//   - Scan alive: polls until a new results URL is published, the build
//     reaches a terminal state, or Timeout expires (error, pipeline-safe).
func (c *Client) WaitForLatestScanResults(name string, opts WaitOptions) (*Application, error) {
	logf, sleep := opts.withDefaults()

	app, err := c.GetApplication(name)
	if err != nil {
		return nil, err
	}
	initialURL := BuildLatestScanResultsURL(app.ResultsURL)

	info, err := c.GetBuildInfo(app.ID)
	if err != nil {
		return nil, err
	}

	if info.AppID == "" {
		logf("no builds found for %q; nothing to wait for", name)
		return app, nil
	}

	status := info.StaticStatus()

	if info.Build.ResultsReady {
		logf("latest scan for %q already published (status %q)", name, status)
		return app, nil
	}

	if !ScanInProgress(status) {
		return nil, fmt.Errorf(
			"latest scan for %q is not running (status %q); no new results to wait for",
			name,
			status,
		)
	}

	logf("scan for %q in progress (status %q); waiting up to %s for new results",
		name, status, opts.Timeout)

	deadline := time.Now().Add(opts.Timeout)

	for {
		info, err := c.GetBuildInfo(app.ID)
		if err != nil {
			return nil, err
		}
		status = info.StaticStatus()

		if ScanTerminalFailure(status) {
			return nil, fmt.Errorf(
				"scan for %q ended with status %q; no new results published",
				name,
				status,
			)
		}

		current, err := c.GetApplication(name)
		if err != nil {
			return nil, err
		}
		currentURL := BuildLatestScanResultsURL(current.ResultsURL)

		if info.Build.ResultsReady || (currentURL != "" && currentURL != initialURL) {
			logf("new results published for %q (status %q)", name, status)
			return current, nil
		}

		if !ScanInProgress(status) {
			return nil, fmt.Errorf(
				"scan for %q left the running state (%q) without publishing new results",
				name,
				status,
			)
		}

		remaining := time.Until(deadline)
		if remaining <= 0 {
			return nil, fmt.Errorf(
				"timed out after %s waiting for new results of %q (last status %q)",
				opts.Timeout,
				name,
				status,
			)
		}

		logf("still waiting for %q (status %q, %s left)", name, status, remaining.Round(time.Second))
		if opts.Interval < remaining {
			sleep(opts.Interval)
		} else {
			sleep(remaining)
		}
	}
}
