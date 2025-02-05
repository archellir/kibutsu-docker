package types

import "fmt"

// TerminalMessage represents a WebSocket message for terminal communication
type TerminalMessage struct {
	// Type of message: "input", "output", "resize", "error", or "exit"
	Type string `json:"type"`

	// Data contains the message payload (command input or command output)
	Data string `json:"data,omitempty"`

	// Cols and Rows are used for terminal resize events
	Cols uint `json:"cols,omitempty"`
	Rows uint `json:"rows,omitempty"`
}

// TerminalSize represents the dimensions of a terminal
type TerminalSize struct {
	// Cols represents the number of columns in the terminal
	Cols uint `json:"cols"`

	// Rows represents the number of rows in the terminal
	Rows uint `json:"rows"`
}

// ExecConfig represents the configuration for a container exec session
type ExecConfig struct {
	// Command to execute in the container
	Cmd []string `json:"cmd"`

	// WorkingDir specifies the working directory for the exec process
	WorkingDir string `json:"workingDir,omitempty"`

	// Environment variables for the exec process
	Env []string `json:"env,omitempty"`

	// User to run the exec process as
	User string `json:"user,omitempty"`

	// Privileged runs the exec process with extended privileges
	Privileged bool `json:"privileged,omitempty"`

	// Tty allocates a pseudo-TTY for the exec process
	Tty bool `json:"tty"`
}

// TerminalError represents an error that occurred during terminal operations
type TerminalError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Code    int    `json:"code,omitempty"`
}

func (e *TerminalError) Error() string {
	return fmt.Sprintf("terminal error: %s - %s", e.Type, e.Message)
}

// Common terminal errors
var (
	ErrTerminalClosed = &TerminalError{
		Type:    "terminal_closed",
		Message: "terminal connection closed",
	}
	ErrInvalidMessage = &TerminalError{
		Type:    "invalid_message",
		Message: "invalid terminal message format",
	}
	ErrResizeFailed = &TerminalError{
		Type:    "resize_failed",
		Message: "failed to resize terminal",
	}
	ErrExecFailed = &TerminalError{
		Type:    "exec_failed",
		Message: "failed to execute command in container",
	}
)
