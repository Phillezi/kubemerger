package utils

import (
	"os"
	"strings"
)

func GetHome() string {
	home, err := os.UserHomeDir()
	if err == nil {
		return home
	}
	if home := strings.TrimSpace(os.Getenv("HOME")); home != "" {
		return home
	}
	return ""
}

func Expand(in string) string {
	return strings.ReplaceAll(in, "~", GetHome())
}
