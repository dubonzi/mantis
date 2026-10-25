package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/americanas-go/log"
)

func FileNotFound(path string) error {
	return fmt.Errorf("file '%s' not found", path)
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
				return fmt.Errorf("error walking mapping path %q: %w", filePath, err)
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
						return fmt.Errorf("error processing file [ %s ]: %w", filePath, err)
					}

					if mapping.Scenario != nil {
						err = loader.scenarioHandler.AddScenario(mapping)
						if err != nil {
							return fmt.Errorf("error adding mapping from file [ %s ]: %w", filePath, err)
						}
					} else {
						err = mappings.Put(mapping)
						if err != nil {
							return fmt.Errorf("error adding mapping from file [ %s ]: %w", filePath, err)
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
		return fmt.Errorf("invalid scenario states: %w", err)
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
		return nil, fmt.Errorf("error reading file '%s': %w", path, err)
	}
	return content, nil
}

func (loader Loader) processMapping(mapping *Mapping, filePath, responsesPath string) error {
	if mapping.Response.BodyFile != "" {
		bodyContent, err := loadFile(filepath.Join(responsesPath, mapping.Response.BodyFile))
		if err != nil {
			return fmt.Errorf("error loading response body file for mapping: %w", err)
		}
		var compacted bytes.Buffer
		if err := json.Compact(&compacted, bodyContent); err == nil {
			bodyContent = compacted.Bytes()
		}
		mapping.Response.Body = string(bodyContent)
	}

	err := loader.regexCache.AddFromMapping(*mapping)
	if err != nil {
		return fmt.Errorf("error compiling regex patterns: %w", err)
	}

	err = loader.jsonPathCache.AddExpressions(mapping.Request.Body.JsonPath)
	if err != nil {
		return fmt.Errorf("error parsing jsonpath expressions: %w", err)
	}

	mapping.CalcMaxScoreAndCost()
	mapping.FilePath = filePath

	return nil
}
