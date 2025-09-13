package utils

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/joho/godotenv"
)

const (
	EnvDev   = "development"
	EnvProd  = "production"
	EnvStage = "staging"
)

type EnvProvider interface {
	Get(key string) (string, bool)
	Set(key, value string)
}

type (
	OSEnvProvider struct {
		prefix string
	}

	Env struct {
		EnvProvider
	}
)

const (
	baseEnvFile = ".env"
)

func NewEnv(disableDotEnv bool, prefix ...string) Env {
	if !disableDotEnv {
		MustLoadEnv()
	}

	p := ""
	if len(prefix) > 0 {
		p = prefix[0]
	}

	return Env{
		EnvProvider: OSEnvProvider{
			prefix: p,
		},
	}
}

func MustLoadEnv(basePath ...string) {
	env := baseEnvFile
	if len(basePath) > 0 {
		env = filepath.Join(basePath[0], baseEnvFile)
	}

	if !FileExists(env) {
		panic("dotenv file not found: " + env)
	}

	if err := godotenv.Load(env); err != nil {
		// Removed slog.Error to avoid race conditions in parallel tests
		return
	}

	// Removed slog.Debug to avoid race conditions in parallel tests
}

func (p OSEnvProvider) Get(key string) (string, bool) {
	return os.LookupEnv(p.prefix + key)
}

func (p OSEnvProvider) Set(key, value string) {
	if err := os.Setenv(p.prefix+key, value); err != nil {
		// Removed slog.Debug to avoid race conditions in parallel tests
		panic(err)
	}
}

func GetStringsEnv(e Env, key string, defaultValue []string) []string {
	value, exists := e.Get(key)
	if !exists {
		return defaultValue
	}

	vals := make([]string, 0)

	for v := range strings.SplitSeq(value, ",") {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}

		vals = append(vals, v)
	}

	return vals
}

// GetEnv Gets an environment variable or returns a default value
func GetEnv(e Env, key, defaultValue string) string {
	value, exists := e.Get(key)
	if !exists {
		return defaultValue
	}

	return value
}

func GetFloatEnv[T float32 | float64](e Env, key string, defaultValue T) T {
	value, exists := e.Get(key)
	if !exists {
		return defaultValue
	}

	var v T
	parsed, err := strconv.ParseFloat(value, int(unsafe.Sizeof(v)*8))
	if err == nil {
		return T(parsed)
	}

	panic("invalid float value: " + value + " " + err.Error())
}

func GetIntEnv[T int | int8 | int16 | int32 | int64](e Env, key string, defaultValue T) T {
	value, exists := e.Get(key)
	if !exists {
		return defaultValue
	}

	var v T

	parsed, err := strconv.ParseInt(value, 10, int(unsafe.Sizeof(v)*8))
	if err == nil {
		return T(parsed)
	}

	panic("invalid integer value: " + value + " " + err.Error())
}

func GetUintEnv[T uint | uint8 | uint16 | uint32 | uint64](e Env, key string, defaultValue T) T {
	value, exists := e.Get(key)
	if !exists {
		return defaultValue
	}

	var v T
	parsed, err := strconv.ParseUint(value, 10, int(unsafe.Sizeof(v)*8))
	if err == nil {
		return T(parsed)
	}

	panic("invalid unsigned value: " + value + " " + err.Error())
}

func GetBoolEnv(e Env, key string, defaultValue bool) bool {
	value, exists := e.Get(key)
	if !exists {
		return defaultValue
	}

	parsed, err := strconv.ParseBool(value)
	if err == nil {
		return parsed
	}

	panic("invalid boolean value: " + value + " " + err.Error())
}

// GetDurationEnv Gets a duration from environment variable or returns default
func GetDurationEnv(e Env, key string, defaultValue time.Duration) time.Duration {
	value, exists := e.Get(key)

	if !exists {
		return defaultValue
	}

	if parsed, err := time.ParseDuration(value); err == nil {
		return parsed
	}

	// Try parsing as seconds if not a duration string
	seconds, err := strconv.ParseInt(value, 10, 64)
	if err == nil {
		return time.Duration(seconds) * time.Second
	}

	panic("invalid duration value: " + value + " " + err.Error())
}

func GetKeyEnv(e Env, k string, defaults []byte) []byte {
	value, exists := e.Get(k)
	if !exists {
		return defaults
	}

	if value == "" {
		return defaults
	}

	parsed, err := ParseKey(value)
	if err != nil {
		panic("invalid key: " + value + " " + err.Error())
	}

	return parsed
}
