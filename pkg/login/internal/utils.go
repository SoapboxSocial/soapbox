package internal

import (
	"crypto/rand"
	"fmt"
	"io"
	"regexp"
)

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func ValidateEmail(email string) bool {
	return len(email) < 254 && emailRegex.MatchString(email)
}

var usernameRegex = regexp.MustCompile("^([a-z0-9_]+)*$")

func ValidateUsername(username string) bool {
	return len(username) < 100 && len(username) > 2 && usernameRegex.MatchString(username)
}

func GenerateToken() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", b), nil
}

func GeneratePin() (string, error) {
	max := 6
	b := make([]byte, max)
	n, err := io.ReadAtLeast(rand.Reader, b, max)
	if n != max {
		return "", err
	}

	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}

	return string(b), nil
}

var table = []byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
