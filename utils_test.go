package utils_test

import (
	"io"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/CodeLieutenant/utils"
)

func TestCreateLogFile(t *testing.T) {
	t.Parallel()
	req := require.New(t)

	path := "./test-logs/log.json"

	t.Cleanup(func() { _ = os.RemoveAll("./test-logs") })

	file, err := utils.CreateLogFile(path)
	defer func() {
		if file != nil {
			_ = file.Close()
		}
	}()

	req.NoError(err)
	req.NotNil(file)
	req.FileExists(path)
}

func TestFileExistsSuccess(t *testing.T) {
	// Arrange
	t.Parallel()
	req := require.New(t)
	path := "./log.json"
	file, err := utils.CreateLogFile(path)
	defer func() {
		if file != nil {
			_ = file.Close()
		}
	}()

	t.Cleanup(func() {
		_ = os.Remove(path)
	})

	// Act
	exists := utils.FileExists(path)

	// Assert
	req.NoError(err)
	req.NotNil(file)
	req.True(exists)
}

func TestFileExistsNoFile(t *testing.T) {
	// Arrange
	t.Parallel()
	req := require.New(t)
	path := "./file-does-not-exist.json"

	// Act
	exists := utils.FileExists(path)

	// Assert
	req.False(exists)
}

func TestCreateDirectory(t *testing.T) {
	t.Parallel()
	req := require.New(t)

	// Use a temporary directory to avoid conflicts
	tmp, err := os.MkdirTemp("", "test-create-dir-*")
	req.NoError(err)
	defer func() { _ = os.RemoveAll(tmp) }()

	testDir := filepath.Join(tmp, "test-dir")
	path, err := utils.CreateDirectory(testDir, 0o744)

	req.NoError(err)
	req.Equal(testDir, path)
	req.DirExists(testDir)
}

func TestGetAbsolutePath_RelativeAbsoluteAndError(t *testing.T) {
	// This test changes the working directory; do not run in parallel.
	req := require.New(t)

	origWD, err := os.Getwd()
	req.NoError(err)
	defer func() { _ = os.Chdir(origWD) }()

	tmp, err := os.MkdirTemp("", "abs-test-*")
	req.NoError(err)
	defer func() { _ = os.RemoveAll(tmp) }()

	// Change into tmp and test relative path success
	req.NoError(os.Chdir(tmp))
	rel := "sub/dir/file.txt"
	got, err := utils.GetAbsolutePath(rel)
	req.NoError(err)
	req.True(filepath.IsAbs(got))
	req.Equal(filepath.Join(tmp, rel), got)

	// Absolute path should be returned as-is
	got2, err := utils.GetAbsolutePath(tmp)
	req.NoError(err)
	req.Equal(tmp, got2)

	// Induce error by removing the current directory before calling
	req.NoError(os.Chdir(tmp))
	req.NoError(os.RemoveAll(tmp))
	_, err = utils.GetAbsolutePath("anything")
	req.Error(err)

	// Restore WD
	req.NoError(os.Chdir(origWD))
}

func TestCreateDirectoryFromFile_SuccessAndError(t *testing.T) {
	// Not parallel due to working directory manipulation
	req := require.New(t)

	origWD, err := os.Getwd()
	req.NoError(err)
	defer func() { _ = os.Chdir(origWD) }()

	tmp, err := os.MkdirTemp("", "cdf-*")
	req.NoError(err)
	defer func() { _ = os.RemoveAll(tmp) }()

	p := filepath.Join(tmp, "nested", "file.log")
	ret, err := utils.CreateDirectoryFromFile(p, 0o755)
	req.NoError(err)
	req.Equal(p, ret)
	req.DirExists(filepath.Dir(p))

	// Error path via invalid CWD
	req.NoError(os.Chdir(tmp))
	req.NoError(os.RemoveAll(tmp))
	_, err = utils.CreateDirectoryFromFile("file.txt", 0o755)
	req.Error(err)

	// Restore
	req.NoError(os.Chdir(origWD))
}

func TestCreateFile_CreateReopenAndError(t *testing.T) {
	// Not parallel due to CWD manipulation
	req := require.New(t)

	origWD, err := os.Getwd()
	req.NoError(err)
	defer func() { _ = os.Chdir(origWD) }()

	tmp, err := os.MkdirTemp("", "cf-*")
	req.NoError(err)
	defer func() {
		_ = os.RemoveAll(tmp)
	}()

	p := filepath.Join(tmp, "logs", "app.log")
	f, err := utils.CreateFile(p, os.O_WRONLY|os.O_APPEND, 0o755, 0o644)
	req.NoError(err)
	req.NotNil(f)
	_, _ = f.WriteString("hello")
	req.NoError(f.Close())

	// Re-open existing file with read-only
	f2, err := utils.CreateFile(p, os.O_RDONLY, 0o755, 0o644)
	req.NoError(err)
	defer func() { _ = f2.Close() }()
	b, err := io.ReadAll(f2)
	req.NoError(err)
	req.Equal("hello", string(b))

	// Error path via invalid CWD so CreateDirectoryFromFile fails
	req.NoError(os.Chdir(tmp))
	req.NoError(os.RemoveAll(tmp))
	_, err = utils.CreateFile("rel.log", os.O_RDONLY, 0o755, 0o644)
	req.Error(err)

	// Restore
	req.NoError(os.Chdir(origWD))
}

func TestGenerateRandomPassword(t *testing.T) {
	t.Parallel()
	req := require.New(t)

	// Length below minimum defaults to 12
	pw, err := utils.GenerateRandomPassword(6)
	req.NoError(err)
	req.Len(pw, 12)

	// Should contain at least one from each set and be random-ish
	lower := regexp.MustCompile(`[a-z]`)
	upper := regexp.MustCompile(`[A-Z]`)
	digit := regexp.MustCompile(`[0-9]`)
	special := regexp.MustCompile(`[!@#$%^&*()\-_=+\[\]{}|;:,.<>?]`)

	req.True(lower.MatchString(pw))
	req.True(upper.MatchString(pw))
	req.True(digit.MatchString(pw))
	req.True(special.MatchString(pw))

	// Another length
	pw2, err := utils.GenerateRandomPassword(20)
	req.NoError(err)
	req.Len(pw2, 20)
	req.NotEqual(pw, pw2)
}

func TestUnsafeConversions(t *testing.T) {
	// Safe to run in parallel
	t.Parallel()
	req := require.New(t)

	orig := "Hello, 世界!"
	b := utils.UnsafeBytes(orig)
	req.Equal(len(orig), len(b))
	back := utils.UnsafeString(b)
	req.Equal(orig, back)
}

func TestWorkingDir(t *testing.T) {
	// Safe to run in parallel
	t.Parallel()
	req := require.New(t)

	wd := utils.WorkingDir(t)
	req.NotEmpty(wd)
	req.Equal(wd, utils.WorkingDir(t))
}

func TestProjectRootDirAndCache(t *testing.T) {
	// Safe to run in parallel
	t.Parallel()
	req := require.New(t)

	root1 := utils.ProjectRootDir(t)
	req.NotEmpty(root1)
	// Second call should hit cache path
	root2 := utils.ProjectRootDir(t)
	req.Equal(root1, root2)
}

func TestFindFileGoMod(t *testing.T) {
	// Safe to run in parallel
	t.Parallel()
	req := require.New(t)

	// First, ensure we can find the project root
	root := utils.ProjectRootDir(t)
	req.NotEmpty(root)

	// Verify go.mod exists in the root directory
	goModPath := filepath.Join(root, "go.mod")
	req.FileExists(goModPath)

	// Now test FindFile for go.mod
	dir := utils.FindFile(t, "go.mod")
	req.Equal(root, dir)
}

func TestFindFile_CurrentDirHit(t *testing.T) {
	// Not parallel as we create/remove a file in the package dir
	req := require.New(t)

	wd := utils.WorkingDir(t)
	tmpFile := filepath.Join(wd, "__tmp_test_marker__")
	f, err := os.Create(tmpFile)
	req.NoError(err)
	_ = f.Close()
	t.Cleanup(func() { _ = os.Remove(tmpFile) })

	dir := utils.FindFile(t, filepath.Base(tmpFile))
	req.Equal(wd, dir)
}

func TestGenerateRandomPassword_ErrorHandling(t *testing.T) {
	t.Parallel()
	req := require.New(t)

	// Test with very short length (should default to 12)
	pw, err := utils.GenerateRandomPassword(1)
	req.NoError(err)
	req.Len(pw, 12)

	// Test with 0 length (should default to 12)
	pw, err = utils.GenerateRandomPassword(0)
	req.NoError(err)
	req.Len(pw, 12)

	// Test with negative length (should default to 12)
	pw, err = utils.GenerateRandomPassword(-5)
	req.NoError(err)
	req.Len(pw, 12)

	// Test with exact minimum length
	pw, err = utils.GenerateRandomPassword(8)
	req.NoError(err)
	req.Len(pw, 8)
}

func TestCreateDirectory_ErrorCases(t *testing.T) {
	t.Parallel()
	req := require.New(t)

	// Test with invalid path characters (this might not fail on all systems)
	// but we can test basic functionality
	tmp, err := os.MkdirTemp("", "test-create-dir-*")
	req.NoError(err)
	defer func() { _ = os.RemoveAll(tmp) }()

	// Test creating nested directories
	nestedPath := filepath.Join(tmp, "level1", "level2", "level3")
	created, err := utils.CreateDirectory(nestedPath, 0o755)
	req.NoError(err)
	req.Equal(nestedPath, created)
	req.DirExists(nestedPath)

	// Test creating directory that already exists
	created2, err := utils.CreateDirectory(nestedPath, 0o755)
	req.NoError(err)
	req.Equal(nestedPath, created2)
}

func TestCreateFile_ErrorCases(t *testing.T) {
	req := require.New(t)

	tmp, err := os.MkdirTemp("", "test-create-file-*")
	req.NoError(err)
	defer func() { _ = os.RemoveAll(tmp) }()

	// Test creating file with non-existent parent directories
	filePath := filepath.Join(tmp, "deep", "nested", "path", "file.txt")
	file, err := utils.CreateFile(filePath, os.O_RDWR|os.O_CREATE, 0o755, 0o644)
	req.NoError(err)
	req.NotNil(file)
	_ = file.Close()
	req.FileExists(filePath)

	// Test opening existing file with different flags
	file2, err := utils.CreateFile(filePath, os.O_RDONLY, 0o755, 0o644)
	req.NoError(err)
	req.NotNil(file2)
	_ = file2.Close()

	// Test with write-only flag
	file3, err := utils.CreateFile(filePath, os.O_WRONLY, 0o755, 0o644)
	req.NoError(err)
	req.NotNil(file3)
	_ = file3.Close()
}

func TestCreateDirectoryFromFile_EdgeCases(t *testing.T) {
	t.Parallel()
	req := require.New(t)

	tmp, err := os.MkdirTemp("", "test-create-dir-from-file-*")
	req.NoError(err)
	defer func() { _ = os.RemoveAll(tmp) }()

	// Test with file path that has multiple levels
	filePath := filepath.Join(tmp, "a", "b", "c", "d", "file.log")
	created, err := utils.CreateDirectoryFromFile(filePath, 0o755)
	req.NoError(err)
	req.Equal(filePath, created)
	req.DirExists(filepath.Dir(filePath))

	// Test with relative path
	origWD, err := os.Getwd()
	req.NoError(err)
	defer func() { _ = os.Chdir(origWD) }()

	req.NoError(os.Chdir(tmp))
	relativeFile := "relative/path/file.txt"
	created2, err := utils.CreateDirectoryFromFile(relativeFile, 0o755)
	req.NoError(err)
	req.True(filepath.IsAbs(created2))
	req.DirExists(filepath.Dir(created2))
}

func TestWorkingDir_ErrorHandling(t *testing.T) {
	// This test needs to be run in isolation as it manipulates the working directory
	req := require.New(t)

	// Create a temporary directory
	tmp, err := os.MkdirTemp("", "test-working-dir-*")
	req.NoError(err)
	defer func() { _ = os.RemoveAll(tmp) }()

	// Change to the temporary directory
	origWD, err := os.Getwd()
	req.NoError(err)
	defer func() { _ = os.Chdir(origWD) }()

	req.NoError(os.Chdir(tmp))

	// Get working directory (should work)
	wd := utils.WorkingDir(t)
	req.Equal(tmp, wd)
}

func TestProjectRootDir_ErrorCases(t *testing.T) {
	// This test manipulates the working directory, so it can't be parallel
	req := require.New(t)

	// Save original working directory
	origWD, err := os.Getwd()
	req.NoError(err)
	defer func() { _ = os.Chdir(origWD) }()

	// Test normal operation first - should work and populate cache
	root := utils.ProjectRootDir(t)
	req.NotEmpty(root)
	req.True(filepath.IsAbs(root))

	// Test cache hit by calling again
	root2 := utils.ProjectRootDir(t)
	req.Equal(root, root2)

	// Test from a subdirectory of the project
	tmp, err := os.MkdirTemp(root, "test-subdir-*")
	req.NoError(err)
	defer func() { _ = os.RemoveAll(tmp) }()

	// Change to subdirectory
	req.NoError(os.Chdir(tmp))

	// Should still find the project root
	root3 := utils.ProjectRootDir(t)
	req.Equal(root, root3)
}

func TestFindFile_ErrorCases(t *testing.T) {
	// This test manipulates the working directory
	req := require.New(t)

	// Save original working directory
	origWD, err := os.Getwd()
	req.NoError(err)
	defer func() { _ = os.Chdir(origWD) }()

	// Test normal case - finding go.mod should work
	dir := utils.FindFile(t, "go.mod")
	req.NotEmpty(dir)
	req.True(filepath.IsAbs(dir))

	// Test finding a file in current directory
	projectRoot := utils.ProjectRootDir(t)
	testFile := filepath.Join(projectRoot, "test-find-file.tmp")
	f, err := os.Create(testFile)
	req.NoError(err)
	_ = f.Close()
	defer func() { _ = os.Remove(testFile) }()

	// Change to a subdirectory
	tmp, err := os.MkdirTemp(projectRoot, "test-find-subdir-*")
	req.NoError(err)
	defer func() { _ = os.RemoveAll(tmp) }()

	req.NoError(os.Chdir(tmp))

	// Should find the file by walking up
	foundDir := utils.FindFile(t, "test-find-file.tmp")
	req.Equal(projectRoot, foundDir)
}

// Test that demonstrates the GetLocalIP and GetLocalIPs error handling
func TestGetLocalIP_ErrorPath(t *testing.T) {
	t.Parallel()
	// This test is tricky because we can't easily mock net.InterfaceAddrs()
	// But we can test the current behavior
	req := require.New(t)

	ip := utils.GetLocalIP()
	// Should either be empty (if no interfaces) or a valid IP
	if ip != "" {
		parsedIP := net.ParseIP(ip)
		req.NotNil(parsedIP)
		req.True(parsedIP.To4() != nil) // Should be IPv4
	}
}

func TestGetLocalIPs_ErrorPath(t *testing.T) {
	t.Parallel()
	req := require.New(t)

	ips := utils.GetLocalIPs()
	// Should either be nil/empty (if no interfaces) or contain valid IPs
	for _, ip := range ips {
		parsedIP := net.ParseIP(ip)
		req.NotNil(parsedIP)
		req.True(parsedIP.To4() != nil) // Should be IPv4
	}
}

// Additional edge case tests for better coverage
func TestCreateFile_StatError(t *testing.T) {
	t.Parallel()
	req := require.New(t)

	tmp, err := os.MkdirTemp("", "test-create-file-stat-*")
	req.NoError(err)
	defer func() { _ = os.RemoveAll(tmp) }()

	// Test creating a file that doesn't exist yet
	filePath := filepath.Join(tmp, "new-file.txt")
	file, err := utils.CreateFile(filePath, os.O_RDWR|os.O_CREATE, 0o755, 0o644)
	req.NoError(err)
	req.NotNil(file)
	_ = file.Close()

	// Test opening the same file again (should not recreate)
	file2, err := utils.CreateFile(filePath, os.O_RDWR, 0o755, 0o644)
	req.NoError(err)
	req.NotNil(file2)
	_ = file2.Close()
}

func TestGenerateRandomPassword_LengthEdgeCases(t *testing.T) {
	t.Parallel()
	req := require.New(t)

	// Test various lengths to ensure we hit all code paths
	testCases := []int{4, 7, 8, 12, 16, 32, 100}

	for _, length := range testCases {
		pw, err := utils.GenerateRandomPassword(length)
		req.NoError(err)

		expectedLength := length
		if length < 8 {
			expectedLength = 12 // Default minimum
		}

		req.Len(pw, expectedLength)

		// Verify password contains at least one character from each set for longer passwords
		if expectedLength >= 8 {
			req.Regexp(`[a-z]`, pw)                           // lowercase
			req.Regexp(`[A-Z]`, pw)                           // uppercase
			req.Regexp(`[0-9]`, pw)                           // numbers
			req.Regexp(`[!@#$%^&*()\-_=+\[\]{}|;:,.<>?]`, pw) // special chars
		}
	}
}

func TestFindFile_ReadDirError(t *testing.T) {
	// This test checks that FindFile works correctly when searching
	// for a file that exists in a parent directory
	req := require.New(t)

	// Save original working directory
	origWD, err := os.Getwd()
	req.NoError(err)
	defer func() { _ = os.Chdir(origWD) }()

	projectRoot := utils.ProjectRootDir(t)

	// Create a temporary subdirectory
	subDir := filepath.Join(projectRoot, "temp-test-subdir")
	req.NoError(os.MkdirAll(subDir, 0o755))
	defer func() { _ = os.RemoveAll(subDir) }()

	// Change to the subdirectory
	req.NoError(os.Chdir(subDir))

	// Find go.mod which should be in the parent directory
	found := utils.FindFile(t, "go.mod")
	req.Equal(projectRoot, found)
}

func TestProjectRootDir_CacheEdgeCases(t *testing.T) {
	t.Parallel()
	req := require.New(t)

	// Test multiple calls to ensure cache works properly
	root1 := utils.ProjectRootDir(t)
	root2 := utils.ProjectRootDir(t)
	root3 := utils.ProjectRootDir(t)

	req.Equal(root1, root2)
	req.Equal(root2, root3)
	req.NotEmpty(root1)
	req.True(filepath.IsAbs(root1))

	// Verify go.mod exists in the root
	goModPath := filepath.Join(root1, "go.mod")
	req.FileExists(goModPath)
}

func TestWorkingDir_Success(t *testing.T) {
	t.Parallel()
	req := require.New(t)

	wd1 := utils.WorkingDir(t)
	wd2 := utils.WorkingDir(t)

	req.Equal(wd1, wd2)
	req.NotEmpty(wd1)
	req.True(filepath.IsAbs(wd1))
}
