package file

import (
	"io/ioutil"
	"os"
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

func SizeE(name string) (size int64, err error) {
	p, err := Info(name)
	if err != nil {
		return
	}
	return p.Size(), nil
}
