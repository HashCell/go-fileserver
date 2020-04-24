package util

import (
	"os"
	"crypto/sha1"
	"io"
	"encoding/hex"
)

func FileSha1(file *os.File) string {
	_sha1 := sha1.New()
	io.Copy(_sha1, file)
	return hex.EncodeToString(_sha1.Sum(nil))
}