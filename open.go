package file

import (
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/shamaton/msgpack"
)

func Open(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func Move(from, to string) error {
	return os.Rename(from, to)
}

func Copy(from, to string) error {
	sourceFileStat, err := os.Stat(from)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return errors.New("source is not a regular file")
	}

	source, err := os.Open(from)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(to)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

func OpenE(filename string) []byte {
	res, _ := os.ReadFile(filename)
	return res
}

func LoadJson(filename string, v any) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func LoadMsgpack(filename string, v any) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return msgpack.Decode(data, v)
}
