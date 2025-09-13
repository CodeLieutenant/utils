package utils

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func FindFile(tb testing.TB, fileName string) string {
	tb.Helper()
	workingDir := WorkingDir(tb)
	root := ProjectRootDir(tb)

	for entries, err := os.ReadDir(workingDir); err == nil; {
		for _, entry := range entries {
			if entry.Name() == fileName {
				return workingDir
			}
		}

		if workingDir == root {
			tb.Error("got to he project root, file not found")
			tb.FailNow()
		}

		workingDir, err = GetAbsolutePath(filepath.Join(workingDir, ".."))
		if err != nil {
			tb.Errorf("failed to get absolute path from %s", filepath.Join(workingDir, ".."))
			tb.FailNow()
		}

		entries, err = os.ReadDir(workingDir)
	}

	tb.Errorf("failed to find file %s", fileName)
	tb.FailNow()

	return ""
}

var (
	projectRootDirCache   = make(map[string]string, 10)
	projectRootDirCacheMu sync.RWMutex
)

const gomod = "go.mod"

func ProjectRootDir(tb testing.TB) string {
	tb.Helper()
	originalWorkingDir := WorkingDir(tb)
	workingDir := originalWorkingDir

	projectRootDirCacheMu.RLock()
	if dir, ok := projectRootDirCache[originalWorkingDir]; ok {
		projectRootDirCacheMu.RUnlock()

		return dir
	}
	projectRootDirCacheMu.RUnlock()

	for entries, err := os.ReadDir(workingDir); err == nil; {
		for _, entry := range entries {
			if entry.Name() == gomod {
				projectRootDirCacheMu.Lock()
				projectRootDirCache[originalWorkingDir] = workingDir
				projectRootDirCacheMu.Unlock()

				return workingDir
			}
		}

		if workingDir == "/" {
			tb.Error("got to FS Root, file not found")
			tb.FailNow()
		}

		workingDir, err = GetAbsolutePath(filepath.Join(workingDir, ".."))
		if err != nil {
			tb.Errorf("failed to get absolute path from %s", filepath.Join(workingDir, ".."))
			tb.FailNow()
		}

		entries, err = os.ReadDir(workingDir)
	}

	tb.Errorf("%s not found", gomod)
	tb.FailNow()

	return ""
}

func WorkingDir(tb testing.TB) string {
	tb.Helper()
	wd, err := os.Getwd()
	if err != nil {
		tb.Error(err)
		tb.FailNow()
	}

	return wd
}
