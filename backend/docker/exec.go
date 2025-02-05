package docker

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// ExecManager handles container exec operations
type ExecManager struct {
	client *client.Client
	execs  sync.Map // Maps exec IDs to active exec instances
}

// ExecConfig represents the configuration for an exec instance
type ExecConfig struct {
	Cmd          []string
	Tty          bool
	AttachStdin  bool
	AttachStdout bool
	AttachStderr bool
	WorkingDir   string
	Env          []string
	User         string
	Privileged   bool
}

// ExecInstance represents an active exec instance
type ExecInstance struct {
	ID     string
	Config ExecConfig
	conn   types.HijackedResponse
	mu     sync.Mutex
}

// NewExecManager creates a new exec manager
func NewExecManager(client *client.Client) *ExecManager {
	return &ExecManager{
		client: client,
	}
}

// Create creates a new exec instance in a container
func (m *ExecManager) Create(ctx context.Context, containerID string, config ExecConfig) (*ExecInstance, error) {
	execConfig := types.ExecConfig{
		Cmd:          config.Cmd,
		Tty:          config.Tty,
		AttachStdin:  config.AttachStdin,
		AttachStdout: config.AttachStdout,
		AttachStderr: config.AttachStderr,
		WorkingDir:   config.WorkingDir,
		Env:          config.Env,
		User:         config.User,
		Privileged:   config.Privileged,
	}

	resp, err := m.client.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create exec: %w", err)
	}

	execAttach, err := m.client.ContainerExecAttach(ctx, resp.ID, types.ExecStartCheck{
		Tty: config.Tty,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to attach to exec: %w", err)
	}

	instance := &ExecInstance{
		ID:     resp.ID,
		Config: config,
		conn:   execAttach,
	}

	m.execs.Store(resp.ID, instance)
	return instance, nil
}

// Resize changes the size of the TTY
func (m *ExecManager) Resize(ctx context.Context, execID string, height, width uint) error {
	return m.client.ContainerExecResize(ctx, execID, container.ResizeOptions{
		Height: height,
		Width:  width,
	})
}

// Start starts the exec instance and handles I/O
func (i *ExecInstance) Start(stdin io.Reader, stdout, stderr io.Writer) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	// Handle TTY mode
	if i.Config.Tty {
		go func() {
			io.Copy(stdout, i.conn.Reader)
		}()
	} else {
		// Use Docker's multiplexed I/O for non-TTY mode
		go func() {
			_, err := stdCopy(stdout, stderr, i.conn.Reader)
			if err != nil && err != io.EOF {
				// Log error but don't fail the exec
				fmt.Printf("Error copying exec output: %v\n", err)
			}
		}()
	}

	// Copy stdin if attached
	if i.Config.AttachStdin && stdin != nil {
		go func() {
			io.Copy(i.conn.Conn, stdin)
		}()
	}

	return nil
}

// Close closes the exec instance and cleans up resources
func (i *ExecInstance) Close() error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.conn.Conn != nil {
		i.conn.Close()
		return nil
	}
	return nil
}

// GetExitCode gets the exit code of the exec instance
func (m *ExecManager) GetExitCode(ctx context.Context, execID string) (int, error) {
	inspect, err := m.client.ContainerExecInspect(ctx, execID)
	if err != nil {
		return -1, fmt.Errorf("failed to inspect exec: %w", err)
	}
	return inspect.ExitCode, nil
}

// Remove removes an exec instance from the manager
func (m *ExecManager) Remove(execID string) {
	if instance, ok := m.execs.Load(execID); ok {
		if exec, ok := instance.(*ExecInstance); ok {
			exec.Close()
		}
		m.execs.Delete(execID)
	}
}

// stdCopy is a helper function to handle multiplexed streams
func stdCopy(stdout, stderr io.Writer, src io.Reader) (written int64, err error) {
	buf := make([]byte, 32*1024)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			// Determine if this is stdout or stderr
			var dst io.Writer
			if buf[0] == 1 {
				dst = stdout
			} else if buf[0] == 2 {
				dst = stderr
			}

			if dst != nil {
				nw, ew := dst.Write(buf[8:nr])
				if nw > 0 {
					written += int64(nw)
				}
				if ew != nil {
					err = ew
					break
				}
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return written, err
}
