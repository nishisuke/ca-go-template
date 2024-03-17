package gen

import (
	"os"
	"path/filepath"

	"golang.org/x/tools/imports"
)

func mkdir(dir ...string) (string, error) {
	filePath := filepath.Join(dir...)
	if err := os.MkdirAll(filePath, 0750); err != nil {
		return "", err
	}
	return filePath, nil
}

func write(fileName string, data []byte) error {
	imported, err := imports.Process(fileName, data, nil)
	if err != nil {
		return err
	}

	err = os.WriteFile(fileName, imported, 0o600)
	if err != nil {
		return err
	}
	return nil
}
