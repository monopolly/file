package file

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"

	"github.com/bkaradzic/go-lz4"
	"github.com/klauspost/pgzip"
	"github.com/shamaton/msgpack"
)

func Uncompress(b []byte, n interface{}) (err error) {
	rdata := bytes.NewReader(b)
	r, err := gzip.NewReader(rdata)
	if err != nil {
		return
	}
	s, err := ioutil.ReadAll(r)
	if err != nil {
		return
	}
	err = msgpack.Decode(s, &n)
	return
}

func Compress(a interface{}) (res []byte) {
	body, _ := msgpack.Encode(a)
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	gz.Write(body)
	gz.Flush()
	gz.Close()
	return b.Bytes()
}

func Gzip(body []byte) (res []byte) {
	var b bytes.Buffer
	gz := pgzip.NewWriter(&b)
	gz.Write(body)
	gz.Flush()
	gz.Close()
	return b.Bytes()
}

func UnGzip(b []byte) (res []byte, err error) {
	rdata := bytes.NewReader(b)
	r, err := pgzip.NewReader(rdata)
	if err != nil {
		return
	}
	return ioutil.ReadAll(r)
}

func LZ4(in []byte) (out []byte) {
	out, _ = lz4.Encode(nil, in)
	return
}

func UnLZ4(in []byte) (out []byte) {
	out, _ = lz4.Decode(nil, in)
	return
}

func Zip(to string, files ...string) (err error) {
	archive, err := os.Create(to)
	if err != nil {
		return
	}
	defer archive.Close()
	zipWriter := zip.NewWriter(archive)

	//files
	for _, filename := range files {
		f, err := os.Open(filename)
		if err != nil {
			continue
		}
		defer f.Close()
		w, err := zipWriter.Create(filename)
		if err != nil {
			continue
		}
		if _, err := io.Copy(w, f); err != nil {
			continue
		}
	}

	zipWriter.Close()
	return

}

// only files
// no dirs
func Unzip(zipfile string, f func(name string, body []byte)) (err error) {
	archive, err := zip.OpenReader(zipfile)
	if err != nil {
		return
	}
	defer archive.Close()

	for _, af := range archive.File {

		fileInArchive, err2 := af.Open()
		if err2 != nil {
			continue
		}

		var b bytes.Buffer
		if _, err := io.Copy(&b, fileInArchive); err != nil {
			continue
		}

		f(af.Name, b.Bytes())

		err = fileInArchive.Close()

	}

	return
}
