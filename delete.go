package file

import (
	"os"
	"path/filepath"
)

func Delete(filename string) error {
	return os.Remove(filename)
}

func DeleteMask(mask string) (err error) {
	files, err := filepath.Glob(mask)
	if err != nil {
		return
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			continue
		}
	}
	return
}
