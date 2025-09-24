package store

import (
	v1 "github.com/llmariner/user-manager/api/v1"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

// Project is a model for project.
type Project struct {
	gorm.Model

	ProjectID      string `gorm:"uniqueIndex"`
	OrganizationID string

	TenantID string `gorm:"uniqueIndex:idx_projects_tenant_id_title"`
	Title    string `gorm:"uniqueIndex:idx_projects_tenant_id_title"`

	IsDefault bool

	// KubernetesNamespace is the namespace where the fine-tuning jobs for the organization run.
	// TODO(kenji): Currently we don't set the unique constraint so that multiple orgs can use the same namespace,
	// but revisit the design.
	KubernetesNamespace string

	// Assignments is the assignments of the project. It is a marshaled v1.ProjectAssignments.
	Assignments []byte
}

// ToProto converts the organization to proto.
func (p *Project) ToProto() (*v1.Project, error) {
	// Populate the KubernetesNamespace and Assignments from each other
	// for backward compatibility .
	// TODO(kenji): Remove this once all clients are updated.
	var (
		kn = p.KubernetesNamespace
		as []*v1.ProjectAssignment
	)

	if kn != "" {
		// Populate the Assignments from KubernetesNamespace.
		as = []*v1.ProjectAssignment{
			{
				Namespace: kn,
			},
		}
	} else {
		// Populate the KubernetesNamespace from Assignments.
		var asproto v1.ProjectAssignments
		if err := proto.Unmarshal(p.Assignments, &asproto); err != nil {
			return nil, err
		}
		as = asproto.Assignments

		for _, a := range asproto.Assignments {
			if a.ClusterId == "" {
				kn = a.Namespace
				break
			}
		}
	}

	return &v1.Project{
		Id:                  p.ProjectID,
		OrganizationId:      p.OrganizationID,
		Title:               p.Title,
		KubernetesNamespace: kn,
		Assignments:         as,
		CreatedAt:           p.CreatedAt.UTC().Unix(),
		IsDefault:           p.IsDefault,
	}, nil
}

// CreateProjectParams is the parameters for CreateProject.xo
type CreateProjectParams struct {
	ProjectID      string
	OrganizationID string
	TenantID       string
	Title          string
	Assignments    []*v1.ProjectAssignment
	IsDefault      bool
}

// CreateProject creates a new project.
func (s *S) CreateProject(p CreateProjectParams) (*Project, error) {
	return CreateProjectInTransaction(s.db, p)
}

// CreateProjectInTransaction creates a new project in a transaction.
func CreateProjectInTransaction(tx *gorm.DB, p CreateProjectParams) (*Project, error) {
	asb, err := proto.Marshal(&v1.ProjectAssignments{Assignments: p.Assignments})
	if err != nil {
		return nil, err
	}

	project := &Project{
		ProjectID:      p.ProjectID,
		OrganizationID: p.OrganizationID,
		TenantID:       p.TenantID,
		Title:          p.Title,
		Assignments:    asb,
		IsDefault:      p.IsDefault,
	}
	if err := tx.Create(project).Error; err != nil {
		return nil, err
	}
	return project, nil
}

// GetProjectParams is the parameters for GetProject.
type GetProjectParams struct {
	TenantID       string
	OrganizationID string
	ProjectID      string
}

// GetProject gets an project.
func (s *S) GetProject(p GetProjectParams) (*Project, error) {
	var project Project
	if err := s.db.Where("tenant_id = ? AND organization_id = ? AND project_id = ?", p.TenantID, p.OrganizationID, p.ProjectID).First(&project).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

// GetDefaultProject gets a default project.
func (s *S) GetDefaultProject(tenantID string) (*Project, error) {
	var prj Project
	if err := s.db.Where("tenant_id = ? AND is_default = ?", tenantID, true).First(&prj).Error; err != nil {
		return nil, err
	}
	return &prj, nil
}

// ListProjectsByTenantIDAndOrganizationID lists projects by tenant ID and organization ID.
func (s *S) ListProjectsByTenantIDAndOrganizationID(tenantID, orgID string) ([]*Project, error) {
	var projects []*Project
	if err := s.db.Where("tenant_id = ? AND organization_id = ?", tenantID, orgID).Find(&projects).Order("title").Error; err != nil {
		return nil, err
	}
	return projects, nil
}

// ListProjectsByTenantID lists projects by tenant ID.
func (s *S) ListProjectsByTenantID(tenantID string) ([]*Project, error) {
	var projects []*Project
	if err := s.db.Where("tenant_id = ?", tenantID).Order("title").Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

// ListAllProjects lists all projects.
func (s *S) ListAllProjects() ([]*Project, error) {
	var projects []*Project
	if err := s.db.Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

// DeleteProject deletes an project.
func (s *S) DeleteProject(projectID string) error {
	return DeleteProjectInTransaction(s.db, projectID)
}

// DeleteProjectInTransaction deletes an project in a transaction.
func DeleteProjectInTransaction(tx *gorm.DB, projectID string) error {
	res := tx.Unscoped().Where("project_id = ?", projectID).Delete(&Project{})
	if err := res.Error; err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdateProject updates an existing project.
func (s *S) UpdateProject(
	projectID string,
	updates map[string]interface{},
) error {
	res := s.db.Model(&Project{}).Where("project_id = ?", projectID).
		Updates(updates)
	if err := res.Error; err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
