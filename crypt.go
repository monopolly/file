package file

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
)

func SaveCrypt(filename, pass string, v interface{}) (err error) {

	f, err := os.Create(filename)
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()
	body, err := json.Marshal(v)
	body = Crypt(body, pass)
	_, err = w.Write(body)
	if err != nil {
		return
	}

	return
}

func LoadCrypt(filename, pass string, v interface{}) (err error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	data = Decrypt(data, pass)
	err = json.Unmarshal(data, &v)
	return
}

func Crypt(data []byte, key string) (v []byte) {
	block, err := aes.NewCipher(hashTo32Bytes(key))
	if err != nil {
		return
	}
	b := base64.StdEncoding.EncodeToString(data)
	ciphertext := make([]byte, aes.BlockSize+len(b))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))
	return ciphertext
}

func Decrypt(db []byte, key string) (data []byte) {
	block, err := aes.NewCipher(hashTo32Bytes(key))
	if err != nil {
		return
	}
	if len(db) < aes.BlockSize {
		return
	}
	iv := db[:aes.BlockSize]
	db = db[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(db, db)
	data, err = base64.StdEncoding.DecodeString(string(db))
	if err != nil {
		return
	}

	return data
}

func hashTo32Bytes(input string) []byte {
	hasher := sha256.New()
	hasher.Write([]byte(input))
	stringToSHA256 := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return []byte(stringToSHA256[:32])
}
