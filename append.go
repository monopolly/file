package file

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
)

func Append(filename string, body []byte) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o666)
	if err != nil {
		return err
	}

	defer f.Close()
	_, err = f.Write(body)
	return err
}

func AppendLine(filename string, line []byte) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o666)
	if err != nil {
		return err
	}

	defer f.Close()
	if _, err := f.Write(line); err != nil {
		return err
	}
	_, err = f.WriteString("\n")
	return err
}

// csv
func Log(filename string, body ...any) (err error) {
	lines := make([]string, 0, len(body))
	for _, line := range body {
		lines = append(lines, fmt.Sprint(line))
	}

	var b bytes.Buffer
	w := csv.NewWriter(&b)
	if err = w.Write(lines); err != nil {
		return err
	}
	w.Flush()
	if err = w.Error(); err != nil {
		return err
	}

	return Append(filename, b.Bytes())
}

func Appends(filename string, list ...[]byte) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o666)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, body := range list {
		if _, err := f.Write(body); err != nil {
			return err
		}
		if _, err := f.WriteString("\n"); err != nil {
			return err
		}
	}
	return nil
}
