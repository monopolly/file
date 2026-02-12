package file

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func Rewrite(filename string, body []byte) (err error) {
	ioutil.WriteFile(filename, body, os.ModePerm)
	return
}

func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func Info(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func Size(name string) int64 {
	p, err := Info(name)
	if err != nil {
		return 0
	}
	return p.Size()
}

// file size
func Int(name string) int {
	p, err := Info(name)
	if err != nil {
		return 0
	}
	return int(p.Size())
}

func Ext(name string) (res string) {
	res = filepath.Ext(name)
	return strings.TrimPrefix(res, ".")
}
func Filename(path string) (res string) {
	return filepath.Base(path)
}

func SizeE(name string) (size int64, err error) {
	p, err := Info(name)
	if err != nil {
		return
	}
	return p.Size(), nil
}
