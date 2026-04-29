package file

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/bkaradzic/go-lz4"
	"github.com/klauspost/pgzip"
	"github.com/shamaton/msgpack/v3"
)

func Uncompress(b []byte, n any) error {
	r, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer r.Close()

	body, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	return msgpack.Unmarshal(body, n)
}

func Compress(v any) ([]byte, error) {
	body, err := msgpack.Marshal(v)
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer

	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(body); err != nil {
		_ = gz.Close()
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func Gzip(body []byte) ([]byte, error) {
	var b bytes.Buffer

	gz := pgzip.NewWriter(&b)
	if _, err := gz.Write(body); err != nil {
		_ = gz.Close()
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func UnGzip(b []byte) ([]byte, error) {
	r, err := pgzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer r.Close()

	return io.ReadAll(r)
}

func LZ4(in []byte) ([]byte, error) {
	return lz4.Encode(nil, in)
}

func UnLZ4(in []byte) ([]byte, error) {
	return lz4.Decode(nil, in)
}

func Zip(to string, files ...string) error {
	archive, err := os.Create(to)
	if err != nil {
		return err
	}
	defer archive.Close()

	zipWriter := zip.NewWriter(archive)
	defer zipWriter.Close()

	for _, filename := range files {
		if err := addFileToZip(zipWriter, filename); err != nil {
			return err
		}
	}

	return nil
}

func addFileToZip(zipWriter *zip.Writer, filename string) error {
	info, err := os.Stat(filename)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return nil
	}

	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Name = filepath.ToSlash(filename)
	header.Method = zip.Deflate

	w, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, f)
	return err
}

// Unzip reads only files, not directories.
func Unzip(zipfile string, f func(name string, body []byte)) error {
	if f == nil {
		return errors.New("callback cannot be nil")
	}

	archive, err := zip.OpenReader(zipfile)
	if err != nil {
		return err
	}
	defer archive.Close()

	for _, af := range archive.File {
		if af.FileInfo().IsDir() {
			continue
		}

		body, err := readZipFile(af)
		if err != nil {
			return err
		}

		f(af.Name, body)
	}

	return nil
}

func readZipFile(af *zip.File) ([]byte, error) {
	fileInArchive, err := af.Open()
	if err != nil {
		return nil, err
	}
	defer fileInArchive.Close()

	return io.ReadAll(fileInArchive)
}
