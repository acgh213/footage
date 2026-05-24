package session

import "time"

// Session is a named focus set of video files with associated metadata.
type Session struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Files     []string  `json:"files"`
	ActiveIdx int       `json:"active_idx"`
	PresetName string   `json:"preset_name,omitempty"`
}

// ActiveFile returns the currently active video file path, or "" if none.
func (s *Session) ActiveFile() string {
	if len(s.Files) == 0 || s.ActiveIdx < 0 || s.ActiveIdx >= len(s.Files) {
		return ""
	}
	return s.Files[s.ActiveIdx]
}

// AddFile appends a file to the session if not already present.
func (s *Session) AddFile(path string) {
	for _, f := range s.Files {
		if f == path {
			return
		}
	}
	s.Files = append(s.Files, path)
	s.UpdatedAt = time.Now()
}

// RemoveFile removes a file by path and adjusts ActiveIdx.
func (s *Session) RemoveFile(path string) {
	for i, f := range s.Files {
		if f == path {
			s.Files = append(s.Files[:i], s.Files[i+1:]...)
			if s.ActiveIdx >= len(s.Files) && s.ActiveIdx > 0 {
				s.ActiveIdx = len(s.Files) - 1
			}
			s.UpdatedAt = time.Now()
			return
		}
	}
}

// SetActive sets the active file by index. Clamps to valid range.
func (s *Session) SetActive(idx int) {
	if idx < 0 {
		idx = 0
	}
	if idx >= len(s.Files) {
		idx = len(s.Files) - 1
	}
	s.ActiveIdx = idx
	s.UpdatedAt = time.Now()
}
