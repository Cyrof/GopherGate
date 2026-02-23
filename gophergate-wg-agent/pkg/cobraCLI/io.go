package cobraCLI

import (
	"fmt"

	"github.com/spf13/cobra"
)

func out(cmd *cobra.Command, s string)            { fmt.Fprint(cmd.OutOrStdout(), s) }
func outf(cmd *cobra.Command, f string, a ...any) { fmt.Fprintf(cmd.OutOrStdout(), f, a...) }

func errf(cmd *cobra.Command, f string, a ...any) { fmt.Fprintf(cmd.ErrOrStderr(), f, a...) }
