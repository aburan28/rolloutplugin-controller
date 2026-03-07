package pluginclient

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestVerifyPlugin_ValidChecksum(t *testing.T) {
	// Create a temp file with known content
	tmpDir := t.TempDir()
	pluginPath := filepath.Join(tmpDir, "testplugin")
	content := []byte("test plugin content")
	if err := os.WriteFile(pluginPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	// Compute the expected sha256
	h := sha256.New()
	h.Write(content)
	expectedHash := fmt.Sprintf("%x", h.Sum(nil))

	// Verify should succeed with correct checksum
	if err := VerifyPlugin(pluginPath, expectedHash); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestVerifyPlugin_InvalidChecksum(t *testing.T) {
	// Create a temp file with known content
	tmpDir := t.TempDir()
	pluginPath := filepath.Join(tmpDir, "testplugin")
	content := []byte("test plugin content")
	if err := os.WriteFile(pluginPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	// Verify should fail with wrong checksum
	if err := VerifyPlugin(pluginPath, "wrongchecksum"); err == nil {
		t.Fatal("expected error for wrong checksum, got nil")
	}
}

func TestVerifyPlugin_MissingFile(t *testing.T) {
	if err := VerifyPlugin("/nonexistent/path/plugin", "somechecksum"); err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
