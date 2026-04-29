package file

import (
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
	_ = os.MkdirAll(volume, 0o777)
	for _, x := range subdirs {
		_ = os.MkdirAll(path.Join(volume, x), 0o777)
	}
}

func Dir(volume string, subdirs ...string) {
	CreateDirectory(volume, subdirs...)
}

func Save(filename string, body []byte) (err error) {
	dir, _, full := Path(filename)
	if err = os.MkdirAll(dir, 0o777); err != nil {
		return err
	}
	return os.WriteFile(full, body, 0o666)
}

func SaveP(body []byte, filename ...string) (err error) {
	return Save(filepath.Join(filename...), body)
}

func Json(filename string, body any) (err error) {
	dir, _, full := Path(filename)
	if err = os.MkdirAll(dir, 0o777); err != nil {
		return err
	}
	data, err := ffjson.Marshal(body)
	if err != nil {
		return err
	}
	return os.WriteFile(full, data, 0o666)
}

func Msgpack(filename string, body any) (err error) {
	dir, _, full := Path(filename)
	if err = os.MkdirAll(dir, 0o777); err != nil {
		return err
	}
	data, err := msgpack.Encode(body)
	if err != nil {
		return err
	}
	return os.WriteFile(full, data, 0o666)
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
	_, _ = a.Write(append(p, '\n'))
}

func Writer(filename string) (f writer) {
	_ = os.Remove(filename)
	outbase, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o666)
	if err != nil {
		panic(err)
	}
	f = writer{outbase}
	return
}
