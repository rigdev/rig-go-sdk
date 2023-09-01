package main

import (
	"context"
	"fmt"
	"log"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/authentication"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig-go-sdk"
)

func main() {
	LoginExample()
}

func LoginExample() {
	userEmail := "foobar@example.com"
	userPassword := "mypassword"

	apiKey := "31aef3ea-affa-47b9-94cb-8c7d552c055b"

	client := rig.NewClient()

	ctx := context.Background()

	loginRes, err := client.Authentication().Login(ctx, &connect.Request[authentication.LoginRequest]{
		Msg: &authentication.LoginRequest{
			Method: &authentication.LoginRequest_UserPassword{
				UserPassword: &authentication.UserPassword{
					Identifier: &model.UserIdentifier{
						Identifier: &model.UserIdentifier_Email{
							Email: userEmail,
						},
					},
					Password: userPassword,
					ApiKey:   apiKey,
				},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Set the access and refresh token on the client.
	client.SetAccessToken(loginRes.Msg.GetToken().GetAccessToken(), loginRes.Msg.GetToken().GetRefreshToken())

	getRes, err := client.Authentication().Get(ctx, &connect.Request[authentication.GetRequest]{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("email: ", getRes.Msg.GetUserInfo().GetEmail())
}
