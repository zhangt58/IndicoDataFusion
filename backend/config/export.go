package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"

	"github.com/pkg/errors"
	"golang.org/x/crypto/pbkdf2"
)

const (
	// PBKDF2 parameters
	pbkdf2Iterations = 100000
	pbkdf2KeyLen     = 32 // AES-256
	saltLen          = 32
)

// ExportedConfig represents the encrypted configuration export format
type ExportedConfig struct {
	Version string `json:"version"` // format version for future compatibility
	Salt    string `json:"salt"`    // base64-encoded salt for PBKDF2
	Data    string `json:"data"`    // base64-encoded encrypted data
}

// TokenRetriever is a function that retrieves an API token by name from secure storage
type TokenRetriever func(name string) (string, error)

// TokenStorer is a function that stores an API token by name to secure storage
type TokenStorer func(name, token string) error

// ExportConfig encrypts and exports the configuration with revealed API tokens
// The tokenRetriever function is used to fetch actual token values from secure storage (e.g., keyring)
func ExportConfig(cfg *Config, password string, tokenRetriever TokenRetriever) ([]byte, error) {
	if password == "" {
		return nil, errors.New("password cannot be empty")
	}

	// Clone the config and populate API tokens with actual values from secure storage
	exportCfg := &Config{
		ActiveDataSource: cfg.ActiveDataSource,
		Cache:            cfg.Cache,
		DataSources:      cfg.DataSources,
		APITokens:        make([]APITokenEntry, len(cfg.APITokens)),
		ChartSettings:    cfg.ChartSettings,
	}

	// Retrieve actual token values from secure storage
	for i, entry := range cfg.APITokens {
		exportCfg.APITokens[i] = entry
		if tokenRetriever != nil && entry.Name != "" {
			token, err := tokenRetriever(entry.Name)
			if err != nil {
				// If token retrieval fails, keep empty token field
				// This allows partial exports if some tokens are missing
				exportCfg.APITokens[i].Token = ""
			} else {
				exportCfg.APITokens[i].Token = token
			}
		}
	}

	// Serialize config to JSON
	plaintext, err := json.Marshal(exportCfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal config")
	}
	// Generate random salt
	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, errors.Wrap(err, "failed to generate salt")
	}
	// Derive key from password using PBKDF2
	key := pbkdf2.Key([]byte(password), salt, pbkdf2Iterations, pbkdf2KeyLen, sha256.New)
	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cipher")
	}
	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create GCM")
	}
	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, errors.Wrap(err, "failed to generate nonce")
	}
	// Encrypt the data (nonce is prepended automatically by Seal)
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	// Create export structure
	exported := ExportedConfig{
		Version: "1",
		Salt:    base64.StdEncoding.EncodeToString(salt),
		Data:    base64.StdEncoding.EncodeToString(ciphertext),
	}
	// Marshal to JSON for final output
	output, err := json.MarshalIndent(exported, "", "  ")
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal export")
	}
	return output, nil
}

// ImportConfig decrypts and imports the configuration
// The tokenStorer function is used to save actual token values to secure storage (e.g., keyring)
func ImportConfig(data []byte, password string, tokenStorer TokenStorer) (*Config, error) {
	if password == "" {
		return nil, errors.New("password cannot be empty")
	}
	// Parse the exported structure
	var exported ExportedConfig
	if err := json.Unmarshal(data, &exported); err != nil {
		return nil, errors.Wrap(err, "failed to parse export file")
	}
	// Check version
	if exported.Version != "1" {
		return nil, errors.Errorf("unsupported export version: %s", exported.Version)
	}
	// Decode salt
	salt, err := base64.StdEncoding.DecodeString(exported.Salt)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode salt")
	}
	// Decode ciphertext
	ciphertext, err := base64.StdEncoding.DecodeString(exported.Data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode data")
	}
	// Derive key from password using PBKDF2
	key := pbkdf2.Key([]byte(password), salt, pbkdf2Iterations, pbkdf2KeyLen, sha256.New)
	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cipher")
	}
	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create GCM")
	}
	// Extract nonce and encrypted data
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	nonce, encryptedData := ciphertext[:nonceSize], ciphertext[nonceSize:]
	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt (wrong password?)")
	}
	// Unmarshal config
	var cfg Config
	if err := json.Unmarshal(plaintext, &cfg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal config")
	}

	// Store API tokens to secure storage
	if tokenStorer != nil {
		for i, entry := range cfg.APITokens {
			if entry.Name != "" && entry.Token != "" {
				if err := tokenStorer(entry.Name, entry.Token); err != nil {
					// Log error but continue with other tokens
					// Consider collecting errors and returning them
				}
				// Clear the token from the config structure (it's now in keyring)
				cfg.APITokens[i].Token = ""
			}
		}
	}

	return &cfg, nil
}
