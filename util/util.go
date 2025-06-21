package util

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindDir finds the first directory at or above the current directory that
// contains a particular subdirectory. It returns the full path to the directory
// and the relative from the parent of the directory to the current directory.
func FindDir(targetDir string) (dirPath string, relPath string, err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", err
	}
	originalCwd := cwd
	for {
		candidate := filepath.Join(cwd, targetDir)
		st, err := os.Stat(candidate)
		if err == nil && st.IsDir() {
			relPath, err := filepath.Rel(cwd, originalCwd)
			if err != nil {
				return "", "", err
			}
			return candidate, relPath, nil
		}
		if cwd == "/" || err != nil && !os.IsNotExist(err) {
			break
		}
		cwd = filepath.Dir(cwd)
	}

	return "", "", fmt.Errorf("%s not found as directory at or above current directory", targetDir)
}

func CountTrue(args ...bool) int {
	count := 0
	for _, arg := range args {
		if arg {
			count += 1
		}
	}
	return count
}
