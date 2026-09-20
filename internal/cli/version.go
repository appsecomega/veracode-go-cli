package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/appsecomega/veracode-go-cli/internal/version"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Mostra a versão do binário",
		Run: func(cmd *cobra.Command, _ []string) {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), version.Short())
		},
	}
}
