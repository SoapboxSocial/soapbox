package images

import (
	"io/ioutil"
	"path/filepath"
)

type Backend struct {
	path string
}

func NewImagesBackend(path string) *Backend {
	return &Backend{
		path: path,
	}
}

func (ib *Backend) Store(bytes []byte) (string, error) {
	file, err := ioutil.TempFile(ib.path, "*.png")
	if err != nil {
		return "", err
	}

	defer file.Close()

	_, err = file.Write(bytes)
	if err != nil {
		return "", nil
	}

	return filepath.Base(file.Name()), nil
}
