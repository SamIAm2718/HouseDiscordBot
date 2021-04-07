package utils

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
)

func ReadJSONFromDisk(path string, o interface{}) error {
	rawData, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(rawData, o)
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

func WriteJSONToDisk(path string, o interface{}) error {
	jsonData, err := json.Marshal(o)
	if err != nil {
		return err
	}

	err = os.WriteFile(path, jsonData, 0666)
	if errors.Is(err, os.ErrNotExist) {
		dir := getDir(path)

		err = os.Mkdir(dir, 0755)
		if err != nil {
			return err
		}

		err = os.WriteFile(path, jsonData, 0666)
		if err != nil {
			return err
		}
	}
	return err
}
