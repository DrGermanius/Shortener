package util

import "github.com/DrGermanius/Shortener/internal/app/config"

func FullLink(s string) string {
	return config.Config().BaseURL + "/" + s
}
