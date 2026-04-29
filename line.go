package file

import (
	"bufio"
	"encoding/csv"
	"io"
	"os"
	"strings"
)

const maxLineSize = 512 * 1024

func lineScanner(r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), maxLineSize)
	return scanner
}

// читает файл по строчно
func Lines(fi string, limit uint, h func([]byte)) {
	file, err := os.Open(fi)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := lineScanner(file)
	var l uint
	for scanner.Scan() {
		h(append([]byte(nil), scanner.Bytes()...))
		l++
		if limit > 0 && l >= limit {
			return
		}
	}
}

// читает файл по строчно
func Play(filename string, h func(line string)) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := lineScanner(file)
	for scanner.Scan() {
		h(scanner.Text())
	}
}

// читает файл по строчно
func PlayBytes(filename string, h func(line []byte)) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := lineScanner(file)
	for scanner.Scan() {
		h(append([]byte(nil), scanner.Bytes()...))
	}
	return scanner.Err()
}

func PlayStop(filename string, h func(line string) (stop bool)) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := lineScanner(file)
	for scanner.Scan() {
		if h(scanner.Text()) {
			return
		}
	}
}

func CSV(file string, delim byte, h func(line []string) (stop bool)) {
	f, err := os.Open(file)
	if err != nil {
		return
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = rune(delim)
	for {
		p, err := r.Read()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return
		}
		for pos := range p {
			p[pos] = strings.TrimSpace(p[pos])
		}
		if h(p) {
			return
		}
	}
}

// fieldname = value
func SQL(filename string, h func(map[string]string)) {
	fields := SQLStruct(filename)
	SQLLines(filename, func(lines []string) {
		h(tosqlstruct(lines, fields))
	})
}

func tosqlstruct(lines []string, fields []string) (res map[string]string) {
	res = make(map[string]string)
	if len(lines) != len(fields) {
		return
	}
	for pos, x := range fields {
		res[x] = lines[pos]
	}
	return
}

func SQLStruct(filename string) (fields []string) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := lineScanner(file)
	var count int
	for scanner.Scan() {
		count++
		if count > 20 {
			return
		}

		r := scanner.Text()
		if !strings.Contains(r, "INSERT INTO") {
			continue
		}

		start := strings.Index(r, "(")
		if start == -1 {
			return
		}
		start++

		end := strings.Index(r, ") VALUES")
		if end == -1 {
			return
		}

		return strings.Split(r[start:end], ", ")
	}

	return
}

func SQLLines(filename string, h func(lines []string)) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := lineScanner(file)
	for scanner.Scan() {
		r := scanner.Text()
		if len(r) == 0 || !strings.HasPrefix(r, "(") {
			continue
		}

		r = strings.TrimPrefix(r, "(")
		r = strings.TrimSuffix(r, "),")
		r = strings.ReplaceAll(r, "NULL", "''")

		lines := strings.Split(r, "', ")
		for pos := range lines {
			lines[pos] = strings.ReplaceAll(lines[pos], "'", "")
		}
		h(lines)
	}
}

// считает строчки
func Count(filename string) (count int) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := lineScanner(file)
	for scanner.Scan() {
		count++
	}
	return
}
