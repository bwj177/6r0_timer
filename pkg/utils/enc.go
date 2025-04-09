package utils

import (
	"crypto/md5"
	"encoding/hex"
)

const secret = "6r0-Timer"

func EncryptPassword(oPassword string) string {
	h := md5.New()
	h.Write([]byte(secret))
	h.Write([]byte(oPassword))
	return hex.EncodeToString(h.Sum(nil))
}
