package utils

import (
	"crypto/rand"
	"io/fs"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"unsafe"
)

//nolint:gosec
/*#nosec G103*/
func UnsafeBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

//nolint:gosec
/*#nosec G103*/
func UnsafeString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// GetAbsolutePath Returns absolute path string for a given directory and error if directory doesent exist
func GetAbsolutePath(path string) (string, error) {
	var err error

	if !filepath.IsAbs(path) {
		path, err = filepath.Abs(path)
		if err != nil {
			return "", err
		}

		return path, nil
	}

	return path, err
}

// CreateDirectoryFromFile Provides a way of creating a directory from a path
// Returns created directory path and error if fails
func CreateDirectoryFromFile(path string, perm fs.FileMode) (string, error) {
	p, err := GetAbsolutePath(path)
	if err != nil {
		return "", err
	}

	directory := filepath.Dir(p)

	if err = os.MkdirAll(directory, perm); err != nil {
		return "", err
	}

	return p, nil
}

// CreateFile Creates file for given directory with flags and permissions for directory and file
// Returns file instance and error if it fails
func CreateFile(path string, flags int, dirMode, mode fs.FileMode) (*os.File, error) {
	path, err := CreateDirectoryFromFile(path, dirMode|fs.ModeDir)
	if err != nil {
		return nil, err
	}

	if _, err = os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		var file *os.File
		//#nosec G304
		file, err = os.Create(path)
		if err != nil {
			return nil, err
		}

		if err = file.Chmod(mode); err != nil {
			return nil, err
		}

		if err = file.Close(); err != nil {
			return nil, err
		}
	}

	//#nosec G304
	return os.OpenFile(path, flags, mode)
}

// CreateLogFile Creates write only appendable file with permission 0o744 for directory and file for given path
func CreateLogFile(path string) (*os.File, error) {
	return CreateFile(path, os.O_WRONLY|os.O_APPEND, 0o744, fs.FileMode(0o744)|os.ModeAppend)
}

// FileExists Provides a way of checking if file exists. Returns bool
func FileExists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}

// CreateDirectory Provides a way of creating directory with permissions for a given path
// Returns string path of created directory and error if fails
func CreateDirectory(path string, perm fs.FileMode) (string, error) {
	p, err := GetAbsolutePath(path)
	if err != nil {
		return "", err
	}

	if err = os.MkdirAll(p, perm); err != nil {
		return "", err
	}

	return p, nil
}

const (
	lowercase = "abcdefghijklmnopqrstuvwxyz"
	uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbers   = "0123456789"
	special   = "!@#$%^&*()-_=+[]{}|;:,.<>?"
	allChars  = lowercase + uppercase + numbers + special
)

var charSets = []string{lowercase, uppercase, numbers, special}

// GenerateRandomPassword generates a cryptographically secure random password
// with specified length containing uppercase letters, lowercase letters, numbers, and special characters
func GenerateRandomPassword(length int) (string, error) {
	if length < 8 {
		length = 12 // Default to 12 characters minimum for security
	}

	var password strings.Builder
	password.Grow(length)

	// Ensure at least one character from each set
	for _, charset := range charSets {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		_ = password.WriteByte(charset[randomIndex.Int64()])
	}

	// Fill the rest with random characters from all sets
	for password.Len() < length {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(allChars))))
		if err != nil {
			return "", err
		}
		_ = password.WriteByte(allChars[randomIndex.Int64()])
	}
	// Shuffle the password to avoid predictable patterns
	passwordBytes := UnsafeBytes(password.String())

	for i := len(passwordBytes) - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", err
		}
		passwordBytes[i], passwordBytes[j.Int64()] = passwordBytes[j.Int64()], passwordBytes[i]
	}

	return UnsafeString(passwordBytes), nil
}
