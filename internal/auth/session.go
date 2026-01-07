package auth

type SessionStore interface {
	CreateSession(userID int64) (string, error)
	GetUserIDByToken(token string) (int64, bool)
	DeleteSessionByToken(token string)
}
