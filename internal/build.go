package internal

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
)

func Build(repo *Repository, packageName string) error {
	pkg, ok := repo.Packages[packageName]
	if !ok {
		return fmt.Errorf("no such package: %q", packageName)
	}
	packageDir := path.Join(repo.OutDir, "node_modules", pkg.Name)

	if err := os.MkdirAll(packageDir, 0755); err != nil {
		return err
	}

	if err := EnsureTmp(repo); err != nil {
		return err
	}

	dependencies := make(map[string]string)

	depPrefix := "/Users/brandonbloom/Projects/unirepo/example/node_modules/"
	isFileFromDeps := func(filepath string) bool {
		return strings.HasPrefix(filepath, depPrefix)
	}

	var buildPlugin = api.Plugin{
		Name: "unirepo",
		Setup: func(build api.PluginBuild) {
			build.OnResolve(
				api.OnResolveOptions{
					Filter: `.*`,
				},
				func(args api.OnResolveArgs) (api.OnResolveResult, error) {
					if isFileFromDeps(args.Importer) {
						return api.OnResolveResult{}, nil
					}
					moduleName := args.Path
					if version, ok := repo.Dependencies[moduleName]; ok {
						dependencies[moduleName] = version
					}
					return api.OnResolveResult{}, nil
				},
			)
		},
	}

	mainRelpath := "index.cjs.js"
	result := api.Build(api.BuildOptions{
		EntryPoints: []string{pkg.Entrypoint},
		Outfile:     path.Join(packageDir, mainRelpath),
		Bundle:      true,
		Platform:    api.PlatformNode,
		Format:      api.FormatCommonJS,
		Write:       true,
		LogLevel:    api.LogLevelWarning,
		Plugins: []api.Plugin{
			buildPlugin,
		},
	})

	pkgMetadata := PackageMetadata{
		Name:         pkg.Name,
		Description:  pkg.Description,
		Main:         mainRelpath,
		Dependencies: dependencies,
	}
	if err := WritePackageJSON(pkgMetadata, packageDir); err != nil {
		return err
	}

	if len(result.Errors) > 0 {
		return fmt.Errorf("build error")
	}

	return nil
}