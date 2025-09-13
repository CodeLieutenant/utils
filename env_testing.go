package utils

import (
	"maps"
	"path/filepath"
	"sync"
	"testing"

	"github.com/joho/godotenv"
)

type mapEnvProvider struct {
	prefix string
	env    map[string]string
	mu     sync.Mutex
}

func (p *mapEnvProvider) Get(key string) (string, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	v, ok := p.env[p.prefix+key]

	return v, ok
}

func (p *mapEnvProvider) Set(key, value string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.env[p.prefix+key] = value
}

func NewTestEnv(tb testing.TB) Env {
	tb.Helper()

	provider := &mapEnvProvider{
		env: make(map[string]string),
	}

	provider.mu.Lock()
	defer provider.mu.Unlock()
	root := ProjectRootDir(tb)

	files := []string{".env", ".env.local", ".env.test", ".env.testing"}
	for _, file := range files {
		file = filepath.Join(root, file)
		if !FileExists(file) {
			// Removed slog.Debug to avoid race conditions in parallel tests
			continue
		}

		vals, err := godotenv.Read(file)
		if err != nil {
			tb.Fatal(err)
		}

		maps.Copy(provider.env, vals)
	}

	// Removed slog.Info and slog.Warn to avoid race conditions in parallel tests

	return Env{
		EnvProvider: provider,
	}
}
