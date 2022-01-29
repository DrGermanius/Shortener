package auth

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"

	"github.com/DrGermanius/Shortener/internal/app"
	"github.com/DrGermanius/Shortener/internal/app/config"

	"github.com/google/uuid"
)

const AuthCookie = "UserID"

func CheckSignature(s string) (string, error) {
	if len(s) < 36 {
		return "", app.ErrInvalidSignature
	}

	data := []byte(s)
	uid := data[:36]

	sign, err := calculateSignature(uid)
	if err != nil {
		return "", err
	}

	if hmac.Equal(sign, data[36:]) {
		return string(uid), nil
	}

	return "", app.ErrInvalidSignature
}

func GetSignature() (string, error) {
	uid := uuid.New()
	uidByte := []byte(uid.String())

	res, err := calculateSignature(uidByte)
	if err != nil {
		return "", err
	}

	return uid.String() + string(res), nil
}

func calculateSignature(b []byte) ([]byte, error) {
	key := []byte(config.Config().AuthKey)
	h := hmac.New(md5.New, key)
	_, err := h.Write(b)
	if err != nil {
		return nil, err
	}

	res := h.Sum(nil)
	return []byte(hex.EncodeToString(res)), nil
}
