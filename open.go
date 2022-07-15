package file

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"

	"github.com/shamaton/msgpack"
)

func Open(filename string) (res []byte, err error) {
	res, err = ioutil.ReadFile(filename)
	return
}

func Move(from, to string) (err error) {
	return os.Rename(from, to)
}

func Copy(from, to string) (err error) {
	sourceFileStat, err := os.Stat(from)
	if err != nil {
		return
	}

	if !sourceFileStat.Mode().IsRegular() {
		return
	}

	source, err := os.Open(from)
	if err != nil {
		return
	}
	defer source.Close()

	destination, err := os.Create(to)
	if err != nil {
		return
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return
}

func OpenE(filename string) (res []byte) {
	res, _ = ioutil.ReadFile(filename)
	return
}

func LoadJson(filename string, v interface{}) (err error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	return json.Unmarshal(data, &v)
}

func LoadMsgpack(filename string, v interface{}) (err error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	return msgpack.Decode(data, &v)
}
