package file

import (
	"bytes"
	"io"
	"os"
)

// \n counter
func LineCount(f string) (int, error) {
	r, err := os.Open(f)
	if err != nil {
		return 0, err
	}
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)
		switch {
		case err == io.EOF:
			return count, nil
		case err != nil:
			return count, err
		}
	}
}
