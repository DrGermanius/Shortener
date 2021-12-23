package auth

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"errors"

	"github.com/google/uuid"
)

const AuthCookie = "UserID"

var key = []byte("secret")

func CheckSignature(s string) (string, error) {
	data := []byte(s)
	uid := data[:36]

	sign, err := calculateSignature(uid)
	if err != nil {
		return "", err
	}

	if hmac.Equal(sign, data[36:]) {
		return string(uid), nil
	}

	return "", errors.New("invalid Signature")
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
	h := hmac.New(md5.New, key)
	_, err := h.Write(b)
	if err != nil {
		return []byte{}, err
	}

	res := h.Sum(nil)
	return []byte(hex.EncodeToString(res)), nil
}
