package utils

import (
	"crypto/sha256"
	"crypto/sha3"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"strings"

	"golang.org/x/crypto/blake2b"
)

const (
	Base64Prefix = "base64:"
)

var ErrInvalidKey = errors.New("invalid key format")

func ParseKey(key string) ([]byte, error) {
	if key == "" {
		return nil, ErrInvalidKey
	}

	// Check if the key has a base64 prefix
	if after, ok := strings.CutPrefix(key, Base64Prefix); ok {
		// Remove the prefix and decode using base64
		rawKey := after

		return base64.StdEncoding.DecodeString(rawKey)
	}

	// If no prefix, decode using hex
	return hex.DecodeString(key)
}

func ParseHasher(algo string) func() hash.Hash {
	switch algo {
	case "sha256":
		return sha256.New
	case "sha512/256":
		return sha512.New512_256
	case "sha3-256":
		return func() hash.Hash {
			return sha3.New256()
		}
	case "sha3-512":
		return func() hash.Hash {
			return sha3.New512()
		}
	case "blake2b":
		return func() hash.Hash {
			hasher, _ := blake2b.New256(nil)

			return hasher
		}
	default:
		panic(fmt.Sprintf("hasher is invalid: check your config: %v", algo))
	}
}
