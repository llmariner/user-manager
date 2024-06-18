package store

import (
	"testing"

	gerrors "github.com/llm-operator/common/pkg/gormlib/errors"
	"github.com/stretchr/testify/assert"
)

func TestProject(t *testing.T) {
	s, tearDown := NewTest(t)
	defer tearDown()

	p := &Project{
		TenantID:            "t1",
		ProjectID:           "p1",
		OrganizationID:      "o1",
		Title:               "Test Project",
		KubernetesNamespace: "test-namespace",
	}

	gotPrj, err := s.CreateProject(CreateProjectParams{
		TenantID:            p.TenantID,
		ProjectID:           p.ProjectID,
		OrganizationID:      p.OrganizationID,
		Title:               p.Title,
		KubernetesNamespace: p.KubernetesNamespace,
	})
	assert.NoError(t, err)
	assert.NotNil(t, gotPrj)
	assert.Equal(t, p.TenantID, gotPrj.TenantID)
	assert.Equal(t, p.ProjectID, gotPrj.ProjectID)
	assert.Equal(t, p.Title, gotPrj.Title)

	gotPrj, err = s.GetProject(GetProjectParams{
		TenantID:       p.TenantID,
		OrganizationID: p.OrganizationID,
		ProjectID:      p.ProjectID,
	})
	assert.NoError(t, err)
	assert.NotNil(t, gotPrj)
	assert.Equal(t, p.TenantID, gotPrj.TenantID)
	assert.Equal(t, p.ProjectID, gotPrj.ProjectID)
	assert.Equal(t, p.Title, gotPrj.Title)

	_, err = s.CreateProject(CreateProjectParams{
		TenantID:            p.TenantID,
		ProjectID:           "p2",
		OrganizationID:      p.OrganizationID,
		Title:               "Test Project 2",
		KubernetesNamespace: "ns",
	})
	assert.NoError(t, err)
	gotPrjs, err := s.ListProjectsByTenantIDAndOrganizationID(p.TenantID, p.OrganizationID)
	assert.NoError(t, err)
	assert.Len(t, gotPrjs, 2)

	err = s.DeleteProject(p.ProjectID)
	assert.NoError(t, err)
	gotPrj, err = s.GetProject(GetProjectParams{
		TenantID:       p.TenantID,
		OrganizationID: p.OrganizationID,
		ProjectID:      p.ProjectID,
	})
	assert.Error(t, err)
	assert.Nil(t, gotPrj)
}

func TestGetDefaultProject(t *testing.T) {
	s, tearDown := NewTest(t)
	defer tearDown()

	params := []CreateProjectParams{
		{
			TenantID:       "t0",
			ProjectID:      "p0",
			OrganizationID: "o0",
			Title:          "Test Organization 0",
			IsDefault:      true,
		},
		{
			TenantID:       "t0",
			ProjectID:      "p1",
			OrganizationID: "o1",
			Title:          "Test Organization 1",
			IsDefault:      false,
		},
		{
			TenantID:       "t2",
			ProjectID:      "p2",
			OrganizationID: "o2",
			Title:          "Test Organization 2",
			IsDefault:      true,
		},
	}
	for _, p := range params {
		_, err := s.CreateProject(p)
		assert.NoError(t, err)
	}

	tcs := []struct {
		tenantID      string
		wantProjectID string
	}{
		{
			tenantID:      "t0",
			wantProjectID: "p0",
		},
		{
			tenantID:      "t2",
			wantProjectID: "p2",
		},
	}

	for _, tc := range tcs {
		got, err := s.GetDefaultProject(tc.tenantID)
		assert.NoError(t, err)
		assert.Equal(t, tc.wantProjectID, got.ProjectID)
	}
}

func TestListAllProjects(t *testing.T) {
	s, tearDown := NewTest(t)
	defer tearDown()

	_, err := s.CreateProject(CreateProjectParams{
		TenantID:            "t0",
		ProjectID:           "p0",
		OrganizationID:      "o0",
		Title:               "Test Project 0",
		KubernetesNamespace: "ns",
	})
	assert.NoError(t, err)

	_, err = s.CreateProject(CreateProjectParams{
		TenantID:            "t1",
		ProjectID:           "p1",
		OrganizationID:      "o1",
		Title:               "Test Project 1",
		KubernetesNamespace: "ns",
	})
	assert.NoError(t, err)

	gotPrjs, err := s.ListAllProjects()
	assert.NoError(t, err)
	assert.Len(t, gotPrjs, 2)
}

func TestProject_UniqueConstraint(t *testing.T) {
	s, tearDown := NewTest(t)
	defer tearDown()

	_, err := s.CreateProject(CreateProjectParams{
		TenantID:            "t0",
		ProjectID:           "p0",
		OrganizationID:      "o0",
		Title:               "Test Project",
		KubernetesNamespace: "ns0",
	})
	assert.NoError(t, err)

	// Same title.
	_, err = s.CreateProject(CreateProjectParams{
		TenantID:            "t0",
		ProjectID:           "p1",
		OrganizationID:      "o1",
		Title:               "Test Project",
		KubernetesNamespace: "ns1",
	})
	assert.Error(t, err)
	assert.True(t, gerrors.IsUniqueConstraintViolation(err))

	// Same title, but different tenant.
	_, err = s.CreateProject(CreateProjectParams{
		TenantID:            "t1",
		ProjectID:           "p2",
		OrganizationID:      "o2",
		Title:               "Test Project",
		KubernetesNamespace: "ns2",
	})
	assert.NoError(t, err)
}
