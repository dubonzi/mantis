package app

import (
	"encoding/json"
	"io/fs"
	"io/ioutil"
	"path/filepath"

	"github.com/americanas-go/config"
	"github.com/americanas-go/log"
	"github.com/pkg/errors"
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
	rootPath := config.String("loader.path.mappings")
	if rootPath == "" {
		rootPath = "files/mapping"
	}

	mappings := make(Mappings)
	return mappings, f.loadMappings(rootPath, mappings)
}

func (f *FileLoader) loadMappings(rootPath string, mappings Mappings) error {
	err := filepath.WalkDir(
		rootPath,
		func(path string, d fs.DirEntry, err error) error {
			if d != nil && !d.IsDir() {
				log.Tracef("reading file '%s'", path)
				m, err := f.decodeFile(path)
				if err != nil {
					return err
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
	content, err := ioutil.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Mapping{}, FileNotFound(path)
		}
		return Mapping{}, errors.Wrapf(err, "error reading file '%s'", path)
	}

	var m Mapping
	return m, json.Unmarshal(content, &m)
}
