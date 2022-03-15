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

type Loader interface {
	GetMappings() (Mappings, error)
}

type FileLoader struct{}

func NewFileLoader() *FileLoader {
	return &FileLoader{}
}

func (f *FileLoader) GetMappings() (Mappings, error) {
	mappingsPath := config.String("loader.path.mapping")
	if mappingsPath == "" {
		mappingsPath = "files/mapping"
	}

	responsesPath := config.String("loader.path.response")
	if responsesPath == "" {
		responsesPath = "files/response"
	}

	mappings := make(Mappings)
	return mappings, f.loadMappings(mappingsPath, responsesPath, mappings)
}

func (f *FileLoader) loadMappings(mappingsPath string, responsesPath string, mappings Mappings) error {
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
						return err
					}
					m.Response.Body = spaceRegex.ReplaceAllString(string(bodyContent), "$1")
				}

				err = mappings.Put(m)
				if err != nil {
					return err
				}
			}
			return nil
		},
	)

	if err != nil {
		return err
	}

	log.Info("mappings loaded successfuly")
	return nil
}

func (*FileLoader) decodeFile(path string) (Mapping, error) {
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
