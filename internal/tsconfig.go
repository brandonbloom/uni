package internal

import (
	"encoding/json"
	"io/ioutil"
)

type TsConfigMetadata struct {
	Extends         string                         `json:"extends,omitempty"`
	CompilerOptions TsConfigCompileOptionsMetadata `json:"compilerOptions,omitempty"`
	Exclude         []string                       `json:"exclude,omitempty"`
	Include         []string                       `json:"include,omitempty"`
	References      []TsConfigReferenceMetadata    `json:"references,omitempty"`
}

type TsConfigCompileOptionsMetadata struct {
	Incremental      bool                `json:"incremental,omitempty"`
	Declaration      bool                `json:"declaration,omitempty"`
	SourceMap        bool                `json:"sourceMap,omitempty"`
	Composite        bool                `json:"composite,omitempty"`
	ImportHelpers    bool                `json:"importHelpers,omitempty"`
	Strict           bool                `json:"strict,omitempty"`
	EsModuleInterop  bool                `json:"esModuleInterop,omitempty"`
	Target           string              `json:"target,omitempty"`
	ModuleResolution string              `json:"moduleResolution,omitempty"`
	Module           string              `json:"module,omitempty"`
	RootDir          string              `json:"rootDir,omitempty"`
	BaseURL          string              `json:"baseUrl,omitempty"`
	Paths            map[string][]string `json:"paths,omitempty"`
	OutDir           string              `json:"outDir,omitempty"`
}

type TsConfigReferenceMetadata struct {
	Path string `json:"path,omitempty"`
}

func WriteTsConfigJSON(metadata TsConfigMetadata, filename string) error {
	return WriteJSON(filename, metadata)
}

func ReadTConfigJSON(filename string) (*TsConfigMetadata, error) {
	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var metadata TsConfigMetadata
	if err := json.Unmarshal(bs, &metadata); err != nil {
		return nil, err
	}
	return &metadata, nil
}
