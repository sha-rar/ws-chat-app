package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// LoadDotEnv reads settings from a .env file into the environment, without
// overriding variables that are already set. It returns the file it loaded,
// or "" if there was none.
//
// If ENV_FILE is set, that file is loaded and must exist. Otherwise it looks in
// the working directory and then each parent up to the repository root (the
// nearest directory containing .git), so `go run` works from backend/ with the
// .env in the project root. Outside a repository, only the working directory
// is checked.
func LoadDotEnv() (string, error) {
	path := os.Getenv("ENV_FILE")
	if path == "" {
		dir, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("find .env: %w", err)
		}
		if path = findDotEnv(dir); path == "" {
			return "", nil
		}
	}

	if err := godotenv.Load(path); err != nil {
		return "", fmt.Errorf("load %s: %w", path, err)
	}
	return path, nil
}

func findDotEnv(dir string) string {
	root := repoRoot(dir)
	for {
		if candidate := filepath.Join(dir, ".env"); isFile(candidate) {
			return candidate
		}
		if root == "" || dir == root {
			return ""
		}
		dir = filepath.Dir(dir)
	}
}

// repoRoot returns the nearest directory at or above dir containing .git, or ""
func repoRoot(dir string) string {
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}
