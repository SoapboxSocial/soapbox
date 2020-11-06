package images

import (
	"io/ioutil"
	"os"
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
	return ib.store(ib.path, bytes)
}

func (ib *Backend) StoreGroupPhoto(bytes []byte) (string, error) {
	return ib.store(ib.path+"/groups", bytes)
}

func (ib *Backend) Remove(name string) error {
	return os.Remove(ib.path + "/" + name)
}

func (ib *Backend) store(path string, bytes []byte) (string, error) {
	file, err := ioutil.TempFile(path, "*.png")
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
