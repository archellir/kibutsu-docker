package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"golang.org/x/net/websocket"
)

type TerminalHandler struct {
	client *client.Client
}

type TerminalMessage struct {
	Type    string `json:"type"`
	Data    string `json:"data"`
	Cols    uint   `json:"cols,omitempty"`
	Rows    uint   `json:"rows,omitempty"`
	Command string `json:"command,omitempty"`
}

func NewTerminalHandler(client *client.Client) *TerminalHandler {
	return &TerminalHandler{client: client}
}

func (h *TerminalHandler) HandleTerminal(w http.ResponseWriter, r *http.Request) {
	containerId := strings.TrimPrefix(r.URL.Path, "/api/containers/")
	containerId = strings.TrimSuffix(containerId, "/exec")

	// Verify container exists and is running
	ctx := r.Context()
	container, err := h.client.ContainerInspect(ctx, containerId)
	if err != nil {
		http.Error(w, "Container not found", http.StatusNotFound)
		return
	}

	if !container.State.Running {
		http.Error(w, "Container is not running", http.StatusBadRequest)
		return
	}

	// Upgrade connection to websocket
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		h.handleConnection(ctx, ws, containerId)
	}).ServeHTTP(w, r)
}

func (h *TerminalHandler) handleConnection(ctx context.Context, ws *websocket.Conn, containerId string) {
	// Create exec configuration
	execConfig := types.ExecConfig{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          []string{"/bin/sh"},
	}

	// Create exec instance
	exec, err := h.client.ContainerExecCreate(ctx, containerId, execConfig)
	if err != nil {
		log.Printf("Error creating exec: %v", err)
		return
	}

	// Attach to exec instance
	resp, err := h.client.ContainerExecAttach(ctx, exec.ID, types.ExecStartCheck{Tty: true})
	if err != nil {
		log.Printf("Error attaching to exec: %v", err)
		return
	}
	defer resp.Close()

	// Start copying data between websocket and container
	var wg sync.WaitGroup
	wg.Add(2)

	// Copy from websocket to container
	go func() {
		defer wg.Done()
		for {
			var msg TerminalMessage
			if err := websocket.JSON.Receive(ws, &msg); err != nil {
				if err != io.EOF {
					log.Printf("Error receiving websocket message: %v", err)
				}
				return
			}

			switch msg.Type {
			case "resize":
				// Handle terminal resize
				if err := h.client.ContainerExecResize(ctx, exec.ID, container.ResizeOptions{
					Height: msg.Rows,
					Width:  msg.Cols,
				}); err != nil {
					log.Printf("Error resizing terminal: %v", err)
				}
			case "input":
				// Send input to container
				if _, err := resp.Conn.Write([]byte(msg.Data)); err != nil {
					log.Printf("Error writing to container: %v", err)
					return
				}
			}
		}
	}()

	// Copy from container to websocket
	go func() {
		defer wg.Done()
		buffer := make([]byte, 4096)
		for {
			n, err := resp.Reader.Read(buffer)
			if err != nil {
				if err != io.EOF {
					log.Printf("Error reading from container: %v", err)
				}
				return
			}

			msg := TerminalMessage{
				Type: "output",
				Data: string(buffer[:n]),
			}

			if err := websocket.JSON.Send(ws, msg); err != nil {
				log.Printf("Error sending websocket message: %v", err)
				return
			}
		}
	}()

	// Wait for both goroutines to complete
	wg.Wait()

	// Get exec instance info to check exit code
	inspect, err := h.client.ContainerExecInspect(ctx, exec.ID)
	if err != nil {
		log.Printf("Error inspecting exec instance: %v", err)
		return
	}

	// Send exit message
	exitMsg := TerminalMessage{
		Type: "exit",
		Data: fmt.Sprintf("Process exited with code %d", inspect.ExitCode),
	}
	websocket.JSON.Send(ws, exitMsg)
}
