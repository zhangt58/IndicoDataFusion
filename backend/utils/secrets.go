package utils

import (
	"fmt"

	"github.com/zalando/go-keyring"
)

const keyringService = "IndicoDataFusion:API-Tokens"

func SetAPITokenSecret(name, token string) error {
	if name == "" {
		return fmt.Errorf("token name cannot be empty")
	}
	return keyring.Set(keyringService, name, token)
}

func GetAPITokenSecret(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("token name cannot be empty")
	}
	return keyring.Get(keyringService, name)
}

func DeleteAPITokenSecret(name string) error {
	if name == "" {
		return fmt.Errorf("token name cannot be empty")
	}
	return keyring.Delete(keyringService, name)
}

// IsKeyringNotFound returns true when err represents a "secret not found" error
// from the go-keyring library (keyring.ErrNotFound).
func IsKeyringNotFound(err error) bool {
	return err == keyring.ErrNotFound
}
