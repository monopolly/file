package file

import (
	"encoding/base64"
)

func Base64(b []byte) (res string) {
	return base64.RawStdEncoding.EncodeToString(b)
}
func Unbase64(b string) (res []byte) {
	res, _ = base64.RawStdEncoding.DecodeString(b)
	return
}
