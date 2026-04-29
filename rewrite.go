package file

import (
	"os"
	"path/filepath"
	"strings"
)

func Rewrite(filename string, body []byte) error {
	return os.WriteFile(filename, body, 0o666)
}

func Exists(name string) bool {
	_, err := os.Stat(name)
	return err == nil || !os.IsNotExist(err)
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
	return strings.TrimPrefix(filepath.Ext(name), ".")
}

func Filename(name string) string {
	return filepath.Base(name)
}

func SizeE(name string) (size int64, err error) {
	p, err := Info(name)
	if err != nil {
		return
	}
	return p.Size(), nil
}
