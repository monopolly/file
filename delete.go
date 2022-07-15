package file

import (
	"os"
)

func Delete(filename string) error {
	return os.Remove(filename)
}
