package docker

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func writeContext(t *testing.T, dir, name, host, current string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "config.json"),
		[]byte(`{"currentContext":"`+current+`"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256([]byte(name))
	metaDir := filepath.Join(dir, "contexts", "meta", hex.EncodeToString(sum[:]))
	if err := os.MkdirAll(metaDir, 0o755); err != nil {
		t.Fatal(err)
	}
	meta := `{"Name":"` + name + `","Endpoints":{"docker":{"Host":"` + host + `"}}}`
	if err := os.WriteFile(filepath.Join(metaDir, "meta.json"), []byte(meta), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestContextHost(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DOCKER_CONFIG", dir)
	t.Setenv("DOCKER_CONTEXT", "")
	writeContext(t, dir, "remote", "tcp://10.0.0.5:2375", "remote")

	if got := ContextHost(); got != "tcp://10.0.0.5:2375" {
		t.Fatalf("active context should resolve its host, got %q", got)
	}

	writeContext(t, dir, "remote", "tcp://10.0.0.5:2375", "default")
	if got := ContextHost(); got != "" {
		t.Fatalf("default context should resolve to empty (use FromEnv), got %q", got)
	}

	t.Setenv("DOCKER_CONTEXT", "remote")
	if got := ContextHost(); got != "tcp://10.0.0.5:2375" {
		t.Fatalf("DOCKER_CONTEXT should override currentContext, got %q", got)
	}
}
