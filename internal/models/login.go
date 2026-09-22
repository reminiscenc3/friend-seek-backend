package models

import (
	"strings"
	"unicode/utf8"
)

// NormalizeLogin brings a login to its canonical form: trimmed and lowercased,
// so that "Oleg" and " oleg " address the same user everywhere.
func NormalizeLogin(login string) (string, error) {
	login = strings.ToLower(strings.TrimSpace(login))

	if login == "" || utf8.RuneCountInString(login) > MaxLoginLength {
		return "", ErrInvalidLogin
	}

	return login, nil
}

// NormalizeLogin canonicalizes the login the request carries, in place.
func (r *LoginRequest) NormalizeLogin() error {
	login, err := NormalizeLogin(r.Login)
	if err != nil {
		return err
	}

	r.Login = login

	return nil
}
