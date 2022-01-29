package app

import (
	"crypto/sha256"
	"encoding/base64"

	"github.com/DrGermanius/Shortener/internal/app/config"
)

func ShortLink(l []byte) string {
	sha := sha256.New()
	sha.Write(l)

	s := base64.URLEncoding.EncodeToString(sha.Sum(nil)[:6])
	return s
}

func FullLink(s string) string {
	return config.Config().BaseURL + "/" + s
}
