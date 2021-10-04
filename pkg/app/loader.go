package app

type Loader interface {
	GetMappings() Mappings
}

type FileLoader struct{}

func NewFileLoader() Loader {
	return &FileLoader{}
}

func (f *FileLoader) GetMappings() Mappings {
	return make(Mappings)
}
