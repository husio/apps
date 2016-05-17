package unote

import (
	"crypto/rand"
	"encoding/base64"
)

func generateId() string {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	s := base64.URLEncoding.EncodeToString(b)
	return s[:32]
}
