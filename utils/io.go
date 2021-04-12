package utils

import (
	"encoding/gob"
	"errors"
	"os"
	"strings"
)

func WriteGobToDisk(path string, o interface{}) error {
	//check if file exists and if not creates a directory for it
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		dir := getDir(path)

		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	err = gob.NewEncoder(file).Encode(o)

	return err
}

func getDir(s string) string {
	fileStruct := strings.Split(s, "/")

	if len(fileStruct) <= 1 {
		return ""
	} else {
		fileStruct = fileStruct[:len(fileStruct)-1]
		dir := ""
		for i, file := range fileStruct {
			dir += file

			if i != len(fileStruct)-1 {
				dir += "/"
			}
		}

		return dir
	}
}
