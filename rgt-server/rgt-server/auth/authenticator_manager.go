package auth

type AuthenticatorManager struct {
	authenticators map[string]UserAuthenticator
}

func NewAuthenticatorManager() *AuthenticatorManager {
	return &AuthenticatorManager{
		authenticators: make(map[string]UserAuthenticator),
	}
}

func (m *AuthenticatorManager) AddAuthenticator(authId string, authenticator UserAuthenticator) {
	m.authenticators[authId] = authenticator
}

func (m *AuthenticatorManager) GetAuthenticator(authId string) UserAuthenticator {
	return m.authenticators[authId]
}

func (m *AuthenticatorManager) AuthenticateUser(authId, username, password string) bool {
	authenticator, found := m.authenticators[authId]
	if found {
		return authenticator.Authenticate(username, password)
	}
	return true
}
