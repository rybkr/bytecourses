package auth

import (
    "time"
)

// SessionStore owns the lifecycle of authentication sessions.
type SessionStore interface {
    // Create creates a new session for the given user ID.
    // Returns a session identifier and an absolute expiration time.
    // The store becomes the owner of the session.
    Create(userID int64) (string, time.Time, error)
    
    // GetUserIDByToken resolves a session token to a user ID.
    // Returns (userID, true) if the session exists and is valid, else (0, false).
    GetUserIDByToken(token string) (int64, bool)

    // Delete invalidates a session token.
    // Deleting a non-existent token is a no-op.
    Delete(token string)
}
