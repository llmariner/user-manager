package store

import (
	v1 "github.com/llm-operator/user-manager/api/v1"
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
}

// ToProto converts the organization to proto.
func (p *Project) ToProto() *v1.Project {
	return &v1.Project{
		Id:                  p.ProjectID,
		OrganizationId:      p.OrganizationID,
		Title:               p.Title,
		KubernetesNamespace: p.KubernetesNamespace,
		CreatedAt:           p.CreatedAt.UTC().Unix(),
	}
}

// CreateProjectParams is the parameters for CreateProject.xo
type CreateProjectParams struct {
	ProjectID           string
	OrganizationID      string
	TenantID            string
	Title               string
	KubernetesNamespace string
	IsDefault           bool
}

// CreateProject creates a new project.
func (s *S) CreateProject(p CreateProjectParams) (*Project, error) {
	return CreateProjectInTransaction(s.db, p)
}

// CreateProjectInTransaction creates a new project in a transaction.
func CreateProjectInTransaction(tx *gorm.DB, p CreateProjectParams) (*Project, error) {
	project := &Project{
		ProjectID:           p.ProjectID,
		OrganizationID:      p.OrganizationID,
		TenantID:            p.TenantID,
		Title:               p.Title,
		KubernetesNamespace: p.KubernetesNamespace,
		IsDefault:           p.IsDefault,
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
	if err := s.db.Where("tenant_id = ? AND organization_id = ? ", tenantID, orgID).Find(&projects).Error; err != nil {
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
