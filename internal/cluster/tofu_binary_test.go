package cluster

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveTofuBinary(t *testing.T) {
	t.Run("configured path wins without PATH lookup", func(t *testing.T) {
		t.Setenv("PATH", "")

		got, err := resolveTofuBinary("  /opt/tools/custom-tofu  ")
		if err != nil {
			t.Fatalf("resolveTofuBinary() error = %v", err)
		}
		if got != "/opt/tools/custom-tofu" {
			t.Fatalf("resolveTofuBinary() = %q, want configured path", got)
		}
	})

	t.Run("discovers tofu on isolated PATH", func(t *testing.T) {
		binDir := t.TempDir()
		writeExecutable(t, filepath.Join(binDir, "tofu"))
		t.Setenv("PATH", binDir)

		got, err := resolveTofuBinary("")
		if err != nil {
			t.Fatalf("resolveTofuBinary() error = %v", err)
		}
		if got != "tofu" {
			t.Fatalf("resolveTofuBinary() = %q, want tofu", got)
		}
	})

	t.Run("falls back to terraform when tofu is absent", func(t *testing.T) {
		t.Setenv("PATH", t.TempDir())

		got, err := resolveTofuBinary("")
		if err != nil {
			t.Fatalf("resolveTofuBinary() error = %v", err)
		}
		if got != "terraform" {
			t.Fatalf("resolveTofuBinary() = %q, want terraform", got)
		}
	})
}

func writeExecutable(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write executable fixture: %v", err)
	}
}
