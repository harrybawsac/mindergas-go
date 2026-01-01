package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMain_CSVDryRun(t *testing.T) {
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

	if os.Getenv("TEST_MAIN") == "1" {
		os.Args = []string{"cmd", "--config", configPath, "--dry-run"}
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_CSVDryRun", "--config", configPath, "--dry-run")
	cmd.Env = append(os.Environ(), "TEST_MAIN=1")
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
	if os.Getenv("TEST_MAIN") == "1" {
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_MissingConfig", "--config", "/nonexistent/path.json")
	cmd.Env = append(os.Environ(), "TEST_MAIN=1")
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected error for missing config, got nil")
	}
}
