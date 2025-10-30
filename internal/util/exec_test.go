package util

import (
	"strings"
	"testing"
	"time"
)

func TestExecuteCommand_Success(t *testing.T) {
	result := ExecuteCommand(ExecConfig{
		Command: "echo",
		Args:    []string{"hello world"},
	})

	if result.Error != nil {
		t.Fatalf("expected no error, got %v", result.Error)
	}

	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}

	stdout := strings.TrimSpace(result.Stdout)
	if stdout != "hello world" {
		t.Errorf("expected stdout 'hello world', got '%s'", stdout)
	}
}

func TestExecuteCommand_Failure(t *testing.T) {
	result := ExecuteCommand(ExecConfig{
		Command: "sh",
		Args:    []string{"-c", "exit 42"},
	})

	if result.Error == nil {
		t.Fatal("expected error, got nil")
	}

	if result.ExitCode != 42 {
		t.Errorf("expected exit code 42, got %d", result.ExitCode)
	}
}

func TestExecuteCommand_NonExistent(t *testing.T) {
	result := ExecuteCommand(ExecConfig{
		Command: "this-command-does-not-exist-12345",
		Args:    []string{},
	})

	if result.Error == nil {
		t.Fatal("expected error for non-existent command, got nil")
	}

	if result.ExitCode == 0 {
		t.Error("expected non-zero exit code for non-existent command")
	}
}

func TestExecuteCommand_Timeout(t *testing.T) {
	result := ExecuteCommand(ExecConfig{
		Command: "sleep",
		Args:    []string{"10"},
		Timeout: 100 * time.Millisecond,
	})

	if result.Error == nil {
		t.Fatal("expected timeout error, got nil")
	}

	// Should have a context deadline exceeded or signal killed error
	errStr := result.Error.Error()
	if !strings.Contains(errStr, "deadline exceeded") && !strings.Contains(errStr, "killed") {
		t.Errorf("expected timeout/killed error, got: %v", result.Error)
	}
}

func TestExecuteCommand_DefaultTimeout(t *testing.T) {
	// Quick command should complete with default timeout
	result := ExecuteCommand(ExecConfig{
		Command: "echo",
		Args:    []string{"test"},
		// Timeout not specified, should use default
	})

	if result.Error != nil {
		t.Fatalf("expected no error with default timeout, got %v", result.Error)
	}

	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}
}

func TestExecuteCommand_StdoutStderr(t *testing.T) {
	// Command that outputs to both stdout and stderr
	result := ExecuteCommand(ExecConfig{
		Command: "sh",
		Args:    []string{"-c", "echo 'output'; echo 'error' >&2; exit 1"},
	})

	if result.Error == nil {
		t.Fatal("expected error, got nil")
	}

	if result.ExitCode != 1 {
		t.Errorf("expected exit code 1, got %d", result.ExitCode)
	}

	// Note: stdout might be empty if command exits with error before flush
	// but stderr should contain our error message
	if !strings.Contains(result.Stderr, "error") {
		t.Errorf("expected stderr to contain 'error', got '%s'", result.Stderr)
	}
}

func TestExecuteShellCommand_Success(t *testing.T) {
	result := ExecuteShellCommand("echo 'hello' && echo 'world'", DefaultTimeout)

	if result.Error != nil {
		t.Fatalf("expected no error, got %v", result.Error)
	}

	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}

	stdout := strings.TrimSpace(result.Stdout)
	if !strings.Contains(stdout, "hello") || !strings.Contains(stdout, "world") {
		t.Errorf("expected stdout to contain 'hello' and 'world', got '%s'", stdout)
	}
}

func TestExecuteShellCommand_Pipe(t *testing.T) {
	result := ExecuteShellCommand("echo 'test' | wc -l", DefaultTimeout)

	if result.Error != nil {
		t.Fatalf("expected no error, got %v", result.Error)
	}

	stdout := strings.TrimSpace(result.Stdout)
	if !strings.Contains(stdout, "1") {
		t.Errorf("expected stdout to contain '1', got '%s'", stdout)
	}
}

func TestExecuteShellCommand_Failure(t *testing.T) {
	result := ExecuteShellCommand("exit 99", DefaultTimeout)

	if result.Error == nil {
		t.Fatal("expected error, got nil")
	}

	if result.ExitCode != 99 {
		t.Errorf("expected exit code 99, got %d", result.ExitCode)
	}
}
