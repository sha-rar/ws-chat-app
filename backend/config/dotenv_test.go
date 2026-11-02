package config

import (
	"os"
	"path/filepath"
	"testing"
)

// newRepo creates a fake repository root (a directory containing .git) with a
// nested backend/ directory, and returns both paths
func newRepo(t *testing.T) (root, backend string) {
	t.Helper()
	root = t.TempDir()
	backend = filepath.Join(root, "backend")
	for _, dir := range []string{filepath.Join(root, ".git"), backend} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return root, backend
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

// unsetAfter makes sure variables a test loads don't leak into other tests
func unsetAfter(t *testing.T, keys ...string) {
	t.Helper()
	for _, key := range keys {
		os.Unsetenv(key)
		t.Cleanup(func() { os.Unsetenv(key) })
	}
}

func TestLoadDotEnvFindsRepoRootFileFromSubdirectory(t *testing.T) {
	root, backend := newRepo(t)
	writeFile(t, filepath.Join(root, ".env"), "DOTENV_TEST_PORT=9123\n")
	unsetAfter(t, "DOTENV_TEST_PORT", "ENV_FILE")
	t.Chdir(backend)

	path, err := LoadDotEnv()
	if err != nil {
		t.Fatalf("LoadDotEnv: %v", err)
	}
	if path != filepath.Join(root, ".env") {
		t.Errorf("loaded %q, want the repo root .env", path)
	}
	if got := os.Getenv("DOTENV_TEST_PORT"); got != "9123" {
		t.Errorf("DOTENV_TEST_PORT = %q, want 9123", got)
	}
}

func TestLoadDotEnvPrefersNearestFile(t *testing.T) {
	root, backend := newRepo(t)
	writeFile(t, filepath.Join(root, ".env"), "DOTENV_TEST_WHICH=root\n")
	writeFile(t, filepath.Join(backend, ".env"), "DOTENV_TEST_WHICH=backend\n")
	unsetAfter(t, "DOTENV_TEST_WHICH", "ENV_FILE")
	t.Chdir(backend)

	if _, err := LoadDotEnv(); err != nil {
		t.Fatalf("LoadDotEnv: %v", err)
	}
	if got := os.Getenv("DOTENV_TEST_WHICH"); got != "backend" {
		t.Errorf("DOTENV_TEST_WHICH = %q, want backend", got)
	}
}

func TestLoadDotEnvDoesNotLookAboveRepoRoot(t *testing.T) {
	outer := t.TempDir()
	writeFile(t, filepath.Join(outer, ".env"), "DOTENV_TEST_OUTSIDE=leaked\n")
	root := filepath.Join(outer, "repo")
	if err := os.MkdirAll(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	unsetAfter(t, "DOTENV_TEST_OUTSIDE", "ENV_FILE")
	t.Chdir(root)

	path, err := LoadDotEnv()
	if err != nil {
		t.Fatalf("LoadDotEnv: %v", err)
	}
	if path != "" || os.Getenv("DOTENV_TEST_OUTSIDE") != "" {
		t.Errorf("loaded %q from outside the repository", path)
	}
}

func TestLoadDotEnvOutsideRepoOnlyChecksWorkingDirectory(t *testing.T) {
	parent := t.TempDir()
	if repoRoot(parent) != "" {
		t.Skip("temp directory is inside a git repository")
	}
	writeFile(t, filepath.Join(parent, ".env"), "DOTENV_TEST_PARENT=leaked\n")
	dir := filepath.Join(parent, "app")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	unsetAfter(t, "DOTENV_TEST_PARENT", "ENV_FILE")
	t.Chdir(dir)

	if path, err := LoadDotEnv(); err != nil || path != "" {
		t.Errorf("LoadDotEnv = %q, %v; want nothing loaded", path, err)
	}

	writeFile(t, filepath.Join(dir, ".env"), "DOTENV_TEST_PARENT=here\n")
	if path, err := LoadDotEnv(); err != nil || path != filepath.Join(dir, ".env") {
		t.Errorf("LoadDotEnv = %q, %v; want the working directory's .env", path, err)
	}
}

func TestLoadDotEnvDoesNotOverrideEnvironment(t *testing.T) {
	root, _ := newRepo(t)
	writeFile(t, filepath.Join(root, ".env"), "DOTENV_TEST_KEEP=from-file\n")
	unsetAfter(t, "ENV_FILE")
	t.Setenv("DOTENV_TEST_KEEP", "from-env")
	t.Chdir(root)

	if _, err := LoadDotEnv(); err != nil {
		t.Fatalf("LoadDotEnv: %v", err)
	}
	if got := os.Getenv("DOTENV_TEST_KEEP"); got != "from-env" {
		t.Errorf("DOTENV_TEST_KEEP = %q, want the existing value from-env", got)
	}
}

func TestLoadDotEnvWithoutFile(t *testing.T) {
	_, backend := newRepo(t)
	unsetAfter(t, "ENV_FILE")
	t.Chdir(backend)

	if path, err := LoadDotEnv(); err != nil || path != "" {
		t.Errorf("LoadDotEnv = %q, %v; want nothing loaded and no error", path, err)
	}
}

func TestLoadDotEnvIgnoresDirectoryNamedDotEnv(t *testing.T) {
	root, _ := newRepo(t)
	if err := os.Mkdir(filepath.Join(root, ".env"), 0o755); err != nil {
		t.Fatal(err)
	}
	unsetAfter(t, "ENV_FILE")
	t.Chdir(root)

	if path, err := LoadDotEnv(); err != nil || path != "" {
		t.Errorf("LoadDotEnv = %q, %v; want the directory ignored", path, err)
	}
}

func TestLoadDotEnvUsesEnvFile(t *testing.T) {
	root, backend := newRepo(t)
	writeFile(t, filepath.Join(root, ".env"), "DOTENV_TEST_SOURCE=default\n")
	custom := filepath.Join(root, "staging.env")
	writeFile(t, custom, "DOTENV_TEST_SOURCE=custom\n")
	unsetAfter(t, "DOTENV_TEST_SOURCE")
	t.Setenv("ENV_FILE", custom)
	t.Chdir(backend)

	path, err := LoadDotEnv()
	if err != nil {
		t.Fatalf("LoadDotEnv: %v", err)
	}
	if path != custom || os.Getenv("DOTENV_TEST_SOURCE") != "custom" {
		t.Errorf("loaded %q (DOTENV_TEST_SOURCE=%q), want ENV_FILE's values", path, os.Getenv("DOTENV_TEST_SOURCE"))
	}
}

func TestLoadDotEnvFailsWhenEnvFileIsMissing(t *testing.T) {
	t.Setenv("ENV_FILE", filepath.Join(t.TempDir(), "missing.env"))

	if _, err := LoadDotEnv(); err == nil {
		t.Error("LoadDotEnv succeeded with a missing ENV_FILE")
	}
}
