package file

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"

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
