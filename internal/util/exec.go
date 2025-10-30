package util

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

// CommandResult holds the result of a command execution
type CommandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Error    error
}

// ExecConfig holds configuration for command execution
type ExecConfig struct {
	Command string
	Args    []string
	Timeout time.Duration
}

// DefaultTimeout is the default timeout for command execution (30 seconds)
const DefaultTimeout = 30 * time.Second

// ExecuteCommand executes a command with timeout and captures output
func ExecuteCommand(cfg ExecConfig) CommandResult {
	result := CommandResult{
		ExitCode: -1,
	}

	// Set default timeout if not specified
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create command
	cmd := exec.CommandContext(ctx, cfg.Command, cfg.Args...)

	// Capture stdout and stderr
	stdout, err := cmd.Output()
	if err != nil {
		result.Error = err
		// Try to get stderr from ExitError
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.Stderr = string(exitErr.Stderr)
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.Error = fmt.Errorf("failed to execute command: %w", err)
		}
		return result
	}

	result.Stdout = string(stdout)
	result.ExitCode = 0
	return result
}

// ExecuteShellCommand executes a shell command string (using sh -c).
//
// SECURITY WARNING: This function executes commands through a shell, which means
// the command string can contain shell metacharacters (pipes, redirects, etc.).
// The caller MUST ensure that the command string comes from a trusted source
// (e.g., configuration files with appropriate file permissions) and MUST validate
// any user-provided input that is substituted into the command.
//
// Example of safe usage:
//   - Command from trusted config file: ✓ Safe
//   - Command with validated date substitution: ✓ Safe (if date format is validated)
//   - Command with arbitrary user input: ✗ UNSAFE (command injection risk)
func ExecuteShellCommand(cmd string, timeout time.Duration) CommandResult {
	return ExecuteCommand(ExecConfig{
		Command: "sh",
		Args:    []string{"-c", cmd},
		Timeout: timeout,
	})
}
