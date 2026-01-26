package utils

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const MovieBaseDir = "./movies"

func ValidateAndResolvePath(inputPath string) (string, error) {
	clean := filepath.Clean(inputPath)

	if filepath.IsAbs(clean) {
		return "", errors.New("absolute paths are not allowed")
	}

	fullPath := filepath.Join(MovieBaseDir, clean)

	resolved, err := filepath.EvalSymlinks(fullPath)
	if err != nil {
		return "", errors.New("file does not exist")
	}

	baseAbs, _ := filepath.Abs(MovieBaseDir)
	resolvedAbs, _ := filepath.Abs(resolved)

	if !strings.HasPrefix(resolvedAbs, baseAbs+string(os.PathSeparator)) {
		return "", errors.New("path outside movie directory")
	}

	info, err := os.Stat(resolvedAbs)
	if err != nil || info.IsDir() {
		return "", errors.New("invalid movie file")
	}

	return resolvedAbs, nil
}
