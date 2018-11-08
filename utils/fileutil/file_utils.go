package fileutil

import "io/ioutil"

type FileReader struct{}

func (FileReader) Read(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}
