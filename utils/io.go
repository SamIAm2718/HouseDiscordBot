package utils

import (
	"encoding/gob"
	"os"
	"strings"
)

func WriteGobToDisk(path string, o interface{}) error {

	//check if file exists
	var _, err = os.Stat(path)

	if os.IsNotExist(err) {
		dir := getDir(path)

		err = os.Mkdir(dir, 0755)
		if err != nil {
			return err
		}
	}
	var file, err1 = os.Create(path)
	if err1 != nil {
		return err
	}
	defer file.Close()
	dataEncoder := gob.NewEncoder(file)
	err = dataEncoder.Encode(o)
	if err != nil {
		return err
	}
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