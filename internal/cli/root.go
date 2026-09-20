package cli

import (
	"github.com/spf13/cobra"
)

// NewRootCmd builds the root command.
// New subcommands go here — e.g. rootCmd.AddCommand(newScanCmd()).
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "veracode-go-cli",
		Short: "CLI alternativa para a plataforma Veracode",
		Long: `veracode-go-cli — CLI alternativa para consultar informações
de aplicações na plataforma Veracode.

Exemplos:
  veracode-go-cli app get --name "WebGoat-Legacy-master"
  veracode-go-cli app get --name "WebGoat-Legacy-master" --output json
  veracode-go-cli version`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.AddCommand(newAppCmd())
	rootCmd.AddCommand(newVersionCmd())

	return rootCmd
}

// Execute runs the root command. Kept thin so main.go stays trivial
// and commands remain testable.
func Execute() error {
	return NewRootCmd().Execute()
}
