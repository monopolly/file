package file

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"os"
)

var cryptHeader = []byte("gcm:")

var errInvalidCiphertext = errors.New("invalid ciphertext")

func SaveCrypt(filename, pass string, v any) error {
	body, err := json.Marshal(v)
	if err != nil {
		return err
	}

	body, err = crypt(body, pass)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, body, 0o666)
}

func LoadCrypt(filename, pass string, v any) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	data, err = decrypt(data, pass)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

func Crypt(data []byte, key string) []byte {
	v, err := crypt(data, key)
	if err != nil {
		return nil
	}

	return v
}

func Decrypt(db []byte, key string) []byte {
	data, err := decrypt(db, key)
	if err != nil {
		return nil
	}

	return data
}

func crypt(data []byte, key string) ([]byte, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	out := make([]byte, 0, len(cryptHeader)+len(nonce)+len(data)+gcm.Overhead())
	out = append(out, cryptHeader...)
	out = append(out, nonce...)

	return gcm.Seal(out, nonce, data, cryptHeader), nil
}

func decrypt(db []byte, key string) ([]byte, error) {
	if bytes.HasPrefix(db, cryptHeader) {
		return decryptGCM(db, key)
	}

	return decryptLegacyCFB(db, key)
}

func newGCM(key string) (cipher.AEAD, error) {
	sum := sha256.Sum256([]byte(key))

	block, err := aes.NewCipher(sum[:])
	if err != nil {
		return nil, err
	}

	return cipher.NewGCM(block)
}

func decryptGCM(db []byte, key string) ([]byte, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}

	db = db[len(cryptHeader):]
	if len(db) < gcm.NonceSize() {
		return nil, errInvalidCiphertext
	}

	nonce := db[:gcm.NonceSize()]
	ciphertext := db[gcm.NonceSize():]

	return gcm.Open(nil, nonce, ciphertext, cryptHeader)
}

func decryptLegacyCFB(db []byte, key string) ([]byte, error) {
	block, err := aes.NewCipher(hashTo32Bytes(key))
	if err != nil {
		return nil, err
	}
	if len(db) < aes.BlockSize {
		return nil, errInvalidCiphertext
	}

	decoded := make([]byte, len(db)-aes.BlockSize)
	cfb := cipher.NewCFBDecrypter(block, db[:aes.BlockSize])
	cfb.XORKeyStream(decoded, db[aes.BlockSize:])

	return base64.StdEncoding.AppendDecode(nil, decoded)
}

func hashTo32Bytes(input string) []byte {
	sum := sha256.Sum256([]byte(input))
	key := make([]byte, 32)
	base64.URLEncoding.Encode(key, sum[:24])
	return key
}
