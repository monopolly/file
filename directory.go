package file

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

func DeleteDirectory(dir string) error {
	return os.RemoveAll(dir)
}

func Directory(dir string, h func(f os.FileInfo)) {
	if dir == "" {
		dir = "."
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, f := range files {
		info, err := f.Info()
		if err != nil {
			continue
		}
		h(info)
	}
}

/* all files from dir */
func FileList(dir string, ext ...string) (list []string) {
	switch len(ext) > 0 {
	case true:
		Directory(dir, func(f os.FileInfo) {
			if f.IsDir() {
				return
			}
			for _, x := range ext {
				if strings.HasSuffix(f.Name(), x) {
					list = append(list, filepath.Join(dir, f.Name()))
					return
				}
			}
		})
	default:
		Directory(dir, func(f os.FileInfo) {
			if f.IsDir() {
				return
			}
			list = append(list, filepath.Join(dir, f.Name()))
		})
	}
	return
}

/* all files from all dirs */
func Files(dir string, list *[]string, ext ...string) {
	Directory(dir, func(f os.FileInfo) {
		if f.IsDir() {
			Files(path.Join(dir, f.Name()), list, ext...)
			return
		}
		if len(ext) > 0 {
			if strings.HasSuffix(f.Name(), ext[0]) {
				(*list) = append((*list), filepath.Join(dir, f.Name()))
				return
			}
			return
		}
		(*list) = append((*list), filepath.Join(dir, f.Name()))
	})
}
