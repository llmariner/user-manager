import * as fm from "../../fetch.pb";
import * as GoogleProtobufEmpty from "../../google/protobuf/empty.pb";
export declare enum OrganizationRole {
    ORGANIZATION_ROLE_UNSPECIFIED = "ORGANIZATION_ROLE_UNSPECIFIED",
    ORGANIZATION_ROLE_OWNER = "ORGANIZATION_ROLE_OWNER",
    ORGANIZATION_ROLE_READER = "ORGANIZATION_ROLE_READER"
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
};
export type User = {
    id?: string;
    internalId?: string;
};
export type OrganizationUser = {
    userId?: string;
    internalUserId?: string;
    organizationId?: string;
    role?: OrganizationRole;
};
export type Organization = {
    id?: string;
    title?: string;
    createdAt?: string;
};
export type ProjectUser = {
    userId?: string;
    projectId?: string;
    organizationId?: string;
    role?: ProjectRole;
};
export type Project = {
    id?: string;
    title?: string;
    kubernetesNamespace?: string;
    organizationId?: string;
    createdAt?: string;
};
export type CreateAPIKeyRequest = {
    name?: string;
    projectId?: string;
    organizationId?: string;
};
export type ListAPIKeysRequest = {
    projectId?: string;
    organizationId?: string;
};
export type ListAPIKeysResponse = {
    object?: string;
    data?: APIKey[];
};
export type DeleteAPIKeyRequest = {
    id?: string;
    projectId?: string;
    organizationId?: string;
};
export type DeleteAPIKeyResponse = {
    id?: string;
    object?: string;
    deleted?: boolean;
};
export type CreateOrganizationRequest = {
    title?: string;
};
export type ListOrganizationsRequest = {};
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
};
export type ListProjectsRequest = {
    organizationId?: string;
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
export declare class UsersService {
    static CreateAPIKey(req: CreateAPIKeyRequest, initReq?: fm.InitReq): Promise<APIKey>;
    static ListAPIKeys(req: ListAPIKeysRequest, initReq?: fm.InitReq): Promise<ListAPIKeysResponse>;
    static DeleteAPIKey(req: DeleteAPIKeyRequest, initReq?: fm.InitReq): Promise<DeleteAPIKeyResponse>;
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
}
export declare class UsersInternalService {
    static ListInternalAPIKeys(req: ListInternalAPIKeysRequest, initReq?: fm.InitReq): Promise<ListInternalAPIKeysResponse>;
    static ListInternalOrganizations(req: ListInternalOrganizationsRequest, initReq?: fm.InitReq): Promise<ListInternalOrganizationsResponse>;
    static ListOrganizationUsers(req: ListOrganizationUsersRequest, initReq?: fm.InitReq): Promise<ListOrganizationUsersResponse>;
    static ListProjects(req: ListProjectsRequest, initReq?: fm.InitReq): Promise<ListProjectsResponse>;
    static ListProjectUsers(req: ListProjectUsersRequest, initReq?: fm.InitReq): Promise<ListProjectUsersResponse>;
}
