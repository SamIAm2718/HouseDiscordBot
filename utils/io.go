package utils

import (
	"encoding/gob"
	"errors"
	"os"
)

func WriteGobToDisk(path string, name string, o interface{}) error {

	//check if file exists and if not creates a directory for it
	if _, err := os.Stat(path + "/" + name + ".gob"); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	}

	file, err := os.Create(path + "/" + name + ".gob")
	if err != nil {
		return err
	}
	defer file.Close()

	return gob.NewEncoder(file).Encode(o)
}
