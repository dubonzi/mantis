package app

import (
	"log"
	"os"
)

type Loader interface {
	GetMappings() Mappings
}

type FileLoader struct{}

func NewFileLoader() Loader {
	return &FileLoader{}
}

func (f *FileLoader) GetMappings() Mappings {

	dirs, err := os.ReadDir("files/mapping")
	if err != nil {
		log.Println(err)
		return nil
	}

	for _, entry := range dirs {
		log.Println(entry.Name())
	}
	return nil
}
