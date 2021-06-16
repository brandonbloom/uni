// Cases to handle
//
// without watch
//   build error
//   build ok
//     program fails to start
//     program terminates success
//     program terminates failure
//
// with watch
//   build error
//   build ok                              waitForChange
//     program fails to start                  true      wait for change
//     program terminates prematurely          true      wait for change
//     code changes                            false     restart
//     interrupt                               false     exi

package internal

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"runtime"
	"syscall"
	"time"

	"github.com/evanw/esbuild/pkg/api"
)

var maxProcStopWait = 5 * time.Second

type RunOptions struct {
	Watch      bool
	Entrypoint string
	Args       []string
	BuildOnly  bool
}

// Status code may be returned within an exec.ExitError return value.
func Run(ctx context.Context, repo *Repository, opts RunOptions) error {
	if err := EnsureTmp(repo); err != nil {
		return err
	}

	dir, err := TempDir(repo, "run")
	if err != nil {
		return err
	}
	if !opts.BuildOnly {
		defer os.RemoveAll(dir)
	}

	// See also `shim` in Build.
	script := fmt.Sprintf(`require('source-map-support').install();

const { inspect } = require('util');
process.on('uncaughtException', (exception) => {
  process.stderr.write('uncaught exception: ' + inspect(exception) + '\n', () => {
    process.exit(1);
  });
});
process.on('unhandledRejection', (reason, promise) => {
  process.stderr.write(
    'unhandled rejection at: ' + inspect(promise) + '\nreason: ' + inspect(reason) + '\n',
    () => {
      process.exit(1);
    },
  );
})

const { main } = require('./bundle.js');
if (typeof main === 'function') {
	const args = process.argv.slice(2);
	void (async () => {
		const exitCode = await main(...args);
		process.exit(exitCode ?? 0);
	})();
} else {
	process.stderr.write('error: %s does not export a main function\n', () => {
		process.exit(1);
	});
}
`, opts.Entrypoint)
	scriptPath := path.Join(dir, "script.js")
	if err := ioutil.WriteFile(scriptPath, []byte(script), 0644); err != nil {
		return err
	}

	return buildAndWatch{
		Repository: repo,
		Watch:      opts.Watch && !opts.BuildOnly,
		Esbuild: api.BuildOptions{
			AbsWorkingDir: repo.RootDir,
			EntryPoints:   []string{opts.Entrypoint},
			Outfile:       path.Join(dir, "bundle.js"),
			Bundle:        true,
			Platform:      api.PlatformNode,
			Format:        api.FormatCommonJS,
			Write:         true,
			LogLevel:      api.LogLevelWarning,
			Sourcemap:     api.SourceMapLinked,
			External:      getExternals(repo),
			Loader:        loaders,
		},
		CreateProcess: func() process {
			if opts.BuildOnly {
				return &funcProcess{
					start: func() error {
						fmt.Println(dir)
						return nil
					},
				}
			}

			nodeArgs := append([]string{scriptPath}, opts.Args...)
			node := exec.Command("node", nodeArgs...)
			node.Stdin = os.Stdin
			node.Stdout = os.Stdout
			node.Stderr = os.Stderr
			node.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

			return &cmdProcess{cmd: node}
		},
	}.Run(ctx)
}

type cmdProcess struct {
	cmd *exec.Cmd
}

func (proc *cmdProcess) Start() error {
	return proc.cmd.Start()
}

func (proc *cmdProcess) Stop() error {
	if proc.cmd.Process == nil {
		return nil
	}

	if runtime.GOOS == "windows" {
		return proc.cmd.Process.Kill()
	}

	go func() {
		// TODO: Make the wait time configurable.
		time.Sleep(maxProcStopWait)
		_ = syscall.Kill(-proc.cmd.Process.Pid, syscall.SIGKILL)
	}()

	return syscall.Kill(-proc.cmd.Process.Pid, syscall.SIGTERM)
}

func (proc *cmdProcess) Wait() error {
	if proc.cmd.Process == nil {
		return nil
	}
	return proc.cmd.Wait()
}
