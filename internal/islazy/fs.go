package islazy

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// CreateDir creates a directory if it does not exist, returning the final
// normalized directory as a result.
func CreateDir(dir string) (string, error) {
	var err error

	if strings.HasPrefix(dir, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(homeDir, dir[1:])
	}

	dir, err = filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	return dir, nil
}

// CreateFileWithDir creates a file, relative to a directory, returning the
// final normalized path as a result.
func CreateFileWithDir(destination string) (string, error) {
	dir := filepath.Dir(destination)
	file := filepath.Base(destination)

	if file == "." || file == "/" {
		return "", fmt.Errorf("destination does not appear to be a valid file path: %s", destination)
	}

	absDir, err := CreateDir(dir)
	if err != nil {
		return "", err
	}

	absPath := filepath.Join(absDir, file)
	fileHandle, err := os.Create(absPath)
	if err != nil {
		return "", err
	}
	defer fileHandle.Close()

	return absPath, nil
}

// SafeFileName takes a string and returns a string safe to use as
// a file name.
func SafeFileName(s string) string {
	var builder strings.Builder

	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' {
			builder.WriteRune(r)
		} else {
			builder.WriteRune('-')
		}
	}

	return builder.String()
}

// FileExists returns true if a path exists
func FileExists(path string) bool {
	_, err := os.Stat(path)

	return !os.IsNotExist(err)
}

// MoveFile moves a file from a to b
func MoveFile(sourcePath, destPath string) error {
	if err := os.Rename(sourcePath, destPath); err == nil {
		return nil
	}

	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	err = os.Remove(sourcePath)
	if err != nil {
		return err
	}

	return nil
}
