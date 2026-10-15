package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/appsecomega/veracode-go-cli/internal/config"
	"github.com/appsecomega/veracode-go-cli/internal/veracode"
)

// outputFormat is shared by subcommands that print results.
type outputFormat string

const (
	outputText outputFormat = "text"
	outputJSON outputFormat = "json"
)

// newAppCmd groups application-related subcommands.
// Pattern to scale: add `list`, `create`, etc. as new subcommands here.
func newAppCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "app",
		Short: "Operações com aplicações",
	}

	cmd.AddCommand(newAppGetCmd())

	return cmd
}

// newAppGetCmd implements: veracode-go-cli app get --name "WebGoat-Legacy-master"
func newAppGetCmd() *cobra.Command {
	var (
		name         string
		output       string
		appURLOnly   bool
		scanURLOnly  bool
		wait         bool
		waitTimeout  time.Duration
		waitInterval time.Duration
	)

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Busca uma aplicação pelo nome exato",
		Example: `  veracode-go-cli app get --name "WebGoat-Legacy-master"
  veracode-go-cli app get --name "WebGoat-Legacy-master" --output json
  veracode-go-cli app get --name "WebGoat-Legacy-master" --app-url
  veracode-go-cli app get --name "WebGoat-Legacy-master" --scan-url --wait --timeout 30m`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			name = strings.TrimSpace(name)
			if name == "" {
				return fmt.Errorf("application name is required (use --name)")
			}

			format := outputFormat(strings.ToLower(strings.TrimSpace(output)))
			if format != outputText && format != outputJSON {
				return fmt.Errorf("invalid --output %q: want \"text\" or \"json\"", output)
			}

			if wait && (waitTimeout <= 0 || waitInterval <= 0) {
				return fmt.Errorf("invalid --timeout/--interval: both must be positive durations")
			}

			credentials, err := config.Load()
			if err != nil {
				return err
			}

			client := veracode.NewClient(credentials)

			var application *veracode.Application
			if wait {
				application, err = client.WaitForLatestScanResults(name, veracode.WaitOptions{
					Interval: waitInterval,
					Timeout:  waitTimeout,
					Logf: func(format string, args ...any) {
						_, _ = fmt.Fprintf(cmd.ErrOrStderr(), format+"\n", args...)
					},
				})
			} else {
				application, err = client.GetApplication(name)
			}
			if err != nil {
				return err
			}

			// Bare mode: script-friendly, one URL per line on stdout.
			// --output is ignored here by design.
			if appURLOnly || scanURLOnly {
				return printBareURLs(cmd, application, appURLOnly, scanURLOnly)
			}

			return printApplication(cmd, application, format)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Veracode application name (exact match)")
	cmd.Flags().StringVarP(&output, "output", "o", "text", "Output format: text|json (ignored with --app-url/--scan-url)")
	cmd.Flags().BoolVar(&appURLOnly, "app-url", false, "Print only the Analysis Center application URL")
	cmd.Flags().BoolVar(&scanURLOnly, "scan-url", false, "Print only the latest published scan results URL")
	cmd.Flags().BoolVar(&wait, "wait", false, "Wait for an in-progress scan to publish before printing")
	cmd.Flags().DurationVar(&waitTimeout, "timeout", 15*time.Minute, "Max wait with --wait (e.g. 10m, 1h); command fails instead of hanging")
	cmd.Flags().DurationVar(&waitInterval, "interval", 30*time.Second, "Poll interval with --wait (e.g. 30s, 1m)")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

// printBareURLs prints the requested URLs, one per line, for scripting.
func printBareURLs(cmd *cobra.Command, app *veracode.Application, wantApp, wantScan bool) error {
	out := cmd.OutOrStdout()

	if wantApp {
		_, _ = fmt.Fprintln(out, veracode.BuildAnalysisCenterURL(app.AppProfileURL))
	}

	if wantScan {
		scanURL := veracode.BuildLatestScanResultsURL(app.ResultsURL)
		if scanURL == "" {
			return fmt.Errorf("application %q has no published policy scan yet", app.Profile.Name)
		}
		_, _ = fmt.Fprintln(out, scanURL)
	}

	return nil
}

func printApplication(cmd *cobra.Command, app *veracode.Application, format outputFormat) error {
	analysisCenterURL := veracode.BuildAnalysisCenterURL(app.AppProfileURL)
	latestScanURL := veracode.BuildLatestScanResultsURL(app.ResultsURL)

	if format == outputJSON {
		payload := map[string]any{
			"application":         app.Profile.Name,
			"app_id":              app.ID,
			"guid":                app.GUID,
			"analysis_center":     analysisCenterURL,
			"last_completed_scan": app.LastCompletedScanDate,
			"latest_scan_results": latestScanURL,
		}
		raw, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return fmt.Errorf("encode output: %w", err)
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(raw))
		return nil
	}

	out := cmd.OutOrStdout()
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, "Application:", app.Profile.Name)
	_, _ = fmt.Fprintln(out, "App ID:", app.ID)
	_, _ = fmt.Fprintln(out, "GUID:", app.GUID)
	_, _ = fmt.Fprintln(out, "Analysis Center:", analysisCenterURL)
	if app.LastCompletedScanDate != "" {
		_, _ = fmt.Fprintln(out, "Last completed scan:", app.LastCompletedScanDate)
	}
	if latestScanURL != "" {
		_, _ = fmt.Fprintln(out, "Latest scan results:", latestScanURL)
	} else {
		_, _ = fmt.Fprintln(out, "Latest scan results: (no published policy scan yet)")
	}

	return nil
}
