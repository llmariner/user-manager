package store

import (
	"testing"
	"time"

	gerrors "github.com/llmariner/common/pkg/gormlib/errors"
	v1 "github.com/llmariner/user-manager/api/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

func TestProject(t *testing.T) {
	s, tearDown := NewTest(t)
	defer tearDown()

	p := &Project{
		TenantID:       "t1",
		ProjectID:      "p1",
		OrganizationID: "o1",
		Title:          "Test Project",
	}
	as := []*v1.ProjectAssignment{
		{
			Namespace: " test-namespace",
		},
	}

	gotPrj, err := s.CreateProject(CreateProjectParams{
		TenantID:       p.TenantID,
		ProjectID:      p.ProjectID,
		OrganizationID: p.OrganizationID,
		Title:          p.Title,
		Assignments:    as,
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
		TenantID:       p.TenantID,
		ProjectID:      "p2",
		OrganizationID: p.OrganizationID,
		Title:          "Test Project 2",
		Assignments: []*v1.ProjectAssignment{
			{
				Namespace: "ns2",
			},
		},
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

func TestToProto(t *testing.T) {
	now := time.Now()

	as := &v1.ProjectAssignments{
		Assignments: []*v1.ProjectAssignment{
			{
				Namespace: "ns0",
			},
			{
				Namespace: "ns1",
				ClusterId: "cluster1",
			},
		},
	}
	asb, err := proto.Marshal(as)
	assert.NoError(t, err)

	tcs := []struct {
		name string
		p    *Project
		want *v1.Project
	}{
		{
			name: "with kubernetes_namespace",
			p: &Project{
				Model: gorm.Model{
					CreatedAt: now,
				},
				TenantID:            "t1",
				ProjectID:           "p1",
				OrganizationID:      "o1",
				Title:               "Test Project",
				KubernetesNamespace: "ns0",
			},
			want: &v1.Project{
				Id:                  "p1",
				OrganizationId:      "o1",
				Title:               "Test Project",
				KubernetesNamespace: "ns0",
				Assignments: []*v1.ProjectAssignment{
					{
						Namespace: "ns0",
					},
				},
				CreatedAt: now.UTC().Unix(),
			},
		},
		{
			name: "without kubernetes_namespace",
			p: &Project{
				Model: gorm.Model{
					CreatedAt: now,
				},
				TenantID:       "t1",
				ProjectID:      "p1",
				OrganizationID: "o1",
				Title:          "Test Project",
				Assignments:    asb,
			},
			want: &v1.Project{
				Id:                  "p1",
				OrganizationId:      "o1",
				Title:               "Test Project",
				KubernetesNamespace: "ns0",
				Assignments: []*v1.ProjectAssignment{
					{
						Namespace: "ns0",
					},
					{
						Namespace: "ns1",
						ClusterId: "cluster1",
					},
				},
				CreatedAt: now.UTC().Unix(),
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.p.ToProto()
			assert.NoError(t, err)
			assert.Truef(t, proto.Equal(tc.want, got), "wanted %+v, but got %+v", tc.want, got)
		})
	}
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
		TenantID:       "t0",
		ProjectID:      "p0",
		OrganizationID: "o0",
		Title:          "Test Project 0",
		Assignments: []*v1.ProjectAssignment{
			{
				Namespace: "ns",
			},
		},
	})
	assert.NoError(t, err)

	_, err = s.CreateProject(CreateProjectParams{
		TenantID:       "t1",
		ProjectID:      "p1",
		OrganizationID: "o1",
		Title:          "Test Project 1",
		Assignments: []*v1.ProjectAssignment{
			{
				Namespace: "ns",
			},
		},
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
		TenantID:       "t0",
		ProjectID:      "p0",
		OrganizationID: "o0",
		Title:          "Test Project",
		Assignments: []*v1.ProjectAssignment{
			{
				Namespace: "ns0",
			},
		},
	})
	assert.NoError(t, err)

	// Same title.
	_, err = s.CreateProject(CreateProjectParams{
		TenantID:       "t0",
		ProjectID:      "p1",
		OrganizationID: "o1",
		Title:          "Test Project",
		Assignments: []*v1.ProjectAssignment{
			{
				Namespace: "ns1",
			},
		},
	})
	assert.Error(t, err)
	assert.True(t, gerrors.IsUniqueConstraintViolation(err))

	// Same title, but different tenant.
	_, err = s.CreateProject(CreateProjectParams{
		TenantID:       "t1",
		ProjectID:      "p2",
		OrganizationID: "o2",
		Title:          "Test Project",
		Assignments: []*v1.ProjectAssignment{
			{
				Namespace: "ns2",
			},
		},
	})
	assert.NoError(t, err)
}
