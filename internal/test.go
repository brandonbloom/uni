package internal

import (
	"os"
	"os/exec"
)

type TestOptions struct {
	Watch    bool
	Coverage bool
}

func Test(repo *Repository, opts TestOptions) error {
	if err := EnsureTmp(repo); err != nil {
		return err
	}

	var args = []string{"jest"}
	if opts.Watch {
		args = append(args, "--watch")
	}
	if opts.Coverage {
		args = append(args, "--coverage")
	}
	jest := exec.Command("npx", args...)
	jest.Stdin = os.Stdin
	jest.Stdout = os.Stdout
	jest.Stderr = os.Stderr
	return jest.Run()
}
