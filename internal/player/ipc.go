package player

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
)

// ipc is a JSON-IPC client for mpv's named-pipe protocol.
// A background goroutine continuously reads all lines from mpv and routes
// each response to the waiting caller by request_id. Event lines (no
// request_id) are discarded — we don't subscribe to any in phase 0.
type ipc struct {
	conn   net.Conn
	nextID atomic.Int64

	mu      sync.Mutex
	pending map[int64]chan response
	closed  bool
}

type response struct {
	Data  json.RawMessage
	Error string
}

func newIPC(conn net.Conn) *ipc {
	c := &ipc{
		conn:    conn,
		pending: make(map[int64]chan response),
	}
	go c.readLoop()
	return c
}

// readLoop runs in a goroutine and dispatches every line mpv sends.
func (c *ipc) readLoop() {
	scanner := bufio.NewScanner(c.conn)
	for scanner.Scan() {
		line := scanner.Bytes()
		var msg struct {
			RequestID int64           `json:"request_id"`
			Data      json.RawMessage `json:"data"`
			Error     string          `json:"error"`
			Event     string          `json:"event"`
		}
		if err := json.Unmarshal(line, &msg); err != nil {
			continue
		}
		if msg.Event != "" || msg.RequestID == 0 {
			continue // event line, not a command response
		}
		c.mu.Lock()
		ch, ok := c.pending[msg.RequestID]
		if ok {
			delete(c.pending, msg.RequestID)
		}
		c.mu.Unlock()
		if ok {
			ch <- response{Data: msg.Data, Error: msg.Error}
		}
	}
	// Connection closed — wake up all pending callers with an error.
	c.mu.Lock()
	c.closed = true
	for id, ch := range c.pending {
		ch <- response{Error: "ipc connection closed"}
		delete(c.pending, id)
	}
	c.mu.Unlock()
}

// send sends a command to mpv and waits for the matching response.
func (c *ipc) send(command ...interface{}) (json.RawMessage, error) {
	id := c.nextID.Add(1)

	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil, fmt.Errorf("mpv IPC connection is closed")
	}
	ch := make(chan response, 1)
	c.pending[id] = ch
	c.mu.Unlock()

	msg := map[string]interface{}{
		"command":    command,
		"request_id": id,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, err
	}
	data = append(data, '\n')

	if _, err := c.conn.Write(data); err != nil {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, fmt.Errorf("write to mpv: %w", err)
	}

	resp := <-ch
	if resp.Error != "" && resp.Error != "success" {
		return nil, fmt.Errorf("mpv: %s", resp.Error)
	}
	return resp.Data, nil
}

func (c *ipc) close() {
	_ = c.conn.Close()
}
