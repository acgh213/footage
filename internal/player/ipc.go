package player

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"sync/atomic"
)


// ipc is a minimal synchronous JSON-IPC client for mpv's named-pipe protocol.
// It is safe for concurrent calls — each call acquires a lock and reads
// responses until it finds one matching its request_id (skipping event lines).
type ipc struct {
	conn   net.Conn
	reader *bufio.Reader
	nextID atomic.Int64
}

func newIPC(conn net.Conn) *ipc {
	return &ipc{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}
}

// send sends a command to mpv and returns the response data.
// The call is synchronous: it reads lines from mpv until the matching
// request_id arrives, discarding any event lines that arrive first.
func (c *ipc) send(command ...interface{}) (json.RawMessage, error) {
	id := c.nextID.Add(1)
	msg := map[string]interface{}{
		"command":    command,
		"request_id": id,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	data = append(data, '\n')

	if _, err := c.conn.Write(data); err != nil {
		return nil, fmt.Errorf("write to mpv: %w", err)
	}

	// Read lines until we see our response. mpv may send event lines first.
	for {
		line, err := c.reader.ReadBytes('\n')
		if err != nil {
			return nil, fmt.Errorf("read from mpv: %w", err)
		}
		var resp struct {
			RequestID int64           `json:"request_id"`
			Data      json.RawMessage `json:"data"`
			Error     string          `json:"error"`
		}
		if err := json.Unmarshal(line, &resp); err != nil {
			continue // skip malformed lines
		}
		if resp.RequestID != id {
			continue // event or response for another request
		}
		if resp.Error != "" && resp.Error != "success" {
			return nil, fmt.Errorf("mpv error: %s", resp.Error)
		}
		return resp.Data, nil
	}
}

func (c *ipc) close() {
	_ = c.conn.Close()
}
