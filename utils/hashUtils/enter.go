package hashUtils

import (
	"crypto/md5"
	"encoding/hex"
	"os"
)

func Md5(data []byte) string {
	m := md5.New()
	m.Write(data)
	return hex.EncodeToString(m.Sum(nil))
}
func FileMd5(file string) (h string, err error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}
	h = Md5(data)
	return h, nil
}
