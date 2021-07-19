package internal

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/goccy/go-yaml"
)

type Repository struct {
	ConfigPath   string
	RootDir      string
	OutDir       string
	DistDir      string
	TmpDir       string
	Engines      map[string]string
	IsWorkspace  bool
	Workspace    Workspace
	Packages     map[string]*Package
	Dependencies map[string]*Dependency
	Url          string
	Registry     string
}

type Workspace struct {
	Name    string
	Version string
}

type Dependency struct {
	Name    string
	Version string
}

type Package struct {
	Name         string
	Version      string
	Folder       string
	Public       bool
	Description  string
	Index        string
	Dependencies map[string]*Dependency
	Executables  map[string]*Executable
}

type TsConfig struct {
}

type Executable struct {
	Name       string
	Entrypoint string
}

const DefaultRegistry = "https://registry.npmjs.org/"

func LoadRepository(searchDir string) (*Repository, error) {
	f, err := openConfigFile(searchDir)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var repo Repository
	repo.ConfigPath = f.Name()
	repo.RootDir = path.Dir(repo.ConfigPath)
	repo.OutDir = path.Join(repo.RootDir, "out")
	repo.DistDir = path.Join(repo.OutDir, "dist")
	repo.TmpDir = path.Join(repo.OutDir, "tmp")

	dec := yaml.NewDecoder(f, yaml.Strict())
	var cfg Config
	err = dec.Decode(&cfg)
	if err == io.EOF {
		// TODO: Should entirely empty config files be allowed? Probably not.
		err = nil
	}
	if err != nil {
		return nil, err
	}

	repo.IsWorkspace = len(cfg.Workspace.Version) > 0

	if repo.IsWorkspace {
		repo.Workspace.Name = cfg.Workspace.Name
		repo.Workspace.Version = cfg.Workspace.Version
		if cfg.Dependencies != nil {
			return nil, errors.New("use dependencies inside workspace")
		}
	}

	repo.Engines = make(map[string]string)
	for engineName, engineVersion := range cfg.Engines {
		repo.Engines[engineName] = engineVersion
	}

	repo.Url = cfg.Repository
	repo.Registry = cfg.Registry
	if repo.Registry == "" {
		repo.Registry = DefaultRegistry
	}

	repo.Packages = make(map[string]*Package)
	for packageName, packageConfig := range cfg.Packages {
		if repo.IsWorkspace {
			if len(packageConfig.Index) > 0 {
				return nil, errors.New("use folder instead index inside workspace")
			}
			if packageConfig.Folder == "" {
				return nil, errors.New("use folder for each package in workspace")
			}
			if strings.Contains(packageConfig.Folder, "\\") {
				return nil, errors.New(packageConfig.Folder + " user normal slashes")
			}
			if strings.HasPrefix(packageConfig.Folder, ".") || strings.HasPrefix(packageConfig.Folder, "/") {
				return nil, errors.New(packageConfig.Folder + " must be relative to workspace folder")
			}
			if strings.HasSuffix(packageConfig.Folder, ".") || strings.HasSuffix(packageConfig.Folder, "/") {
				return nil, errors.New(packageConfig.Folder + " folder ends with a invalid char")
			}
		}
		if packageConfig.Dependencies != nil && (!repo.IsWorkspace) {
			return nil, errors.New("package dependencies is supported only inside workspace")
		}
		pkg := &Package{
			Name:        packageName,
			Public:      packageConfig.Public,
			Description: packageConfig.Description,
			Index:       packageConfig.Index,
			Folder:      packageConfig.Folder,
		}
		pkg.Executables = make(map[string]*Executable)
		for executableName, executableEntrypoint := range packageConfig.Executables {
			pkg.Executables[executableName] = &Executable{
				Name:       executableName,
				Entrypoint: executableEntrypoint,
			}
		}
		if repo.IsWorkspace && packageConfig.Dependencies != nil {
			pkg.Dependencies = make(map[string]*Dependency)
			for dependencyName, dependencyVersion := range packageConfig.Dependencies {
				pkg.Dependencies[dependencyName] = &Dependency{
					Name:    dependencyName,
					Version: dependencyVersion,
				}
			}
		}
		repo.Packages[packageName] = pkg
	}

	repo.Dependencies = make(map[string]*Dependency)
	addDependency := func(name, version string) {
		repo.Dependencies[name] = &Dependency{
			Name:    name,
			Version: version,
		}
	}
	for dependencyName, dependencyVersion := range cfg.Dependencies {
		addDependency(dependencyName, dependencyVersion)
	}
	var requiredDeps map[string]string
	if repo.IsWorkspace {
		requiredDeps = requiredDependenciesWorkspace
	} else {
		requiredDeps = requiredDependenciesMix
	}
	for dependencyName, dependencyVersion := range requiredDeps {
		if _, ok := repo.Dependencies[dependencyName]; !ok {
			addDependency(dependencyName, dependencyVersion)
		}
	}

	return &repo, nil
}

const configName = "uni.yml"

var ErrNoConfig = fmt.Errorf("cannot find %s config file", configName)

func openConfigFile(searchDir string) (*os.File, error) {
	for {
		configPath := path.Join(searchDir, configName)
		f, err := os.Open(configPath)
		if os.IsNotExist(err) {
			searchDir = path.Dir(searchDir)
			if len(searchDir) <= 1 {
				return nil, ErrNoConfig
			}
			continue
		}
		return f, err
	}
}
