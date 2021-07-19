package internal

import (
	"os"
	"os/exec"
)

type InstallDependenciesOptions struct {
	Frozen bool
}

func InstallDependencies(repo *Repository, opts InstallDependenciesOptions) error {

	if err := ConfigureRepository(repo); err != nil {
		return err
	}
	if repo.IsWorkspace {
		for _, pkg := range repo.Packages {
			if err := configurePkg(repo, pkg); err != nil {
				return err
			}
		}

		execYarn := exec.Command("yarn")
		execYarn.Stdin = os.Stdin
		execYarn.Stdout = os.Stdout
		execYarn.Stderr = os.Stderr
		return execYarn.Run()

	} else {
		subcommand := "install"
		if opts.Frozen {
			subcommand = "ci"
		}

		npm := exec.Command("npm", subcommand)
		npm.Stdin = os.Stdin
		npm.Stdout = os.Stdout
		npm.Stderr = os.Stderr
		return npm.Run()
	}
}
