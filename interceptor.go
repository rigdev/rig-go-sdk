package rig

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/golang-jwt/jwt"
	"github.com/rigdev/rig-go-api/api/v1/authentication"
	"github.com/rigdev/rig-go-api/api/v1/authentication/authenticationconnect"
)

var _omitAuth = map[string]struct{}{
	"/api.v1.authentication.Service/Login":             {},
	"/api.v1.authentication.Service/Register":          {},
	"/api.v1.authentication.Service/VerifyEmail":       {},
	"/api.v1.authentication.Service/RefreshToken":      {},
	"/api.v1.authentication.Service/OauthCallback":     {},
	"/api.v1.authentication.Service/SendPasswordReset": {},
	"/api.v1.authentication.Service/ResetPassword":     {},
	"/api.v1.authentication.Service/GetAuthConfig":     {},
}

type authInterceptor struct {
	cfg       *config
	projectId string
}

func (i *authInterceptor) handleAuth(ctx context.Context, h http.Header, method string) {
	if _, ok := _omitAuth[method]; !ok {
		i.setAuthorization(ctx, h)
	}
}

func (i *authInterceptor) setAuthorization(ctx context.Context, h http.Header) {
	at := i.cfg.sm.GetAccessToken()
	if at == "" && i.cfg.login != nil {
		res, err := authenticationconnect.NewServiceClient(i.cfg.hc, i.cfg.host).Login(ctx, &connect.Request[authentication.LoginRequest]{
			Msg: i.cfg.login,
		})
		if err != nil {
			return
		}

		i.cfg.sm.SetAccessToken(res.Msg.GetToken().GetAccessToken(), res.Msg.GetToken().GetRefreshToken())
	}

	c := jwt.StandardClaims{}
	p := jwt.Parser{
		SkipClaimsValidation: true,
	}

	_, _, err := p.ParseUnverified(i.cfg.sm.GetAccessToken(), &c)
	if err != nil {
		return
	}

	if !c.VerifyExpiresAt(time.Now().Add(30*time.Second).Unix(), true) {
		res, err := authenticationconnect.NewServiceClient(i.cfg.hc, i.cfg.host).RefreshToken(ctx, &connect.Request[authentication.RefreshTokenRequest]{
			Msg: &authentication.RefreshTokenRequest{
				RefreshToken: i.cfg.sm.GetRefreshToken(),
			},
		})
		if err != nil {
			return
		}
		i.cfg.sm.SetAccessToken(res.Msg.GetToken().GetAccessToken(), res.Msg.GetToken().GetRefreshToken())
	}

	h.Set("Authorization", fmt.Sprint("Bearer ", i.cfg.sm.GetAccessToken()))
}

func (i *authInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, ar connect.AnyRequest) (connect.AnyResponse, error) {
		i.handleAuth(ctx, ar.Header(), ar.Spec().Procedure)
		return next(ctx, ar)
	}
}

func (i *authInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, s connect.Spec) connect.StreamingClientConn {
		conn := next(ctx, s)
		i.handleAuth(ctx, conn.RequestHeader(), s.Procedure)
		return conn
	}
}

func (i *authInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, shc connect.StreamingHandlerConn) error {
		i.handleAuth(ctx, shc.RequestHeader(), shc.Spec().Procedure)
		return next(ctx, shc)
	}
}
