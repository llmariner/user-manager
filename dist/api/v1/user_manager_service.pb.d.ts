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
    created_at?: string;
    user?: User;
    organization?: Organization;
    project?: Project;
    organization_role?: OrganizationRole;
    project_role?: ProjectRole;
    excluded_from_rate_limiting?: boolean;
};
export type UserOrganizationRoleBinding = {
    organization_id?: string;
    role?: OrganizationRole;
};
export type UserProjectRoleBinding = {
    organization_id?: string;
    project_id?: string;
    role?: ProjectRole;
};
export type User = {
    id?: string;
    internal_id?: string;
    is_service_account?: boolean;
    hidden?: boolean;
    organization_role_bindings?: UserOrganizationRoleBinding[];
    project_role_bindings?: UserProjectRoleBinding[];
};
export type OrganizationUser = {
    user_id?: string;
    internal_user_id?: string;
    organization_id?: string;
    role?: OrganizationRole;
};
export type OrganizationSummary = {
    project_count?: number;
    user_count?: number;
};
export type Organization = {
    id?: string;
    title?: string;
    created_at?: string;
    summary?: OrganizationSummary;
    is_default?: boolean;
};
export type ProjectUser = {
    user_id?: string;
    project_id?: string;
    organization_id?: string;
    role?: ProjectRole;
};
export type ProjectAssignmentNodeSelector = {
    key?: string;
    value?: string;
};
export type ProjectAssignment = {
    cluster_id?: string;
    namespace?: string;
    kueue_queue_name?: string;
    node_selector?: ProjectAssignmentNodeSelector[];
};
export type ProjectAssignments = {
    assignments?: ProjectAssignment[];
};
export type ProjectSummary = {
    user_count?: number;
};
export type Project = {
    id?: string;
    title?: string;
    assignments?: ProjectAssignment[];
    kubernetes_namespace?: string;
    organization_id?: string;
    created_at?: string;
    summary?: ProjectSummary;
    is_default?: boolean;
};
export type CreateAPIKeyRequest = {
    name?: string;
    project_id?: string;
    organization_id?: string;
    is_service_account?: boolean;
    role?: OrganizationRole;
    excluded_from_rate_limiting?: boolean;
};
export type ListProjectAPIKeysRequest = {
    project_id?: string;
    organization_id?: string;
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
    project_id?: string;
    organization_id?: string;
};
export type DeleteAPIKeyResponse = {
    id?: string;
    object?: string;
    deleted?: boolean;
};
export type UpdateAPIKeyRequest = {
    api_key?: APIKey;
    update_mask?: GoogleProtobufField_mask.FieldMask;
};
export type CreateOrganizationRequest = {
    title?: string;
};
export type ListOrganizationsRequest = {
    include_summary?: boolean;
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
    organization_id?: string;
    user_id?: string;
    role?: OrganizationRole;
};
export type ListOrganizationUsersRequest = {
    organization_id?: string;
};
export type ListOrganizationUsersResponse = {
    users?: OrganizationUser[];
};
export type DeleteOrganizationUserRequest = {
    organization_id?: string;
    user_id?: string;
};
export type DeleteOrganizationUserResponse = {
    id?: string;
    object?: string;
    deleted?: boolean;
};
export type CreateProjectRequest = {
    title?: string;
    organization_id?: string;
    kubernetes_namespace?: string;
    assignments?: ProjectAssignment[];
};
export type ListProjectsRequest = {
    organization_id?: string;
    include_summary?: boolean;
};
export type ListProjectsResponse = {
    projects?: Project[];
};
export type DeleteProjectRequest = {
    organization_id?: string;
    id?: string;
};
export type DeleteProjectResponse = {
    id?: string;
    object?: string;
    deleted?: boolean;
};
export type UpdateProjectRequest = {
    project?: Project;
    update_mask?: GoogleProtobufField_mask.FieldMask;
};
export type CreateProjectUserRequest = {
    organization_id?: string;
    project_id?: string;
    user_id?: string;
    role?: ProjectRole;
};
export type ListProjectUsersRequest = {
    organization_id?: string;
    project_id?: string;
};
export type ListProjectUsersResponse = {
    users?: ProjectUser[];
};
export type DeleteProjectUserRequest = {
    organization_id?: string;
    project_id?: string;
    user_id?: string;
};
export type DeleteProjectUserResponse = {
    id?: string;
    object?: string;
    deleted?: boolean;
};
export type GetUserSelfRequest = {};
export type ListUsersRequest = {};
export type ListUsersResponse = {
    users?: User[];
};
export type InternalAPIKey = {
    api_key?: APIKey;
    tenant_id?: string;
};
export type ListInternalAPIKeysRequest = {};
export type ListInternalAPIKeysResponse = {
    api_keys?: InternalAPIKey[];
};
export type InternalOrganization = {
    organization?: Organization;
    tenant_id?: string;
};
export type ListInternalOrganizationsRequest = {};
export type ListInternalOrganizationsResponse = {
    organizations?: InternalOrganization[];
};
export type CreateUserInternalRequest = {
    tenant_id?: string;
    title?: string;
    user_id?: string;
    kubernetes_namespace?: string;
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
    static UpdateProject(req: UpdateProjectRequest, initReq?: fm.InitReq): Promise<Project>;
    static CreateProjectUser(req: CreateProjectUserRequest, initReq?: fm.InitReq): Promise<ProjectUser>;
    static ListProjectUsers(req: ListProjectUsersRequest, initReq?: fm.InitReq): Promise<ListProjectUsersResponse>;
    static DeleteProjectUser(req: DeleteProjectUserRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty>;
    static GetUserSelf(req: GetUserSelfRequest, initReq?: fm.InitReq): Promise<User>;
    static ListUsers(req: ListUsersRequest, initReq?: fm.InitReq): Promise<ListUsersResponse>;
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
