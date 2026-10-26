package helpers

import (
	"os"
	"strings"
)

func EnforceHTTPS(url string) string {
	strings.ToLower(url)
	if !strings.Contains(url, "http") {
		url = "https://" + url
	}
	return url
}

func RemoveDomainError(url string) bool {
	if url == os.Getenv("DOMAIN") {
		return false
	}
	newURL := strings.ReplaceAll(url, "http://www.", "")
	newURL = strings.ReplaceAll(newURL, "https://www.", "")
	newURL = strings.ReplaceAll(newURL, "http://", "")
	newURL = strings.ReplaceAll(newURL, "https://", "")

	if newURL == os.Getenv("DOMAIN") {
		return false
	}
	return true

}
