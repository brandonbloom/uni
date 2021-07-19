package internal

import (
	"os"
	"os/exec"
)

type LintOptions struct {
	Fix bool
}

func Lint(repo *Repository, opts LintOptions) error {
	if err := EnsureTmp(repo); err != nil {
		return err
	}

	var lint *exec.Cmd
	if opts.Fix {
		lint = exec.Command("npx", "eslint", ".", "--ext", ".ts,.tsx", "--fix")
	} else {
		lint = exec.Command("npx", "eslint", ".", "--ext", ".ts,.tsx")
	}
	lint.Stdin = os.Stdin
	lint.Stdout = os.Stdout
	lint.Stderr = os.Stderr
	return lint.Run()
}
