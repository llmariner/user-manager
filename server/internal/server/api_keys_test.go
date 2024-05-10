package server

import (
	"context"
	"testing"

	v1 "github.com/llm-operator/user-manager/api/v1"
	"github.com/llm-operator/user-manager/server/internal/store"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestAPIKey(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st)
	isrv := NewInternal(st)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("Authorization", "dummy"))

	cresp, err := srv.CreateAPIKey(ctx, &v1.CreateAPIKeyRequest{
		Name: "dummy",
	})
	assert.NoError(t, err)
	assert.Equal(t, "dummy", cresp.Name)

	_, err = srv.CreateAPIKey(ctx, &v1.CreateAPIKeyRequest{
		Name: "dummy",
	})
	assert.Error(t, err)
	assert.Equal(t, codes.AlreadyExists, status.Code(err))

	lresp, err := srv.ListAPIKeys(ctx, &v1.ListAPIKeysRequest{})
	assert.NoError(t, err)
	assert.Len(t, lresp.Data, 1)

	lresp, err = isrv.ListAPIKeys(ctx, &v1.ListAPIKeysRequest{})
	assert.NoError(t, err)
	assert.Len(t, lresp.Data, 1)

	_, err = srv.DeleteAPIKey(ctx, &v1.DeleteAPIKeyRequest{
		Id: cresp.Id,
	})
	assert.NoError(t, err)

	lresp, err = srv.ListAPIKeys(ctx, &v1.ListAPIKeysRequest{})
	assert.NoError(t, err)
	assert.Empty(t, lresp.Data)
}
