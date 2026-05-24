package player

import (
	"encoding/json"
	"fmt"
)

// getFloat is a helper on ipc that extracts a float64 from a get_property call.
func (c *ipc) getFloat(command ...interface{}) (float64, error) {
	data, err := c.send(command...)
	if err != nil {
		return 0, err
	}
	if data == nil {
		return 0, fmt.Errorf("nil response from mpv")
	}
	var v float64
	if err := json.Unmarshal(data, &v); err != nil {
		return 0, fmt.Errorf("parse mpv response: %w (raw: %s)", err, data)
	}
	return v, nil
}
