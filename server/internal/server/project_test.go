package server

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-logr/logr/testr"
	"github.com/llmariner/rbac-manager/pkg/auth"
	v1 "github.com/llmariner/user-manager/api/v1"
	"github.com/llmariner/user-manager/server/internal/config"
	"github.com/llmariner/user-manager/server/internal/store"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestProject(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st, testr.New(t))
	isrv := NewInternal(st, testr.New(t))
	ctx := context.Background()

	var orgs []*v1.Organization
	var projs []*v1.Project
	for i := 0; i < 2; i++ {
		title := fmt.Sprintf("Test organization %d", i)
		org, err := srv.CreateOrganization(ctx, &v1.CreateOrganizationRequest{
			Title: title,
		})
		assert.NoError(t, err)
		orgs = append(orgs, org)

		title = fmt.Sprintf("Test project %d", i)
		proj, err := srv.CreateProject(ctx, &v1.CreateProjectRequest{
			Title:               title,
			OrganizationId:      org.Id,
			KubernetesNamespace: "test",
		})
		assert.NoError(t, err)
		assert.Equal(t, title, proj.Title)
		assert.Equal(t, org.Id, proj.OrganizationId)
		assert.Equal(t, "test", proj.KubernetesNamespace)

		projs = append(projs, proj)
	}

	resp, err := srv.ListProjects(ctx, &v1.ListProjectsRequest{
		OrganizationId: orgs[0].Id,
	})
	assert.NoError(t, err)
	assert.Len(t, resp.Projects, 1)
	assert.Equal(t, projs[0].Id, resp.Projects[0].Id)

	resp, err = srv.ListProjects(ctx, &v1.ListProjectsRequest{
		OrganizationId: orgs[1].Id,
	})
	assert.NoError(t, err)
	assert.Len(t, resp.Projects, 1)
	assert.Equal(t, projs[1].Id, resp.Projects[0].Id)

	resp, err = isrv.ListProjects(ctx, &v1.ListProjectsRequest{})
	assert.NoError(t, err)
	assert.Len(t, resp.Projects, 2)

	_, err = srv.DeleteProject(ctx, &v1.DeleteProjectRequest{
		OrganizationId: orgs[0].Id,
		Id:             projs[0].Id,
	})
	assert.NoError(t, err)

	resp, err = srv.ListProjects(ctx, &v1.ListProjectsRequest{
		OrganizationId: orgs[0].Id,
	})
	assert.NoError(t, err)
	assert.Len(t, resp.Projects, 0)
}

func TestProjectUser(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st, testr.New(t))
	isrv := NewInternal(st, testr.New(t))

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("Authorization", "dummy"))
	org, err := srv.CreateOrganization(ctx, &v1.CreateOrganizationRequest{
		Title: "Test organization",
	})
	assert.NoError(t, err)

	// Delete the default user to make the rest of the test simple.
	_, err = srv.DeleteOrganizationUser(ctx, &v1.DeleteOrganizationUserRequest{
		OrganizationId: org.Id,
		UserId:         defaultUserID,
	})
	assert.NoError(t, err)

	proj, err := srv.CreateProject(ctx, &v1.CreateProjectRequest{
		Title:               "Test project",
		OrganizationId:      org.Id,
		KubernetesNamespace: "test",
	})
	assert.NoError(t, err)

	pu0, err := srv.CreateProjectUser(ctx, &v1.CreateProjectUserRequest{
		ProjectId:      proj.Id,
		OrganizationId: org.Id,
		UserId:         "u0",
		Role:           v1.ProjectRole_PROJECT_ROLE_OWNER,
	})
	assert.NoError(t, err)
	assert.Equal(t, "u0", pu0.UserId)

	pu1, err := srv.CreateProjectUser(ctx, &v1.CreateProjectUserRequest{
		ProjectId:      proj.Id,
		OrganizationId: org.Id,
		UserId:         "u1",
		Role:           v1.ProjectRole_PROJECT_ROLE_MEMBER,
	})
	assert.NoError(t, err)
	assert.Equal(t, "u1", pu1.UserId)

	resp, err := srv.ListProjectUsers(ctx, &v1.ListProjectUsersRequest{
		ProjectId:      proj.Id,
		OrganizationId: org.Id,
	})
	assert.NoError(t, err)
	assert.Len(t, resp.Users, 2)

	resp, err = isrv.ListProjectUsers(ctx, &v1.ListProjectUsersRequest{})
	assert.NoError(t, err)
	assert.Len(t, resp.Users, 2)

	_, err = srv.DeleteProjectUser(ctx, &v1.DeleteProjectUserRequest{
		ProjectId:      proj.Id,
		OrganizationId: org.Id,
		UserId:         pu0.UserId,
	})
	assert.NoError(t, err)

	resp, err = srv.ListProjectUsers(ctx, &v1.ListProjectUsersRequest{
		ProjectId:      proj.Id,
		OrganizationId: org.Id,
	})
	assert.NoError(t, err)
	assert.Len(t, resp.Users, 1)
	assert.Equal(t, pu1.UserId, resp.Users[0].UserId)
}

func TestCreateProject_UniqueConstraintViolation(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st, testr.New(t))
	ctx := context.Background()

	o, err := srv.CreateOrganization(ctx, &v1.CreateOrganizationRequest{
		Title: "title",
	})
	assert.NoError(t, err)

	_, err = srv.CreateProject(ctx, &v1.CreateProjectRequest{
		Title:               "title",
		OrganizationId:      o.Id,
		KubernetesNamespace: "ns",
	})
	assert.NoError(t, err)

	_, err = srv.CreateProject(ctx, &v1.CreateProjectRequest{
		Title:               "title",
		OrganizationId:      o.Id,
		KubernetesNamespace: "ns",
	})
	assert.Error(t, err)
	assert.Equal(t, codes.AlreadyExists, status.Code(err))
}

func TestCreateProjectUser_UniqueConstraintViolation(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st, testr.New(t))
	ctx := context.Background()

	o, err := srv.CreateOrganization(ctx, &v1.CreateOrganizationRequest{
		Title: "title",
	})
	assert.NoError(t, err)

	p, err := srv.CreateProject(ctx, &v1.CreateProjectRequest{
		Title:               "title",
		OrganizationId:      o.Id,
		KubernetesNamespace: "ns",
	})
	assert.NoError(t, err)

	_, err = srv.CreateProjectUser(ctx, &v1.CreateProjectUserRequest{
		ProjectId:      p.Id,
		OrganizationId: o.Id,
		UserId:         "u0",
		Role:           v1.ProjectRole_PROJECT_ROLE_OWNER,
	})
	assert.NoError(t, err)

	_, err = srv.CreateProjectUser(ctx, &v1.CreateProjectUserRequest{
		ProjectId:      p.Id,
		OrganizationId: o.Id,
		UserId:         "u0",
		Role:           v1.ProjectRole_PROJECT_ROLE_OWNER,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.AlreadyExists, status.Code(err))
}

func TestCreateDefaultProject(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st, testr.New(t))
	ctx := context.Background()

	org, err := srv.CreateOrganization(ctx, &v1.CreateOrganizationRequest{
		Title: "Test organization",
	})
	assert.NoError(t, err)

	// Delete the default user to make the rest of the test simple.
	_, err = srv.DeleteOrganizationUser(ctx, &v1.DeleteOrganizationUserRequest{
		OrganizationId: org.Id,
		UserId:         defaultUserID,
	})
	assert.NoError(t, err)

	_, err = srv.CreateOrganizationUser(ctx, &v1.CreateOrganizationUserRequest{
		OrganizationId: org.Id,
		UserId:         "u0",
		Role:           v1.OrganizationRole_ORGANIZATION_ROLE_OWNER,
	})
	assert.NoError(t, err)
	_, err = srv.CreateOrganizationUser(ctx, &v1.CreateOrganizationUserRequest{
		OrganizationId: org.Id,
		UserId:         "u1",
		Role:           v1.OrganizationRole_ORGANIZATION_ROLE_READER,
	})
	assert.NoError(t, err)

	c := &config.DefaultProjectConfig{
		Title:               "Default project",
		KubernetesNamespace: "ns",
	}
	err = srv.CreateDefaultProject(ctx, c, org.Id, defaultTenantID)
	assert.NoError(t, err)

	p, err := st.GetDefaultProject(defaultTenantID)
	assert.NoError(t, err)
	assert.Equal(t, c.KubernetesNamespace, p.KubernetesNamespace)
	assert.True(t, p.IsDefault)

	pus, err := st.ListProjectUsersByProjectID(p.ProjectID)
	assert.NoError(t, err)
	assert.Len(t, pus, 1)
	assert.Equal(t, "u0", pus[0].UserID)

	// Default project cannot be deleted.
	_, err = srv.DeleteProject(context.Background(), &v1.DeleteProjectRequest{
		Id:             p.ProjectID,
		OrganizationId: org.Id,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestCreateProject_EnableAuth(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st, testr.New(t))
	srv.enableAuth = true

	o := createDefaultOrg(t, srv, "u0")

	u0Ctx := auth.AppendUserInfoToContext(context.Background(), auth.UserInfo{
		UserID: "u0",
	})

	u1Ctx := auth.AppendUserInfoToContext(context.Background(), auth.UserInfo{
		UserID: "u1",
	})

	req := &v1.CreateProjectRequest{
		Title:               "title",
		OrganizationId:      o.OrganizationID,
		KubernetesNamespace: "n0",
	}

	_, err := srv.CreateProject(u1Ctx, req)
	assert.Error(t, err)
	assert.Equal(t, codes.FailedPrecondition, status.Code(err))

	_, err = srv.CreateProject(u0Ctx, req)
	assert.NoError(t, err)
}

func TestListProjects_EnableAuth(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st, testr.New(t))
	srv.enableAuth = true

	o := createDefaultOrg(t, srv, "u0")

	u0Ctx := auth.AppendUserInfoToContext(context.Background(), auth.UserInfo{
		UserID: "u0",
	})
	_, err := srv.CreateProject(u0Ctx, &v1.CreateProjectRequest{
		Title:               "title",
		OrganizationId:      o.OrganizationID,
		KubernetesNamespace: "n0",
	})
	assert.NoError(t, err)

	req := &v1.ListProjectsRequest{
		OrganizationId: o.OrganizationID,
	}
	resp, err := srv.ListProjects(u0Ctx, req)
	assert.NoError(t, err)
	assert.Len(t, resp.Projects, 1)

	u1Ctx := auth.AppendUserInfoToContext(context.Background(), auth.UserInfo{
		UserID: "u1",
	})
	resp, err = srv.ListProjects(u1Ctx, req)
	assert.NoError(t, err)
	assert.Empty(t, 0, resp.Projects)
}

func TestDeleteProject_EnableAuth(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st, testr.New(t))
	srv.enableAuth = true

	o := createDefaultOrg(t, srv, "u0")

	u0Ctx := auth.AppendUserInfoToContext(context.Background(), auth.UserInfo{
		UserID: "u0",
	})
	p, err := srv.CreateProject(u0Ctx, &v1.CreateProjectRequest{
		Title:               "title",
		OrganizationId:      o.OrganizationID,
		KubernetesNamespace: "n0",
	})
	assert.NoError(t, err)

	u1Ctx := auth.AppendUserInfoToContext(context.Background(), auth.UserInfo{
		UserID: "u1",
	})

	req := &v1.DeleteProjectRequest{
		OrganizationId: o.OrganizationID,
		Id:             p.Id,
	}
	_, err = srv.DeleteProject(u1Ctx, req)
	assert.Error(t, err)
	assert.Equal(t, codes.FailedPrecondition, status.Code(err))

	_, err = srv.DeleteProject(u0Ctx, req)
	assert.NoError(t, err)
}

func TestProjectUser_EnableAuth(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st, testr.New(t))
	srv.enableAuth = true

	o := createDefaultOrg(t, srv, "user0")

	u0Ctx := auth.AppendUserInfoToContext(context.Background(), auth.UserInfo{
		UserID: "user0",
	})

	p, err := srv.CreateProject(u0Ctx, &v1.CreateProjectRequest{
		Title:               "title",
		OrganizationId:      o.OrganizationID,
		KubernetesNamespace: "n0",
	})
	assert.NoError(t, err)

	// Add "u2" to the org.
	_, err = srv.CreateOrganizationUser(u0Ctx, &v1.CreateOrganizationUserRequest{
		OrganizationId: o.OrganizationID,
		UserId:         "user2",
		Role:           v1.OrganizationRole_ORGANIZATION_ROLE_READER,
	})
	assert.NoError(t, err)

	creq := &v1.CreateProjectUserRequest{
		ProjectId:      p.Id,
		OrganizationId: o.OrganizationID,
		UserId:         "user2",
		Role:           v1.ProjectRole_PROJECT_ROLE_MEMBER,
	}
	u1Ctx := auth.AppendUserInfoToContext(context.Background(), auth.UserInfo{
		UserID: "user1",
	})
	_, err = srv.CreateProjectUser(u1Ctx, creq)
	assert.Error(t, err)
	assert.Equal(t, codes.FailedPrecondition, status.Code(err))

	_, err = srv.CreateProjectUser(u0Ctx, creq)
	assert.NoError(t, err)

	lreq := &v1.ListProjectUsersRequest{
		ProjectId:      p.Id,
		OrganizationId: o.OrganizationID,
	}
	_, err = srv.ListProjectUsers(u1Ctx, lreq)
	assert.Error(t, err)
	assert.Equal(t, codes.FailedPrecondition, status.Code(err))

	_, err = srv.ListProjectUsers(u0Ctx, lreq)
	assert.NoError(t, err)

	u2Ctx := auth.AppendUserInfoToContext(context.Background(), auth.UserInfo{
		UserID: "user2",
	})
	_, err = srv.ListProjectUsers(u2Ctx, lreq)
	assert.NoError(t, err)

	dreq := &v1.DeleteProjectUserRequest{
		ProjectId:      p.Id,
		OrganizationId: o.OrganizationID,
		UserId:         "user2",
	}
	_, err = srv.DeleteProjectUser(u1Ctx, dreq)
	assert.Error(t, err)
	assert.Equal(t, codes.FailedPrecondition, status.Code(err))

	_, err = srv.DeleteProjectUser(u2Ctx, dreq)
	assert.Error(t, err)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))

	_, err = srv.DeleteProjectUser(u0Ctx, dreq)
	assert.NoError(t, err)
}
