package app

import (
	"io/fs"
	"os"

	"github.com/americanas-go/log"
)

type Loader interface {
	GetMappings() Mappings
}

type FileLoader struct{}

func NewFileLoader() *FileLoader {
	return &FileLoader{}
}

func (f *FileLoader) GetMappings() Mappings {

	entries, err := os.ReadDir("files/mapping") // TODO: make configurable?
	if err != nil {
		log.Fatal("error reading mapping directory: ", err)
	}

	mps := make(Mappings)

	f.loadMappings(entries, mps)

	return nil
}

func (f *FileLoader) loadMappings(entries []fs.DirEntry, mappings Mappings) {
	for _, entry := range entries {
		if entry.IsDir() {
			nestedDirs, err := os.ReadDir(entry.Name())
			if err != nil {
				log.Fatal("error reading mapping directory: ", err)
			}
			f.loadMappings(nestedDirs, mappings)
		}

	}

}

func (*FileLoader) decodeFile(path string) Mapping {
	return Mapping{}
}
