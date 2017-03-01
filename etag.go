package pttrss

import (
	"crypto/sha1"
	"encoding/base64"

	"regexp"
)

var (
	reg = regexp.MustCompile("=+$")
)

func Etag(content string) string {
	if content == "" {
		return ""
	}

	sha1Bytes := sha1.Sum([]byte(content))
	sha1str := base64.StdEncoding.EncodeToString(sha1Bytes[0:])
	tag := string(reg.ReplaceAll([]byte(sha1str), []byte("")))
	return tag
}
