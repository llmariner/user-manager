import * as fm from "../../fetch.pb";
import * as GoogleProtobufEmpty from "../../google/protobuf/empty.pb";
import * as GoogleProtobufField_mask from "../../google/protobuf/field_mask.pb";
export declare enum OrganizationRole {
    ORGANIZATION_ROLE_UNSPECIFIED = "ORGANIZATION_ROLE_UNSPECIFIED",
    ORGANIZATION_ROLE_OWNER = "ORGANIZATION_ROLE_OWNER",
    ORGANIZATION_ROLE_READER = "ORGANIZATION_ROLE_READER",
    ORGANIZATION_ROLE_TENANT_SYSTEM = "ORGANIZATION_ROLE_TENANT_SYSTEM"
}
export declare enum ProjectRole {
    PROJECT_ROLE_UNSPECIFIED = "PROJECT_ROLE_UNSPECIFIED",
    PROJECT_ROLE_OWNER = "PROJECT_ROLE_OWNER",
    PROJECT_ROLE_MEMBER = "PROJECT_ROLE_MEMBER"
}
export type APIKey = {
    id?: string;
    object?: string;
    name?: string;
    secret?: string;
    createdAt?: string;
    user?: User;
    organization?: Organization;
    project?: Project;
    organizationRole?: OrganizationRole;
    projectRole?: ProjectRole;
};
export type User = {
    id?: string;
    internalId?: string;
    isServiceAccount?: boolean;
};
export type OrganizationUser = {
    userId?: string;
    internalUserId?: string;
    organizationId?: string;
    role?: OrganizationRole;
};
export type OrganizationSummary = {
    projectCount?: number;
    userCount?: number;
};
export type Organization = {
    id?: string;
    title?: string;
    createdAt?: string;
    summary?: OrganizationSummary;
};
export type ProjectUser = {
    userId?: string;
    projectId?: string;
    organizationId?: string;
    role?: ProjectRole;
};
export type ProjectAssignment = {
    clusterId?: string;
    namespace?: string;
};
export type ProjectAssignments = {
    assignments?: ProjectAssignment[];
};
export type ProjectSummary = {
    userCount?: number;
};
export type Project = {
    id?: string;
    title?: string;
    assignments?: ProjectAssignment[];
    kubernetesNamespace?: string;
    organizationId?: string;
    createdAt?: string;
    summary?: ProjectSummary;
};
export type CreateAPIKeyRequest = {
    name?: string;
    projectId?: string;
    organizationId?: string;
    isServiceAccount?: boolean;
    role?: OrganizationRole;
};
export type ListProjectAPIKeysRequest = {
    projectId?: string;
    organizationId?: string;
};
export type ListAPIKeysRequest = {};
export type ListAPIKeysResponse = {
    object?: string;
    data?: APIKey[];
};
export type DeleteAPIKeyRequest = {
    id?: string;
};
export type DeleteProjectAPIKeyRequest = {
    id?: string;
    projectId?: string;
    organizationId?: string;
};
export type DeleteAPIKeyResponse = {
    id?: string;
    object?: string;
    deleted?: boolean;
};
export type UpdateAPIKeyRequest = {
    apiKey?: APIKey;
    updateMask?: GoogleProtobufField_mask.FieldMask;
};
export type CreateOrganizationRequest = {
    title?: string;
};
export type ListOrganizationsRequest = {
    includeSummary?: boolean;
};
export type ListOrganizationsResponse = {
    organizations?: Organization[];
};
export type DeleteOrganizationRequest = {
    id?: string;
};
export type DeleteOrganizationResponse = {
    id?: string;
    object?: string;
    deleted?: boolean;
};
export type CreateOrganizationUserRequest = {
    organizationId?: string;
    userId?: string;
    role?: OrganizationRole;
};
export type ListOrganizationUsersRequest = {
    organizationId?: string;
};
export type ListOrganizationUsersResponse = {
    users?: OrganizationUser[];
};
export type DeleteOrganizationUserRequest = {
    organizationId?: string;
    userId?: string;
};
export type DeleteOrganizationUserResponse = {
    id?: string;
    object?: string;
    deleted?: boolean;
};
export type CreateProjectRequest = {
    title?: string;
    organizationId?: string;
    kubernetesNamespace?: string;
    assignments?: ProjectAssignment[];
};
export type ListProjectsRequest = {
    organizationId?: string;
    includeSummary?: boolean;
};
export type ListProjectsResponse = {
    projects?: Project[];
};
export type DeleteProjectRequest = {
    organizationId?: string;
    id?: string;
};
export type DeleteProjectResponse = {
    id?: string;
    object?: string;
    deleted?: boolean;
};
export type CreateProjectUserRequest = {
    organizationId?: string;
    projectId?: string;
    userId?: string;
    role?: ProjectRole;
};
export type ListProjectUsersRequest = {
    organizationId?: string;
    projectId?: string;
};
export type ListProjectUsersResponse = {
    users?: ProjectUser[];
};
export type DeleteProjectUserRequest = {
    organizationId?: string;
    projectId?: string;
    userId?: string;
};
export type DeleteProjectUserResponse = {
    id?: string;
    object?: string;
    deleted?: boolean;
};
export type GetUserSelfRequest = {};
export type InternalAPIKey = {
    apiKey?: APIKey;
    tenantId?: string;
};
export type ListInternalAPIKeysRequest = {};
export type ListInternalAPIKeysResponse = {
    apiKeys?: InternalAPIKey[];
};
export type InternalOrganization = {
    organization?: Organization;
    tenantId?: string;
};
export type ListInternalOrganizationsRequest = {};
export type ListInternalOrganizationsResponse = {
    organizations?: InternalOrganization[];
};
export type ListUsersRequest = {};
export type ListUsersResponse = {
    users?: User[];
};
export type CreateUserInternalRequest = {
    tenantId?: string;
    title?: string;
    userId?: string;
    kubernetesNamespace?: string;
};
export declare class UsersService {
    static CreateAPIKey(req: CreateAPIKeyRequest, initReq?: fm.InitReq): Promise<APIKey>;
    static ListAPIKeys(req: ListAPIKeysRequest, initReq?: fm.InitReq): Promise<ListAPIKeysResponse>;
    static DeleteAPIKey(req: DeleteAPIKeyRequest, initReq?: fm.InitReq): Promise<DeleteAPIKeyResponse>;
    static UpdateAPIKey(req: UpdateAPIKeyRequest, initReq?: fm.InitReq): Promise<APIKey>;
    static CreateProjectAPIKey(req: CreateAPIKeyRequest, initReq?: fm.InitReq): Promise<APIKey>;
    static ListProjectAPIKeys(req: ListProjectAPIKeysRequest, initReq?: fm.InitReq): Promise<ListAPIKeysResponse>;
    static DeleteProjectAPIKey(req: DeleteProjectAPIKeyRequest, initReq?: fm.InitReq): Promise<DeleteAPIKeyResponse>;
    static CreateOrganization(req: CreateOrganizationRequest, initReq?: fm.InitReq): Promise<Organization>;
    static ListOrganizations(req: ListOrganizationsRequest, initReq?: fm.InitReq): Promise<ListOrganizationsResponse>;
    static DeleteOrganization(req: DeleteOrganizationRequest, initReq?: fm.InitReq): Promise<DeleteOrganizationResponse>;
    static CreateOrganizationUser(req: CreateOrganizationUserRequest, initReq?: fm.InitReq): Promise<OrganizationUser>;
    static ListOrganizationUsers(req: ListOrganizationUsersRequest, initReq?: fm.InitReq): Promise<ListOrganizationUsersResponse>;
    static DeleteOrganizationUser(req: DeleteOrganizationUserRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty>;
    static CreateProject(req: CreateProjectRequest, initReq?: fm.InitReq): Promise<Project>;
    static ListProjects(req: ListProjectsRequest, initReq?: fm.InitReq): Promise<ListProjectsResponse>;
    static DeleteProject(req: DeleteProjectRequest, initReq?: fm.InitReq): Promise<DeleteProjectResponse>;
    static CreateProjectUser(req: CreateProjectUserRequest, initReq?: fm.InitReq): Promise<ProjectUser>;
    static ListProjectUsers(req: ListProjectUsersRequest, initReq?: fm.InitReq): Promise<ListProjectUsersResponse>;
    static DeleteProjectUser(req: DeleteProjectUserRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty>;
    static GetUserSelf(req: GetUserSelfRequest, initReq?: fm.InitReq): Promise<User>;
}
export declare class UsersInternalService {
    static ListInternalAPIKeys(req: ListInternalAPIKeysRequest, initReq?: fm.InitReq): Promise<ListInternalAPIKeysResponse>;
    static ListInternalOrganizations(req: ListInternalOrganizationsRequest, initReq?: fm.InitReq): Promise<ListInternalOrganizationsResponse>;
    static ListOrganizationUsers(req: ListOrganizationUsersRequest, initReq?: fm.InitReq): Promise<ListOrganizationUsersResponse>;
    static ListProjects(req: ListProjectsRequest, initReq?: fm.InitReq): Promise<ListProjectsResponse>;
    static ListProjectUsers(req: ListProjectUsersRequest, initReq?: fm.InitReq): Promise<ListProjectUsersResponse>;
    static ListUsers(req: ListUsersRequest, initReq?: fm.InitReq): Promise<ListUsersResponse>;
    static CreateUserInternal(req: CreateUserInternalRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty>;
}
