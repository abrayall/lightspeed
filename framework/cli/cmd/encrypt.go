package cmd

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"lightspeed/core/lib/properties"
	"lightspeed/core/lib/ui"
)

var encryptKeys = map[string][]string{
	"alpha":   {"f7a2c9e1", "d4b8063f", "5e2a1c9d", "7b4e8f0a", "3c6d9e2f", "5a8b1c4d", "7e0f3a6b", "9c2d5e8f"},
	"beta":    {"3e9f2a6c", "8d1b4e7f", "0a3c6d9e", "2f5a8b1c", "4d7e0f3a", "6b9c2d5e", "8f1a4b7c", "0d3e6f9a"},
	"gamma":   {"b4e7f0a3", "c6d9e2f5", "a8b1c4d7", "e0f3a6b9", "c2d5e8f1", "a4b7c0d3", "e6f93e9f", "2a6c8d1b"},
	"delta":   {"9c2d5e8f", "1a4b7c0d", "3e6f93e9", "f2a6c8d1", "b4e7f0a3", "c6d9e2f5", "a8b1c4d7", "e0f3a6bc"},
	"epsilon": {"1c4d7e0f", "3a6b9c2d", "5e8f1a4b", "7c0d3e6f", "9b4e7f0a", "3c6d9e2f", "5a8b3e9f", "2a6c8d1e"},
	"zeta":    {"d7e0f3a6", "b9c2d5e8", "f1a4b7c0", "d3e6f93e", "9f2a6c8d", "1b4e7f0a", "3c6d9e2f", "5a8b1c4d"},
}

var encryptKeyNames = []string{"alpha", "beta", "gamma", "delta", "epsilon", "zeta"}

var encryptCmd = &cobra.Command{
	Use:   "encrypt [value]",
	Short: "Encrypt a value for use in site.properties",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintHeader(Version)

		plaintext := args[0]

		salt := loadEncryptionSalt()
		if salt == "" {
			ui.PrintError("No encryption key found. Set LIGHTSPEED_KEY or create ~/.lightspeed/key")
			os.Exit(1)
		}

		encrypted, err := encryptValue(plaintext, salt)
		if err != nil {
			ui.PrintError("Encryption failed: %v", err)
			os.Exit(1)
		}

		ui.PrintSuccess("Encrypted value:")
		fmt.Println()
		fmt.Println(encrypted)
		fmt.Println()
	},
}

func loadEncryptionSalt() string {
	if salt := os.Getenv("LIGHTSPEED_KEY"); salt != "" {
		return salt
	}

	if home, err := os.UserHomeDir(); err == nil {
		keyFile := filepath.Join(home, ".lightspeed", "key")
		if data, err := os.ReadFile(keyFile); err == nil {
			if content := strings.TrimSpace(string(data)); content != "" {
				return content
			}
		}
	}

	dir, _ := os.Getwd()
	propsPath := filepath.Join(dir, "site.properties")
	if properties.FileExists(propsPath) {
		props, err := properties.ParseProperties(propsPath)
		if err == nil {
			if key := props.Get("key"); key != "" {
				return deriveKeyFromIdentifier(key)
			}

			if domain := props.Get("domain"); domain != "" {
				return deriveKeyFromIdentifier(domain)
			}

			domains := props.GetList("domains")
			if len(domains) > 0 {
				return deriveKeyFromIdentifier(domains[0])
			}

			if name := props.Get("name"); name != "" {
				return deriveKeyFromIdentifier(name)
			}
		}
	}

	if dir != "" {
		return deriveKeyFromIdentifier(filepath.Base(dir))
	}

	return ""
}

func deriveKeyFromIdentifier(identifier string) string {
	reversed := reverseString(identifier)
	constant := "lightspeed"
	interleaved := ""
	maxLen := len(reversed)
	if len(constant) > maxLen {
		maxLen = len(constant)
	}
	for i := 0; i < maxLen; i++ {
		if i < len(reversed) {
			interleaved += string(reversed[i])
		}
		if i < len(constant) {
			interleaved += string(constant[i])
		}
	}

	transformed := make([]byte, len(interleaved))
	for i := 0; i < len(interleaved); i++ {
		transformed[i] = byte((int(interleaved[i]) + i*7 + 13) % 256)
	}

	hash1 := sha256.Sum256([]byte(fmt.Sprintf("v1:%s:%d", string(transformed), len(identifier))))
	hash1Hex := fmt.Sprintf("%x", hash1)

	half1 := hash1Hex[:32]
	half2 := hash1Hex[32:]
	folded := ""
	for i := 0; i < 32; i++ {
		folded += string(half1[i]) + string(half2[31-i])
	}

	finalHash := sha256.Sum256([]byte(folded + ":" + identifier))
	return fmt.Sprintf("%x", finalHash)
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func encryptValue(plaintext, salt string) (string, error) {
	keyIdx := make([]byte, 1)
	rand.Read(keyIdx)
	keyName := encryptKeyNames[int(keyIdx[0])%len(encryptKeyNames)]
	baseKey := strings.Join(encryptKeys[keyName], "")

	mixedKey := mixKey(baseKey, salt)
	encryptionKey := sha256.Sum256([]byte(mixedKey))

	data := keyName + ":secret:" + plaintext

	block, err := aes.NewCipher(encryptionKey[:])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	iv := make([]byte, 12)
	if _, err := rand.Read(iv); err != nil {
		return "", err
	}

	sealed := gcm.Seal(nil, iv, []byte(data), nil)
	tag := sealed[len(sealed)-16:]
	ciphertext := sealed[:len(sealed)-16]

	combined := append(iv, tag...)
	combined = append(combined, ciphertext...)
	encoded := base64.StdEncoding.EncodeToString(combined)
	rotated := rot13(encoded)
	swapped := swapHalves(rotated)

	return swapped, nil
}

func decryptValue(encrypted, salt string) (string, error) {
	swapped := swapHalves(encrypted)
	rotated := rot13(swapped)

	combined, err := base64.StdEncoding.DecodeString(rotated)
	if err != nil {
		return "", fmt.Errorf("failed to decode: %w", err)
	}

	if len(combined) < 28 {
		return "", fmt.Errorf("invalid encrypted data")
	}

	iv := combined[:12]
	tag := combined[12:28]
	ciphertext := combined[28:]

	// Reconstruct sealed data (ciphertext + tag) for GCM
	sealed := append(ciphertext, tag...)

	// Try each key name until one works
	for _, keyName := range encryptKeyNames {
		baseKey := strings.Join(encryptKeys[keyName], "")
		mixedKey := mixKey(baseKey, salt)
		encryptionKey := sha256.Sum256([]byte(mixedKey))

		block, err := aes.NewCipher(encryptionKey[:])
		if err != nil {
			continue
		}

		gcm, err := cipher.NewGCM(block)
		if err != nil {
			continue
		}

		plaintext, err := gcm.Open(nil, iv, sealed, nil)
		if err != nil {
			continue
		}

		// Plaintext is "keyName:secret:originalValue" - verify key name prefix
		data := string(plaintext)
		if strings.HasPrefix(data, keyName+":") {
			return strings.TrimPrefix(data, keyName+":"), nil
		}
	}

	return "", fmt.Errorf("decryption failed")
}

// resolveEnvironmentValue tries to decrypt a value. If the decrypted result
// starts with "secret:", it was encrypted — strip the marker and return the
// plaintext. Otherwise return the original value unchanged.
func resolveEnvironmentValue(value, salt string) string {
	decrypted, err := decryptValue(value, salt)
	if err != nil {
		return value
	}
	if strings.HasPrefix(decrypted, "secret:") {
		return strings.TrimPrefix(decrypted, "secret:")
	}
	return value
}

func mixKey(key, salt string) string {
	if salt == "" {
		return key
	}

	saltLen := len(salt)
	partLen := (saltLen + 2) / 3

	part1 := salt[:partLen]
	part2 := ""
	part3 := ""
	if partLen < saltLen {
		end := partLen * 2
		if end > saltLen {
			end = saltLen
		}
		part2 = salt[partLen:end]
		if end < saltLen {
			part3 = salt[end:]
		}
	}

	midpoint := len(key) / 2
	return part1 + key[:midpoint] + part2 + key[midpoint:] + part3
}

func rot13(s string) string {
	result := make([]byte, len(s))
	for i, c := range []byte(s) {
		if c >= 'a' && c <= 'z' {
			result[i] = (c-'a'+13)%26 + 'a'
		} else if c >= 'A' && c <= 'Z' {
			result[i] = (c-'A'+13)%26 + 'A'
		} else {
			result[i] = c
		}
	}
	return string(result)
}

func swapHalves(s string) string {
	midpoint := (len(s) + 1) / 2
	return s[midpoint:] + s[:midpoint]
}

func init() {
	rootCmd.AddCommand(encryptCmd)
}
