package auth

import (
    "time"
)

type SessionStore interface {
    InsertSession(userID int64) (string, time.Time, error)
    GetUserIDByToken(token string) (int64, bool)
    DeleteSession(token string)
}
