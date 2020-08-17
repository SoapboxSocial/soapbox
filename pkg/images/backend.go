package images

import (
	"io/ioutil"
	"mime/multipart"
	"path/filepath"
)

type Backend struct {
	path string
}

func (ib *Backend) Store(multipartFile multipart.File) (string, error) {
	file, err := ioutil.TempFile(ib.path, "*.png")
	if err != nil {
		return "", err
	}

	defer file.Close()

	fileBytes, err := ioutil.ReadAll(multipartFile)
	if err != nil {
		return "", nil
	}

	_, err = file.Write(fileBytes)
	if err != nil {
		return "", nil
	}

	return filepath.Base(file.Name()), nil
}
