package rig

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"

	"connectrpc.com/connect"
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

func WithClient(client *http.Client) Option {
	return &withClientOption{client: client}
}

type withClientOption struct {
	client *http.Client
}

func (o *withClientOption) apply(c *config) {
	c.hc = o.client
}

type WithBasicAuth struct {
	cc ClientCredential
}

func WithBasicAuthOption(cc ClientCredential) Option {
	return &WithBasicAuth{cc: cc}
}

func (o *WithBasicAuth) apply(c *config) {
	if o.cc == (ClientCredential{}) {
		o.cc = ClientCredential{
			ClientID:     os.Getenv("RIG_CLIENT_ID"),
			ClientSecret: os.Getenv("RIG_CLIENT_SECRET"),
		}
	}

	base64ClientCredentials := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", o.cc.ClientID, o.cc.ClientSecret)))
	c.basicAuth = fmt.Sprintf("Basic %s", base64ClientCredentials)
}
