package rig

import (
	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/authentication"
)

type Option interface {
	apply(c *config)
}

func WithHost(host string) Option {
	return &withHostOption{host: host}
}

type withHostOption struct {
	host string
}

func (o *withHostOption) apply(c *config) {
	c.host = o.host
}

func WithClientCredentials(cc ClientCredential) Option {
	return &withClientCredentials{cc: cc}
}

type withClientCredentials struct {
	cc ClientCredential
}

func (o *withClientCredentials) apply(c *config) {
	c.login = &authentication.LoginRequest{
		Method: &authentication.LoginRequest_ClientCredentials{
			ClientCredentials: &authentication.ClientCredentials{
				ClientId:     o.cc.ClientID,
				ClientSecret: o.cc.ClientSecret,
			},
		},
	}
}

func WithSessionManager(sm SessionManager) Option {
	return &withSessionManager{sm: sm}
}

type withSessionManager struct {
	sm SessionManager
}

func (o *withSessionManager) apply(c *config) {
	c.sm = o.sm
}

type Interceptor = connect.Interceptor

func WithInterceptors(ics ...Interceptor) Option {
	return &withInterceptors{ics: ics}
}

type withInterceptors struct {
	ics []Interceptor
}

func (o *withInterceptors) apply(c *config) {
	c.ics = o.ics
}
