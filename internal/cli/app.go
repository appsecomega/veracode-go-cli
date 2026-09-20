package cli

import (
	"encoding/json"
	"fmt"
	"strings"

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
		name   string
		output string
	)

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Busca uma aplicação pelo nome exato",
		Example: `  veracode-go-cli app get --name "WebGoat-Legacy-master"
  veracode-go-cli app get --name "WebGoat-Legacy-master" --output json`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			name = strings.TrimSpace(name)
			if name == "" {
				return fmt.Errorf("application name is required (use --name)")
			}

			format := outputFormat(strings.ToLower(strings.TrimSpace(output)))
			if format != outputText && format != outputJSON {
				return fmt.Errorf("invalid --output %q: want \"text\" or \"json\"", output)
			}

			credentials, err := config.Load()
			if err != nil {
				return err
			}

			client := veracode.NewClient(credentials)

			application, err := client.GetApplication(name)
			if err != nil {
				return err
			}

			return printApplication(cmd, application, format)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Veracode application name (exact match)")
	cmd.Flags().StringVarP(&output, "output", "o", "text", "Output format: text|json")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func printApplication(cmd *cobra.Command, app *veracode.Application, format outputFormat) error {
	analysisCenterURL := veracode.BuildAnalysisCenterURL(app.AppProfileURL)

	if format == outputJSON {
		payload := map[string]any{
			"application":     app.Profile.Name,
			"app_id":          app.ID,
			"guid":            app.GUID,
			"analysis_center": analysisCenterURL,
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

	return nil
}
