package app

import (
	"encoding/json"
	"io/fs"
	"io/ioutil"
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
	regexCache    *RegexCache
	jsonPathCache *JSONPathCache
}

func NewLoader(regexCache *RegexCache, jsonPathCache *JSONPathCache) *Loader {
	return &Loader{regexCache, jsonPathCache}
}

func (f *Loader) GetMappings() (Mappings, error) {
	mappingsPath := config.String("loader.path.mapping")
	responsesPath := config.String("loader.path.response")

	mappings := make(Mappings)
	err := f.loadMappings(mappingsPath, responsesPath, mappings)
	if err != nil {
		return mappings, err
	}

	return mappings, nil
}

func (f *Loader) loadMappings(mappingsPath string, responsesPath string, mappings Mappings) error {
	err := filepath.WalkDir(
		mappingsPath,
		func(path string, d fs.DirEntry, err error) error {
			if d != nil && !d.IsDir() {
				log.Tracef("reading file '%s'", path)
				m, err := f.decodeFile(path)
				if err != nil {
					return err
				}

				if m.Response.BodyFile != "" {
					bodyContent, err := loadFile(filepath.Join(responsesPath, m.Response.BodyFile))
					if err != nil {
						return errors.Wrapf(err, "error loading response body file for mapping file [ %s ]", path)
					}
					m.Response.Body = spaceRegex.ReplaceAllString(string(bodyContent), "$1")
				}

				err = f.regexCache.AddFromMapping(&m)
				if err != nil {
					return errors.Wrapf(err, "error adding mapping from file [ %s ]", path)
				}

				err = f.jsonPathCache.AddExpressions(m.Request.Body.JsonPath)
				if err != nil {
					return errors.Wrapf(err, "error adding mapping from file [ %s ]", path)
				}

				err = mappings.Put(&m)
				if err != nil {
					return errors.Wrapf(err, "error adding mapping from file [ %s ]", path)
				}
			}
			return nil
		},
	)

	if err != nil {
		return err
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
	content, err := ioutil.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, FileNotFound(path)
		}
		return nil, errors.Wrapf(err, "error reading file '%s'", path)
	}
	return content, nil
}
