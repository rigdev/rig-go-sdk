package rig

// SessionManager is used by the Client to help maintain the access and refresh tokens.
// By default, an in-memory version will be used. A custom implementation can be provided
// using `WithSessionManager`, if the tokens should be stored e.g. in a config file.
type SessionManager interface {
	GetAccessToken() string
	GetRefreshToken() string

	SetAccessToken(accessToken, refreshToken string)
}

type simpleSessionManager struct {
	accessToken  string
	refreshToken string
}

func (s *simpleSessionManager) GetAccessToken() string {
	return s.accessToken
}

func (s *simpleSessionManager) GetRefreshToken() string {
	return s.refreshToken
}

func (s *simpleSessionManager) SetAccessToken(accessToken, refreshToken string) {
	s.accessToken = accessToken
	s.refreshToken = refreshToken
}
