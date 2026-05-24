package region

import "github.com/oklog/ulid/v2"

func newID(prefix string) string {
	return prefix + "_" + ulid.Make().String()
}

func NewRegionID() string   { return newID("region") }
func NewBookmarkID() string { return newID("bookmark") }
