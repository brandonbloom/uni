package cmd

import (
	"errors"
	"os"
	"os/exec"

	"github.com/deref/uni/internal"
	"github.com/spf13/cobra"
)

var lintOpts = internal.LintOptions{}

func init() {
	rootCmd.AddCommand(lintCmd)
	lintCmd.Flags().BoolVar(&lintOpts.Fix, "fix", false, "fix lint errors")
}

var lintCmd = &cobra.Command{
	Use:                   "lint [flags] <script> [args...]",
	Short:                 "Lint.",
	Long:                  `Run eslint in all workspace.`,
	DisableFlagsInUseLine: true,
	SilenceErrors:         true,
	RunE: func(cmd *cobra.Command, args []string) error {
		repo := mustLoadRepository()
		if err := internal.CheckEngines(repo); err != nil {
			return err
		}

		err := internal.Lint(repo, lintOpts)
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		return err
	},
}
