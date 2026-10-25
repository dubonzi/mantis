package app

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/americanas-go/log"
	"github.com/pkg/errors"
)

func FileNotFound(path string) error {
	return errors.Errorf("file '%s' not found", path)
}

type LoaderConfig struct {
	MappingPath  string
	ResponsePath string
}

type Loader struct {
	config          LoaderConfig
	regexCache      *RegexCache
	jsonPathCache   *JSONPathCache
	scenarioHandler *ScenarioHandler
}

func NewLoader(config LoaderConfig, regexCache *RegexCache, jsonPathCache *JSONPathCache, scenarioHandler *ScenarioHandler) *Loader {
	return &Loader{config, regexCache, jsonPathCache, scenarioHandler}
}

func (loader *Loader) GetMappings() (Mappings, error) {
	mappings := make(Mappings)
	err := loader.loadMappings(loader.config.MappingPath, loader.config.ResponsePath, mappings)
	if err != nil {
		return mappings, err
	}

	return mappings, nil
}

func (loader *Loader) loadMappings(mappingsPath string, responsesPath string, mappings Mappings) error {
	err := filepath.WalkDir(
		mappingsPath,
		func(filePath string, d fs.DirEntry, err error) error {
			if err != nil {
				return errors.Wrapf(err, "error walking mapping path %q", filePath)
			}
			if d != nil && !d.IsDir() {
				log.Debugf("reading file '%s'", filePath)
				loaded, err := loader.decodeMapping(filePath)
				if err != nil {
					return err
				}

				for _, mapping := range loaded {
					err := loader.processMapping(&mapping, filePath, responsesPath)
					if err != nil {
						return errors.Wrapf(err, "error processing file [ %s ]", filePath)
					}

					if mapping.Scenario != nil {
						loader.scenarioHandler.AddScenario(mapping)
					} else {
						err = mappings.Put(mapping)
						if err != nil {
							return errors.Wrapf(err, "error adding mapping from file [ %s ]", filePath)
						}
					}

				}
			}
			return nil
		},
	)

	if err != nil {
		return err
	}

	err = loader.scenarioHandler.ValidateScenarioStates()
	if err != nil {
		return errors.Wrapf(err, "invalid scenario states")
	}

	log.Info("mappings loaded successfully")
	return nil
}

func (*Loader) decodeMapping(path string) ([]Mapping, error) {
	content, err := loadFile(path)
	if err != nil {
		return nil, err
	}

	var mappings []Mapping
	err = json.Unmarshal(content, &mappings)
	if err != nil {
		var m Mapping
		err = json.Unmarshal(content, &m)
		if err != nil {
			return nil, err
		}
		mappings = append(mappings, m)
	}

	return mappings, nil
}

func loadFile(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, FileNotFound(path)
		}
		return nil, errors.Wrapf(err, "error reading file '%s'", path)
	}
	return content, nil
}

func (loader Loader) processMapping(mapping *Mapping, filePath, responsesPath string) error {
	if mapping.Response.BodyFile != "" {
		bodyContent, err := loadFile(filepath.Join(responsesPath, mapping.Response.BodyFile))
		if err != nil {
			return errors.Wrap(err, "error loading response body file for mapping")
		}
		var compacted bytes.Buffer
		if err := json.Compact(&compacted, bodyContent); err == nil {
			bodyContent = compacted.Bytes()
		}
		mapping.Response.Body = string(bodyContent)
	}

	err := loader.regexCache.AddFromMapping(*mapping)
	if err != nil {
		return errors.Wrap(err, "error compiling regex patterns")
	}

	err = loader.jsonPathCache.AddExpressions(mapping.Request.Body.JsonPath)
	if err != nil {
		return errors.Wrap(err, "error parsing jsonpath expressions")
	}

	mapping.CalcMaxScoreAndCost()
	mapping.FilePath = filePath

	return nil
}
