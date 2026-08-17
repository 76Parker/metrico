package signature

import (
	"crypto/sha256"
	"encoding/hex"
)

const HeaderName = "HashSHA256"

func Sum(body []byte, key string) string {
	hash := sha256.New()
	_, _ = hash.Write(body)
	_, _ = hash.Write([]byte(key))

	return hex.EncodeToString(hash.Sum(nil))
}
