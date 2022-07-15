package file

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

//читает файл по строчно
func Lines(fi string, limit uint, h func([]byte)) {
	// open a file
	if file, err := os.Open(fi); err == nil {

		// make sure it gets closed
		defer file.Close()

		// create a new scanner and read the file line by line
		scanner := bufio.NewScanner(file)
		var l uint
		for scanner.Scan() {
			h([]byte(scanner.Text()))
			l++
			if l == limit {
				return
			}
		}

		// check for errors
		if err = scanner.Err(); err != nil {
			return
		}

	} else {
		return
	}
}

//читает файл по строчно
func Play(filename string, h func(line string)) {
	// open a file
	if file, err := os.Open(filename); err == nil {

		// make sure it gets closed
		defer file.Close()

		// create a new scanner and read the file line by line
		scanner := bufio.NewScanner(file)
		var l uint
		for scanner.Scan() {
			h(scanner.Text())
			l++
		}

		// check for errors
		if err = scanner.Err(); err != nil {
			return
		}

	} else {
		return
	}
}

//читает файл по строчно
func PlayBytes(filename string, h func(line []byte)) (er error) {
	// open a file
	if file, err := os.Open(filename); err == nil {

		// make sure it gets closed
		defer file.Close()

		// create a new scanner and read the file line by line
		scanner := bufio.NewScanner(file)
		//var l uint
		for scanner.Scan() {
			h([]byte(scanner.Text()))
			//	l++
		}

		// check for errors
		if er = scanner.Err(); er != nil {
			fmt.Println(er)
			return
		}

	} else {
		fmt.Println(err)
	}

	return
}

func PlayStop(filename string, h func(line string) (stop bool)) {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
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
		for pos, _ := range p {
			p[pos] = strings.TrimSpace(p[pos])
		}
		if h(p) {
			return
		}
	}
}

//fieldname = value
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
	// open a file

	if file, err := os.Open(filename); err == nil {

		// make sure it gets closed
		defer file.Close()

		// create a new scanner and read the file line by line
		scanner := bufio.NewScanner(file)
		var count int
		for scanner.Scan() {
			count++
			if count > 20 {
				return
			}
			r := scanner.Text()
			if strings.Index(r, "INSERT INTO") > -1 {
				fmt.Println("got")
				start := strings.Index(r, "(")
				if start == -1 {
					return
				}
				start++
				fmt.Println("has start")
				end := strings.Index(r, ") VALUES")
				if end == -1 {
					return
				}
				fmt.Println("has end")
				fields = strings.Split(r[start:end], ", ")
				return
			}
		}

		// check for errors
		if err = scanner.Err(); err != nil {
			return
		}

	} else {
		return
	}

	return
}

func SQLLines(filename string, h func(lines []string)) {
	// open a file

	if file, err := os.Open(filename); err == nil {

		// make sure it gets closed
		defer file.Close()

		// create a new scanner and read the file line by line
		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			r := scanner.Text()
			if len(r) == 0 || !strings.HasPrefix(r, "(") {
				continue
			}
			r = strings.TrimPrefix(r, "(")
			r = strings.TrimSuffix(r, "),")
			r = strings.ReplaceAll(r, "NULL", "''")
			lines := strings.Split(r, "', ")
			for pos, _ := range lines {
				lines[pos] = strings.ReplaceAll(lines[pos], "'", "")
			}
			h(lines)
		}

		// check for errors
		if err = scanner.Err(); err != nil {
			return
		}

	} else {
		return
	}
}

//считает строчки
func Count(filename string) (count int) {
	// open a file
	if file, err := os.Open(filename); err == nil {
		if err != nil {
			return
		}
		// make sure it gets closed
		defer file.Close()

		// create a new scanner and read the file line by line
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			count++
		}

		// check for errors
		if err = scanner.Err(); err != nil {
			return
		}

	} else {
		return
	}
	return
}
