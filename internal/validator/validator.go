package validator

import (
	"fmt"
	"strings"
)

var (
	ErrEmptyKey   = fmt.Errorf("key cannot be empty")
	ErrInvalidKey = fmt.Errorf("key contains invalid characters")
)

func ValidateKey(key string) error {
	if key == "" {
		return ErrEmptyKey
	}
	if strings.Contains(key, "=") {
		return ErrInvalidKey
	}
	return nil
}

func ValidateValue(value string) error {
	return nil
}
