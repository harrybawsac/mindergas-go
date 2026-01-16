package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMain_CSVDryRun(t *testing.T) {
	// Build the binary first
	binPath := filepath.Join(t.TempDir(), "mindergas-test")
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("failed to build binary: %v", err)
	}

	tmpDir := t.TempDir()

	csvPath := filepath.Join(tmpDir, "test.csv")
	csvContent := `timestamp,value
2025-12-31T00:00:00,12345.678
`
	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		t.Fatalf("failed to write CSV: %v", err)
	}

	configPath := filepath.Join(tmpDir, "config.json")
	config := `{"csv_path": "` + csvPath + `", "token": "test-token"}`
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cmd := exec.Command(binPath, "--config", configPath, "--dry-run")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "12345.678") {
		t.Errorf("output should contain reading value, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "2025-12-31T00:00:00") {
		t.Errorf("output should contain date, got: %s", outputStr)
	}
}

func TestMain_MissingConfig(t *testing.T) {
	// Build the binary first
	binPath := filepath.Join(t.TempDir(), "mindergas-test")
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("failed to build binary: %v", err)
	}

	cmd := exec.Command(binPath, "--config", "/nonexistent/path.json")
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected error for missing config, got nil")
	}
}
