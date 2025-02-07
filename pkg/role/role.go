package role

import (
	uv1 "github.com/llmariner/user-manager/api/v1"
)

const (
	// Owner is the owner role.
	Owner = "owner"
	// Reader is the reader role.
	Reader = "reader"
	// TenantSystem is the tenant-system role.
	TenantSystem = "tenant-system"
	// Member is the member role.
	Member = "member"
)

// OrganizationRoleToString converts an organization role to a string.
func OrganizationRoleToString(role uv1.OrganizationRole) (string, bool) {
	switch role {
	case uv1.OrganizationRole_ORGANIZATION_ROLE_OWNER:
		return Owner, true
	case uv1.OrganizationRole_ORGANIZATION_ROLE_READER:
		return Reader, true
	case uv1.OrganizationRole_ORGANIZATION_ROLE_TENANT_SYSTEM:
		return TenantSystem, true
	}
	return "", false
}

// OrganizationRoleToProtoEnum converts a string to an organization role.
func OrganizationRoleToProtoEnum(role string) (uv1.OrganizationRole, bool) {
	switch role {
	case Owner:
		return uv1.OrganizationRole_ORGANIZATION_ROLE_OWNER, true
	case Reader:
		return uv1.OrganizationRole_ORGANIZATION_ROLE_READER, true
	case TenantSystem:
		return uv1.OrganizationRole_ORGANIZATION_ROLE_TENANT_SYSTEM, true
	}
	return uv1.OrganizationRole_ORGANIZATION_ROLE_UNSPECIFIED, false
}

// ProjectRoleToString converts an project role to a string.
func ProjectRoleToString(role uv1.ProjectRole) (string, bool) {
	switch role {
	case uv1.ProjectRole_PROJECT_ROLE_OWNER:
		return Owner, true
	case uv1.ProjectRole_PROJECT_ROLE_MEMBER:
		return Member, true
	}
	return "", false
}

// ProjectRoleToProtoEnum converts a string to an project role.
func ProjectRoleToProtoEnum(role string) (uv1.ProjectRole, bool) {
	switch role {
	case Owner:
		return uv1.ProjectRole_PROJECT_ROLE_OWNER, true
	case Member:
		return uv1.ProjectRole_PROJECT_ROLE_MEMBER, true
	}
	return uv1.ProjectRole_PROJECT_ROLE_UNSPECIFIED, false
}
