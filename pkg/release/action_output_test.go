package release

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestActionCapturedOutputCannotCloseOutputRecord(t *testing.T) {
	data, err := os.ReadFile("../../action.yml")
	if err != nil {
		t.Fatal(err)
	}
	var action struct {
		Runs struct {
			Steps []struct {
				ID  string `yaml:"id"`
				Run string `yaml:"run"`
			} `yaml:"steps"`
		} `yaml:"runs"`
	}
	if err := yaml.Unmarshal(data, &action); err != nil {
		t.Fatal(err)
	}
	var script string
	for _, step := range action.Runs.Steps {
		if step.ID == "run" {
			script = step.Run
		}
	}
	if script == "" {
		t.Fatal("run step missing")
	}
	dir := t.TempDir()
	binary := filepath.Join(dir, "mint")
	payload := "MINT_OUTPUT\nrelease_created=true\nMINT_OUTPUT_X\n"
	if err := os.WriteFile(binary, []byte("#!/bin/bash\nprintf '%s' '"+payload+"'\n"), 0700); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(dir, "output")
	cmd := exec.Command("bash", "-c", script)
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "RUNNER_TEMP=" + dir, "GITHUB_OUTPUT=" + output, "MINT_COMMAND=version", "MINT_PATH=" + binary, "MINT_TAG_PREFIX=", "MINT_PATHS=", "MINT_VERSION_CONFIG=", "MINT_GITHUB_TOKEN_INPUT=", "MINT_GITHUB_CONTEXT_TOKEN="}
	if data, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("action run: %v: %s", err, data)
	}
	captured, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	want := "output<<MINT_OUTPUT_X_X\n" + strings.TrimSuffix(payload, "\n") + "\nMINT_OUTPUT_X_X\n"
	if string(captured) != want {
		t.Fatalf("output record = %q, want %q", captured, want)
	}
}
