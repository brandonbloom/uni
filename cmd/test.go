package cmd

import (
	"errors"
	"os"
	"os/exec"

	"github.com/deref/uni/internal"
	"github.com/spf13/cobra"
)

var testOpts = internal.TestOptions{}

func init() {
	rootCmd.AddCommand(testCmd)
	testCmd.Flags().BoolVar(&testOpts.Watch, "watch", false, "re-runs tests when source files change")
	testCmd.Flags().BoolVar(&testOpts.Coverage, "coverage", false, "Run testes with coverage")
}

var testCmd = &cobra.Command{
	Use:                   "test [flags] <script> [args...]",
	Short:                 "Run tests.",
	Long:                  `Run tests using jest.`,
	DisableFlagsInUseLine: true,
	SilenceErrors:         true,
	RunE: func(cmd *cobra.Command, args []string) error {
		repo := mustLoadRepository()
		if err := internal.CheckEngines(repo); err != nil {
			return err
		}

		err := internal.Test(repo, testOpts)
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		return err
	},
}
