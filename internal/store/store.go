package store

import "time"

type UserData struct {
	Timestamps []time.Time
	TotalCount int
}

type Store interface {
	// AddRequest records a new request timestamp for the given user.
	AddRequest(userID string, timestamp time.Time)

	// GetTimestamps returns all request timestamps for a user
	GetTimestamps(userID string, windowStart time.Time) []time.Time

	// GetAllUsers returns data for all users
	GetAllUsers() map[string]*UserData

	// CleanExpired removes timestamps older than windowStart for a user.
	CleanExpired(userID string, windowStart time.Time)
}
