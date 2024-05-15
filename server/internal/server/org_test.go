package server

import (
	"context"
	"fmt"
	"testing"

	v1 "github.com/llm-operator/user-manager/api/v1"
	"github.com/llm-operator/user-manager/server/internal/store"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestOrganization(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("Authorization", "dummy"))

	for i := 0; i < 2; i++ {
		title := fmt.Sprintf("test %d", i)
		cresp, err := srv.CreateOrganization(ctx, &v1.CreateOrganizationRequest{
			Title: title,
		})
		assert.NoError(t, err)
		assert.Equal(t, title, cresp.Title)

		_, err = srv.AddUserToOrganization(ctx, &v1.AddUserToOrganizationRequest{
			OrganizationId: cresp.Id,
			UserId:         "user 1",
		})
		assert.NoError(t, err)
	}

	lresp, err := srv.ListOrganizations(ctx, &v1.ListOrganizationsRequest{})
	assert.NoError(t, err)
	assert.Len(t, lresp.Organizations, 2)

	_, err = srv.DeleteOrganization(ctx, &v1.DeleteOrganizationRequest{
		Id: lresp.Organizations[0].Id,
	})
	assert.NoError(t, err)

	lresp2, err := srv.ListOrganizations(ctx, &v1.ListOrganizationsRequest{})
	assert.NoError(t, err)
	assert.Len(t, lresp2.Organizations, 1)
	assert.Equal(t, lresp2.Organizations[0].Id, lresp.Organizations[1].Id)
}
