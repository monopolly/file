package file

import (
	"bufio"
	"os"
	"path"
	"path/filepath"

	"github.com/pquerna/ffjson/ffjson"
	"github.com/shamaton/msgpack"
)

func Path(filename string) (dir, name, full string) {
	name = path.Base(filename)
	dir = path.Dir(filename)
	full = path.Join(dir, name)
	return
}

func CreateDirectory(volume string, subdirs ...string) {
	os.MkdirAll(volume, os.ModePerm)
	for _, x := range subdirs {
		os.MkdirAll(path.Join(volume, x), os.ModePerm)
	}
}

func Dir(volume string, subdirs ...string) {
	CreateDirectory(volume, subdirs...)
}

func Save(filename string, body []byte) (err error) {
	dir, _, full := Path(filename)
	err = os.MkdirAll(dir, os.ModePerm)
	f, err := os.Create(full)
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()
	_, err = w.Write(body)
	return
}

func SaveP(body []byte, filename ...string) (err error) {
	return Save(filepath.Join(filename...), body)
}

func Json(filename string, body interface{}) (err error) {
	dir, _, full := Path(filename)
	os.MkdirAll(dir, os.ModePerm)
	data, _ := ffjson.Marshal(body)
	f, err := os.Create(full)
	if err != nil {
		return
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()
	_, err = w.Write(data)
	return
}

func Msgpack(filename string, body interface{}) (err error) {
	dir, _, full := Path(filename)
	os.MkdirAll(dir, os.ModePerm)
	data, _ := msgpack.Encode(body)
	f, err := os.Create(full)
	defer f.Close()
	if err != nil {
		return
	}
	w := bufio.NewWriter(f)
	defer w.Flush()
	_, err = w.Write(data)
	return
}

/*
 //decompress
 decompressed := make([]byte, len(toCompress))
 l, err = lz4.UncompressBlock(compressed[:l], decompressed, 0)
 if err != nil {
	 panic(err)
 }
 fmt.Println("\ndecompressed Data:", string(decompressed[:l]))
*/

type writer struct {
	*os.File
}

func (a *writer) WriteLine(p []byte) {
	a.Write(append(p, []byte("\n")...))
}

func Writer(filename string) (f writer) {
	os.Remove(filename)
	outbase, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}
	f = writer{outbase}
	return
}
