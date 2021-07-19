package internal

import (
	"errors"
	"io/ioutil"
	"path"
	"path/filepath"
)

func ConfigureRepository(repo *Repository) error {
	var workspaceName string
	var workspaceVersion string
	if repo.IsWorkspace {
		workspaceName = repo.Workspace.Name
		workspaceVersion = repo.Workspace.Version
		if err := configureGitIgnore(repo); err != nil {
			return err
		}
		if err := configureNvmRc(repo); err != nil {
			return err
		}
		if err := configureRootTsConfigSettings(repo); err != nil {
			return err
		}
		if err := configureRootTsConfigReferences(repo); err != nil {
			return err
		}
		if err := configureRootTsConfigTest(repo); err != nil {
			return err
		}
		if err := configureJest(repo); err != nil {
			return err
		}
		if err := configureEsLint(repo); err != nil {
			return err
		}
	} else {
		workspaceName = "@unirepo/placeholder"
		workspaceVersion = "0.0.1"
	}
	metadata := PackageMetadata{
		Name:         workspaceName,
		Version:      workspaceVersion,
		Private:      true,
		Description:  "GENERATED FILE: DO NOT EDIT! This file is managed by unirepo.",
		Dependencies: make(map[string]string),
		Repository:   repo.Url,
	}
	if repo.IsWorkspace {
		metadata.Workspaces = []string{}
		for _, pkg := range repo.Packages {
			metadata.Workspaces = append(metadata.Workspaces, pkg.Folder)
		}
	} else {
		metadata.Scripts = map[string]string{
			"postinstall": "patch-package",
		}
	}
	addDependenciesToPackageJSON(metadata, repo.Dependencies)
	return WritePackageJSON(metadata, repo.RootDir)
}

func configureGitIgnore(repo *Repository) error {

	gitignore := path.Join(repo.RootDir, ".gitignore")
	var content = `# unirepo
node_modules
package.json
package-lock.json
yarn-lock.*
tsconfig.*
tmp
out
.eslintrc.js
coverage
`

	err := ioutil.WriteFile(gitignore, []byte(content), 0644)
	return err
}

func configureNvmRc(repo *Repository) error {

	npmRc := path.Join(repo.RootDir, ".nvm.rc")
	var content = repo.Engines["node"]

	err := ioutil.WriteFile(npmRc, []byte(content), 0644)
	return err
}

func configureRootTsConfigSettings(repo *Repository) error {
	meta := TsConfigMetadata{
		CompilerOptions: TsConfigCompileOptionsMetadata{
			Incremental:      true,
			Target:           "ES2021",
			Module:           "commonjs",
			Declaration:      true,
			SourceMap:        true,
			ImportHelpers:    true,
			Strict:           true,
			ModuleResolution: "node",
			EsModuleInterop:  true,
			RootDir:          ".",
			BaseURL:          ".",
			Paths:            make(map[string][]string),
		},
		Exclude: []string{"node_modules", "dist", "__tests__", "**/*.test*", "**/*.spec*"},
	}
	return WriteTsConfigJSON(meta, path.Join(repo.RootDir, "tsconfig.settings.json"))
}

func configureRootTsConfigReferences(repo *Repository) error {
	meta := TsConfigMetadata{
		References: []TsConfigReferenceMetadata{},
	}
	for _, pkg := range repo.Packages {
		ref := TsConfigReferenceMetadata{
			Path: "./" + pkg.Folder,
		}
		meta.References = append(meta.References, ref)
	}
	return WriteTsConfigJSON(meta, path.Join(repo.RootDir, "tsconfig.json"))
}

func configureRootTsConfigTest(repo *Repository) error {
	meta := TsConfigMetadata{
		Extends: "./tsconfig.settings.json",
		CompilerOptions: TsConfigCompileOptionsMetadata{
			Target:  "ES2021",
			RootDir: ".",
			BaseURL: ".",
			Paths:   make(map[string][]string),
		},
		Exclude: []string{"node_modules", "dist"},
	}
	for _, pkg := range repo.Packages {
		meta.CompilerOptions.Paths[pkg.Name] = []string{"./" + pkg.Folder + "/src"}
	}
	return WriteTsConfigJSON(meta, path.Join(repo.RootDir, "tsconfig.test.json"))
}

func configurePkg(repo *Repository, pkg *Package) error {
	var packageVersion string
	if pkg.Version == "" {
		if repo.IsWorkspace {
			packageVersion = repo.Workspace.Version
		} else {
			packageVersion = "0.0.1"
		}
	} else {
		packageVersion = pkg.Version
	}

	metadata := PackageMetadata{
		Name:         pkg.Name,
		Version:      packageVersion,
		Description:  "GENERATED FILE: DO NOT EDIT! This file is managed by unirepo.",
		Dependencies: make(map[string]string),
		Repository:   repo.Url,
		Scripts: map[string]string{
			"postinstall": "patch-package",
		},
	}
	addDependenciesToPackageJSON(metadata, pkg.Dependencies)
	err := WritePackageJSON(metadata, repo.RootDir+"/"+pkg.Folder)
	if err == nil {
		err = configurePkgTsConfigReferences(repo, pkg)
	}
	return err
}

func configurePkgTsConfigReferences(repo *Repository, pkg *Package) error {
	relativePathToRoot, err := filepath.Rel("/"+pkg.Folder, "/")
	if err != nil {
		return err
	}
	meta := TsConfigMetadata{
		Extends: relativePathToRoot + "/tsconfig.json",
		CompilerOptions: TsConfigCompileOptionsMetadata{
			OutDir:    "./dist/cjs",
			RootDir:   "./src",
			BaseURL:   ".",
			Composite: true,
			Paths:     make(map[string][]string),
		},
		Exclude: []string{"dist"},
		Include: []string{"src/**/*.ts", "src/**/*.tsx"},
	}
	for _, dependency := range pkg.Dependencies {
		if dependency.Version != "*" {
			return errors.New("package dependencies is supported only inside workspace")
		}
		var depPkg = repo.Packages[dependency.Name]
		if depPkg == nil {
			return errors.New("package not found " + dependency.Name)
		}
		relativePathToDep, err := filepath.Rel(pkg.Folder, depPkg.Folder)
		if err != nil {
			return err
		}

		ref := TsConfigReferenceMetadata{
			Path: relativePathToDep,
		}
		meta.References = append(meta.References, ref)
		meta.CompilerOptions.Paths[dependency.Name] = []string{relativePathToDep + "/src"}
	}
	return WriteTsConfigJSON(meta, path.Join(repo.RootDir, pkg.Folder, "tsconfig.json"))
}

func addDependenciesToPackageJSON(metadata PackageMetadata, dependencies map[string]*Dependency) {
	for dependencyName, dependency := range dependencies {
		metadata.Dependencies[dependencyName] = dependency.Version
	}
}

func configureJest(repo *Repository) error {
	const content = `const { pathsToModuleNameMapper } = require('ts-jest/utils')
const { compilerOptions } = require('./tsconfig.test')

const moduleNameMapper = pathsToModuleNameMapper(compilerOptions.paths, { prefix: '<rootDir>/' })

module.exports = {
	preset: 'ts-jest',
	modulePathIgnorePatterns: ['dist'],
	testPathIgnorePatterns: ['node_modules', 'dist'],
	testRegex: '(\\.(test|spec|steps))\\.(ts|tsx)$',
	globals: {
		'ts-jest': {
			tsConfig: 'tsconfig.test.json'
		}
	},
	moduleNameMapper,
	transform: {
		'^.+\\.tsx?$': 'esbuild-jest'
	},
	coverageReporters: [
		'text',
		'html',
		'cobertura',
		'json-summary'
	],
	coverageThreshold: {
		global: {
			lines: 90,
			statements: 90,
			functions: 90,
			branches: 90
		}
	}
}`
	jestConfig := path.Join(repo.RootDir, "jest.config.js")
	err := ioutil.WriteFile(jestConfig, []byte(content), 0644)
	return err
}

func configureEsLint(repo *Repository) error {
	const content = `module.exports = {
  env: {
    browser: true,
    es2021: true,
    node: true,
    jest: true
  },
  extends: [
    'plugin:react/recommended',
    'standard'
  ],
  parser: '@typescript-eslint/parser',
  parserOptions: {
    ecmaFeatures: {
      jsx: true
    },
    ecmaVersion: 12,
    sourceType: 'module'
  },
  plugins: [
    'react',
    '@typescript-eslint'
  ],
  rules: {
  },
  settings: {
    react: {
      version: 'detect'
    }
  }
}
`
	eslintrc := path.Join(repo.RootDir, ".eslintrc.js")
	err := ioutil.WriteFile(eslintrc, []byte(content), 0644)
	return err
}
