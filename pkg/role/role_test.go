package role

import (
	"testing"

	uv1 "github.com/llmariner/user-manager/api/v1"
	"github.com/stretchr/testify/assert"
)

func TestOrganizationRoleConversion(t *testing.T) {
	role := uv1.OrganizationRole_ORGANIZATION_ROLE_OWNER
	s, ok := OrganizationRoleToString(role)
	assert.True(t, ok)
	assert.Equal(t, Owner, s)
	got, ok := OrganizationRoleToProtoEnum(s)
	assert.True(t, ok)
	assert.Equal(t, role, got)
}

func TestProjectRoleConversion(t *testing.T) {
	role := uv1.ProjectRole_PROJECT_ROLE_OWNER
	s, ok := ProjectRoleToString(role)
	assert.True(t, ok)
	assert.Equal(t, Owner, s)
	got, ok := ProjectRoleToProtoEnum(s)
	assert.True(t, ok)
	assert.Equal(t, role, got)
}
