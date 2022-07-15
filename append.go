package file

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"os"
)

func Append(filename string, body []byte) (err error) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return
	}

	defer f.Close()
	_, err = f.Write(body)
	return
}

func AppendLine(filename string, line []byte) (err error) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return
	}

	defer f.Close()
	f.Write(line)
	f.WriteString("\n")
	return
}

//csv
func Log(filename string, body ...interface{}) (err error) {
	lines := []string{}
	for _, line := range body {
		lines = append(lines, fmt.Sprint(line))
	}
	var b bytes.Buffer
	w := csv.NewWriter(&b)
	err = w.Write(lines)
	if err != nil {
		log.Println(err)
	}
	//аерепрпе
	w.Flush()
	return Append(filename, b.Bytes())
}

func Appends(filename string, list ...[]byte) (err error) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return
	}
	defer f.Close()
	for _, body := range list {
		_, err = f.Write(append(body, '\n'))
		if err != nil {
			continue
		}
	}
	return
}
