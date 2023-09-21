package app

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"

	"github.com/americanas-go/config"
	"github.com/americanas-go/log"
	"github.com/pkg/errors"
)

var (
	spaceRegex = regexp.MustCompile(`\s*(.*)\n`)
)

func FileNotFound(path string) error {
	return errors.Errorf("file '%s' not found", path)
}

type Loader struct {
	regexCache      *RegexCache
	jsonPathCache   *JSONPathCache
	scenarioHandler *ScenarioHandler
}

func NewLoader(regexCache *RegexCache, jsonPathCache *JSONPathCache, scenarioHandler *ScenarioHandler) *Loader {
	return &Loader{regexCache, jsonPathCache, scenarioHandler}
}

func (loader *Loader) GetMappings() (Mappings, error) {
	mappingsPath := config.String("loader.path.mapping")
	responsesPath := config.String("loader.path.response")

	mappings := make(Mappings)
	err := loader.loadMappings(mappingsPath, responsesPath, mappings)
	if err != nil {
		return mappings, err
	}

	return mappings, nil
}

func (loader *Loader) loadMappings(mappingsPath string, responsesPath string, mappings Mappings) error {
	err := filepath.WalkDir(
		mappingsPath,
		func(path string, d fs.DirEntry, err error) error {
			if d != nil && !d.IsDir() {
				log.Tracef("reading file '%s'", path)
				mapping, err := loader.decodeFile(path)
				if err != nil {
					return err
				}

				if mapping.Response.BodyFile != "" {
					bodyContent, err := loadFile(filepath.Join(responsesPath, mapping.Response.BodyFile))
					if err != nil {
						return errors.Wrapf(err, "error loading response body file for mapping file [ %s ]", path)
					}
					mapping.Response.Body = spaceRegex.ReplaceAllString(string(bodyContent), "$1")
				}

				err = loader.regexCache.AddFromMapping(mapping)
				if err != nil {
					return errors.Wrapf(err, "error adding mapping from file [ %s ]", path)
				}

				err = loader.jsonPathCache.AddExpressions(mapping.Request.Body.JsonPath)
				if err != nil {
					return errors.Wrapf(err, "error adding mapping from file [ %s ]", path)
				}

				if mapping.Scenario != nil {
					loader.scenarioHandler.AddScenario(mapping)
				} else {
					err = mappings.Put(mapping)
					if err != nil {
						return errors.Wrapf(err, "error adding mapping from file [ %s ]", path)
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

func (*Loader) decodeFile(path string) (Mapping, error) {
	content, err := loadFile(path)
	if err != nil {
		return Mapping{}, err
	}

	var m Mapping
	err = json.Unmarshal(content, &m)
	if err != nil {
		return Mapping{}, err
	}

	return m, nil
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
