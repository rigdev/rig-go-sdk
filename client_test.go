package rig

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	ctx := context.Background()

	client := NewClient()

	t.Log("successfully created client")

	// test user client connection
	createUserResp, err := client.User().Create(ctx, connect.NewRequest(&user.CreateRequest{Initializers: []*user.Update{
		{Field: &user.Update_Email{Email: "johndoe@acme.com"}},
		{Field: &user.Update_Password{Password: "TeamRig22!"}},
	}}))
	require.NoError(t, err)
	t.Logf("successfully created user with ID: %v", createUserResp.Msg.GetUser().GetUserId())

	_, err = client.User().Delete(ctx, connect.NewRequest(&user.DeleteRequest{
		UserId: createUserResp.Msg.GetUser().GetUserId(),
	}))
	require.NoError(t, err)
	t.Log("successfully deleted user")

	// test group client connection
	createGroupResp, err := client.Group().Create(ctx, connect.NewRequest(&group.CreateRequest{Initializers: []*group.Update{
		{Field: &group.Update_GroupId{GroupId: "my-group"}},
	}}))
	require.NoError(t, err)
	t.Log("successfully created group")

	_, err = client.Group().Delete(ctx, connect.NewRequest(&group.DeleteRequest{
		GroupId: createGroupResp.Msg.GetGroup().GetGroupId(),
	}))
	require.NoError(t, err)
	t.Log("successfully deleted group")
}
