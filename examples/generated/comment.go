package stringsvc

import "time"

// Example structure
type Comment struct {
	Text     string
	Relates  *Comment
	PostedAt time.Time
}
