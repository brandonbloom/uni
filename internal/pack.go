package internal

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type PackResult struct {
	PackagePath string
}

func Pack(repo *Repository, pkg *Package) (PackResult, error) {
	packedDir := path.Join(repo.OutDir, "packed")
	if err := os.MkdirAll(packedDir, 0755); err != nil {
		return PackResult{}, err
	}

	distPath := path.Join(repo.DistDir, pkg.Name)
	metadata, err := ReadPackageJSON(distPath)
	if err != nil {
		return PackResult{}, err
	}

	version := metadata.Version
	if version == "" {
		return PackResult{}, errors.New("version not set in package build")
	}

	strippedName := stripName(pkg.Name)

	packedName := fmt.Sprintf("%s-%s.tgz", strippedName, version)
	packedPath := path.Join(packedDir, packedName)
	packedFile, err := os.Create(packedPath)
	if err != nil {
		return PackResult{}, err
	}
	defer packedFile.Close()

	zw := gzip.NewWriter(packedFile)
	tw := tar.NewWriter(zw)

	filepath.Walk(distPath, func(file string, fi os.FileInfo, error error) error {
		if(error!=nil) {
			return error
		}
		mode := fi.Mode()
		switch {
		case mode.IsDir():
		case mode.IsRegular():
		default:
			return fmt.Errorf("cannot pack irregular file: %q", fi.Name())
		}

		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}

		relpath, err := filepath.Rel(distPath, file)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(path.Join("package", relpath))

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if fi.IsDir() {
			return nil
		}

		data, err := os.Open(file)
		if err != nil {
			return err
		}
		if _, err := io.Copy(tw, data); err != nil {
			_ = data.Close()
			return err
		}
		return data.Close()
	})

	if err := tw.Close(); err != nil {
		return PackResult{}, err
	}
	if err := zw.Close(); err != nil {
		return PackResult{}, err
	}

	resultName := strippedName + ".tgz"
	resultPath := path.Join(packedDir, resultName)
	if err := os.Remove(resultPath); err != nil {
		if !os.IsNotExist(err) {
			return PackResult{}, err
		}
	}
	if err := os.Symlink(packedPath, resultPath); err != nil {
		return PackResult{}, err
	}

	return PackResult{
		PackagePath: packedPath,
	}, nil
}

func stripName(name string) string {
	name = strings.ReplaceAll(name, "@", "")
	name = strings.ReplaceAll(name, "/", "-")
	return name
}
