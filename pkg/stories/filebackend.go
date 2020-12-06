package stories

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

type FileBackend struct {
	path string
}

func NewFileBackend(path string) *FileBackend {
	return &FileBackend{path: path}
}

func (fb *FileBackend) Store(bytes []byte) (string, error) {
	file, err := ioutil.TempFile(fb.path, "*.aac")
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

func (fb *FileBackend) Remove(name string) error {
	return os.Remove(fb.path + "/" + name)
}
