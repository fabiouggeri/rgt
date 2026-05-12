package auth

import (
	"rgt-server/option"
	"strings"
)

type UserAuthenticator interface {
	Authenticate(username, password string) bool
}

type AuthenticatorFactory interface {
	Create(conf map[string]option.Option) UserAuthenticator
}

type PassthroughAuthenticator struct{}

var authenticatorsFactories map[string]AuthenticatorFactory = make(map[string]AuthenticatorFactory, 0)

func AddAuthenticatorFactory(id string, factory AuthenticatorFactory) {
	authenticatorsFactories[strings.ToLower(id)] = factory
}

func RemoveAuthenticatorFactory(id string) {
	delete(authenticatorsFactories, strings.ToLower(id))
}

func NewAuthenticator(conf map[string]option.Option) UserAuthenticator {
	if conf == nil {
		return NewPassthroughAuthenticator()
	}
	mode := conf["mode"]
	if mode == nil {
		return NewPassthroughAuthenticator()
	}
	auth := authenticatorsFactories[strings.ToLower(mode.GetString())]
	if auth == nil {
		return NewPassthroughAuthenticator()
	}
	return auth.Create(conf)
}

func NewPassthroughAuthenticator() UserAuthenticator {
	return &PassthroughAuthenticator{}
}

func (p *PassthroughAuthenticator) Authenticate(username, password string) bool {
	return true
}
