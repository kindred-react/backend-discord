package handlers

import (
	"errors"
)

func parseInt(s string) (int, error) {
	if s == "" {
		return 0, errors.New("empty string")
	}
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, errors.New("invalid character")
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}
