/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../../fetch.pb"
import * as GoogleProtobufEmpty from "../../google/protobuf/empty.pb"
import * as GoogleProtobufField_mask from "../../google/protobuf/field_mask.pb"

export enum OrganizationRole {
  ORGANIZATION_ROLE_UNSPECIFIED = "ORGANIZATION_ROLE_UNSPECIFIED",
  ORGANIZATION_ROLE_OWNER = "ORGANIZATION_ROLE_OWNER",
  ORGANIZATION_ROLE_READER = "ORGANIZATION_ROLE_READER",
  ORGANIZATION_ROLE_TENANT_SYSTEM = "ORGANIZATION_ROLE_TENANT_SYSTEM",
}

export enum ProjectRole {
  PROJECT_ROLE_UNSPECIFIED = "PROJECT_ROLE_UNSPECIFIED",
  PROJECT_ROLE_OWNER = "PROJECT_ROLE_OWNER",
  PROJECT_ROLE_MEMBER = "PROJECT_ROLE_MEMBER",
}

export type APIKey = {
  id?: string
  object?: string
  name?: string
  secret?: string
  created_at?: string
  user?: User
  organization?: Organization
  project?: Project
  organization_role?: OrganizationRole
  project_role?: ProjectRole
  excluded_from_rate_limiting?: boolean
}

export type User = {
  id?: string
  internal_id?: string
  is_service_account?: boolean
  hidden?: boolean
}

export type OrganizationUser = {
  user_id?: string
  internal_user_id?: string
  organization_id?: string
  role?: OrganizationRole
}

export type OrganizationSummary = {
  project_count?: number
  user_count?: number
}

export type Organization = {
  id?: string
  title?: string
  created_at?: string
  summary?: OrganizationSummary
  is_default?: boolean
}

export type ProjectUser = {
  user_id?: string
  project_id?: string
  organization_id?: string
  role?: ProjectRole
}

export type ProjectAssignmentNodeSelector = {
  key?: string
  value?: string
}

export type ProjectAssignment = {
  cluster_id?: string
  namespace?: string
  kueue_queue_name?: string
  node_selector?: ProjectAssignmentNodeSelector[]
}

export type ProjectAssignments = {
  assignments?: ProjectAssignment[]
}

export type ProjectSummary = {
  user_count?: number
}

export type Project = {
  id?: string
  title?: string
  assignments?: ProjectAssignment[]
  kubernetes_namespace?: string
  organization_id?: string
  created_at?: string
  summary?: ProjectSummary
  is_default?: boolean
}

export type CreateAPIKeyRequest = {
  name?: string
  project_id?: string
  organization_id?: string
  is_service_account?: boolean
  role?: OrganizationRole
  excluded_from_rate_limiting?: boolean
}

export type ListProjectAPIKeysRequest = {
  project_id?: string
  organization_id?: string
}

export type ListAPIKeysRequest = {
}

export type ListAPIKeysResponse = {
  object?: string
  data?: APIKey[]
}

export type DeleteAPIKeyRequest = {
  id?: string
}

export type DeleteProjectAPIKeyRequest = {
  id?: string
  project_id?: string
  organization_id?: string
}

export type DeleteAPIKeyResponse = {
  id?: string
  object?: string
  deleted?: boolean
}

export type UpdateAPIKeyRequest = {
  api_key?: APIKey
  update_mask?: GoogleProtobufField_mask.FieldMask
}

export type CreateOrganizationRequest = {
  title?: string
}

export type ListOrganizationsRequest = {
  include_summary?: boolean
}

export type ListOrganizationsResponse = {
  organizations?: Organization[]
}

export type DeleteOrganizationRequest = {
  id?: string
}

export type DeleteOrganizationResponse = {
  id?: string
  object?: string
  deleted?: boolean
}

export type CreateOrganizationUserRequest = {
  organization_id?: string
  user_id?: string
  role?: OrganizationRole
}

export type ListOrganizationUsersRequest = {
  organization_id?: string
}

export type ListOrganizationUsersResponse = {
  users?: OrganizationUser[]
}

export type DeleteOrganizationUserRequest = {
  organization_id?: string
  user_id?: string
}

export type DeleteOrganizationUserResponse = {
  id?: string
  object?: string
  deleted?: boolean
}

export type CreateProjectRequest = {
  title?: string
  organization_id?: string
  kubernetes_namespace?: string
  assignments?: ProjectAssignment[]
}

export type ListProjectsRequest = {
  organization_id?: string
  include_summary?: boolean
}

export type ListProjectsResponse = {
  projects?: Project[]
}

export type DeleteProjectRequest = {
  organization_id?: string
  id?: string
}

export type DeleteProjectResponse = {
  id?: string
  object?: string
  deleted?: boolean
}

export type UpdateProjectRequest = {
  project?: Project
  update_mask?: GoogleProtobufField_mask.FieldMask
}

export type CreateProjectUserRequest = {
  organization_id?: string
  project_id?: string
  user_id?: string
  role?: ProjectRole
}

export type ListProjectUsersRequest = {
  organization_id?: string
  project_id?: string
}

export type ListProjectUsersResponse = {
  users?: ProjectUser[]
}

export type DeleteProjectUserRequest = {
  organization_id?: string
  project_id?: string
  user_id?: string
}

export type DeleteProjectUserResponse = {
  id?: string
  object?: string
  deleted?: boolean
}

export type GetUserSelfRequest = {
}

export type InternalAPIKey = {
  api_key?: APIKey
  tenant_id?: string
}

export type ListInternalAPIKeysRequest = {
}

export type ListInternalAPIKeysResponse = {
  api_keys?: InternalAPIKey[]
}

export type InternalOrganization = {
  organization?: Organization
  tenant_id?: string
}

export type ListInternalOrganizationsRequest = {
}

export type ListInternalOrganizationsResponse = {
  organizations?: InternalOrganization[]
}

export type ListUsersRequest = {
}

export type ListUsersResponse = {
  users?: User[]
}

export type CreateUserInternalRequest = {
  tenant_id?: string
  title?: string
  user_id?: string
  kubernetes_namespace?: string
}

export class UsersService {
  static CreateAPIKey(req: CreateAPIKeyRequest, initReq?: fm.InitReq): Promise<APIKey> {
    return fm.fetchReq<CreateAPIKeyRequest, APIKey>(`/v1/api_keys`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListAPIKeys(req: ListAPIKeysRequest, initReq?: fm.InitReq): Promise<ListAPIKeysResponse> {
    return fm.fetchReq<ListAPIKeysRequest, ListAPIKeysResponse>(`/v1/api_keys?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static DeleteAPIKey(req: DeleteAPIKeyRequest, initReq?: fm.InitReq): Promise<DeleteAPIKeyResponse> {
    return fm.fetchReq<DeleteAPIKeyRequest, DeleteAPIKeyResponse>(`/v1/api_keys/${req["id"]}`, {...initReq, method: "DELETE"})
  }
  static UpdateAPIKey(req: UpdateAPIKeyRequest, initReq?: fm.InitReq): Promise<APIKey> {
    return fm.fetchReq<UpdateAPIKeyRequest, APIKey>(`/v1/api_keys/${req["api_key.id"]}`, {...initReq, method: "PATCH", body: JSON.stringify(req["api_key"])})
  }
  static CreateProjectAPIKey(req: CreateAPIKeyRequest, initReq?: fm.InitReq): Promise<APIKey> {
    return fm.fetchReq<CreateAPIKeyRequest, APIKey>(`/v1/organizations/${req["organization_id"]}/projects/${req["project_id"]}/api_keys`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListProjectAPIKeys(req: ListProjectAPIKeysRequest, initReq?: fm.InitReq): Promise<ListAPIKeysResponse> {
    return fm.fetchReq<ListProjectAPIKeysRequest, ListAPIKeysResponse>(`/v1/organizations/${req["organization_id"]}/projects/${req["project_id"]}/api_keys?${fm.renderURLSearchParams(req, ["organization_id", "project_id"])}`, {...initReq, method: "GET"})
  }
  static DeleteProjectAPIKey(req: DeleteProjectAPIKeyRequest, initReq?: fm.InitReq): Promise<DeleteAPIKeyResponse> {
    return fm.fetchReq<DeleteProjectAPIKeyRequest, DeleteAPIKeyResponse>(`/v1/organizations/${req["organization_id"]}/projects/${req["project_id"]}/api_keys/${req["id"]}`, {...initReq, method: "DELETE"})
  }
  static CreateOrganization(req: CreateOrganizationRequest, initReq?: fm.InitReq): Promise<Organization> {
    return fm.fetchReq<CreateOrganizationRequest, Organization>(`/v1/organizations`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListOrganizations(req: ListOrganizationsRequest, initReq?: fm.InitReq): Promise<ListOrganizationsResponse> {
    return fm.fetchReq<ListOrganizationsRequest, ListOrganizationsResponse>(`/v1/organizations?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static DeleteOrganization(req: DeleteOrganizationRequest, initReq?: fm.InitReq): Promise<DeleteOrganizationResponse> {
    return fm.fetchReq<DeleteOrganizationRequest, DeleteOrganizationResponse>(`/v1/organizations/${req["id"]}`, {...initReq, method: "DELETE"})
  }
  static CreateOrganizationUser(req: CreateOrganizationUserRequest, initReq?: fm.InitReq): Promise<OrganizationUser> {
    return fm.fetchReq<CreateOrganizationUserRequest, OrganizationUser>(`/v1/organizations/${req["organization_id"]}/users`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListOrganizationUsers(req: ListOrganizationUsersRequest, initReq?: fm.InitReq): Promise<ListOrganizationUsersResponse> {
    return fm.fetchReq<ListOrganizationUsersRequest, ListOrganizationUsersResponse>(`/v1/organizations/${req["organization_id"]}/users?${fm.renderURLSearchParams(req, ["organization_id"])}`, {...initReq, method: "GET"})
  }
  static DeleteOrganizationUser(req: DeleteOrganizationUserRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
    return fm.fetchReq<DeleteOrganizationUserRequest, GoogleProtobufEmpty.Empty>(`/v1/organizations/${req["organization_id"]}/users/${req["user_id"]}`, {...initReq, method: "DELETE"})
  }
  static CreateProject(req: CreateProjectRequest, initReq?: fm.InitReq): Promise<Project> {
    return fm.fetchReq<CreateProjectRequest, Project>(`/v1/organizations/${req["organization_id"]}/projects`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListProjects(req: ListProjectsRequest, initReq?: fm.InitReq): Promise<ListProjectsResponse> {
    return fm.fetchReq<ListProjectsRequest, ListProjectsResponse>(`/v1/organizations/${req["organization_id"]}/projects?${fm.renderURLSearchParams(req, ["organization_id"])}`, {...initReq, method: "GET"})
  }
  static DeleteProject(req: DeleteProjectRequest, initReq?: fm.InitReq): Promise<DeleteProjectResponse> {
    return fm.fetchReq<DeleteProjectRequest, DeleteProjectResponse>(`/v1/organizations/${req["organization_id"]}/projects/${req["id"]}`, {...initReq, method: "DELETE"})
  }
  static UpdateProject(req: UpdateProjectRequest, initReq?: fm.InitReq): Promise<Project> {
    return fm.fetchReq<UpdateProjectRequest, Project>(`/v1/organizations/${req["project.organization_id"]}/projects/${req["project.id"]}`, {...initReq, method: "PATCH", body: JSON.stringify(req["project"])})
  }
  static CreateProjectUser(req: CreateProjectUserRequest, initReq?: fm.InitReq): Promise<ProjectUser> {
    return fm.fetchReq<CreateProjectUserRequest, ProjectUser>(`/v1/organizations/${req["organization_id"]}/projects/${req["project_id"]}/users`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListProjectUsers(req: ListProjectUsersRequest, initReq?: fm.InitReq): Promise<ListProjectUsersResponse> {
    return fm.fetchReq<ListProjectUsersRequest, ListProjectUsersResponse>(`/v1/organizations/${req["organization_id"]}/projects/${req["project_id"]}/users?${fm.renderURLSearchParams(req, ["organization_id", "project_id"])}`, {...initReq, method: "GET"})
  }
  static DeleteProjectUser(req: DeleteProjectUserRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
    return fm.fetchReq<DeleteProjectUserRequest, GoogleProtobufEmpty.Empty>(`/v1/organizations/${req["organization_id"]}/projects/${req["project_id"]}/users/${req["user_id"]}`, {...initReq, method: "DELETE"})
  }
  static GetUserSelf(req: GetUserSelfRequest, initReq?: fm.InitReq): Promise<User> {
    return fm.fetchReq<GetUserSelfRequest, User>(`/v1/users:getSelf?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
}
export class UsersInternalService {
  static ListInternalAPIKeys(req: ListInternalAPIKeysRequest, initReq?: fm.InitReq): Promise<ListInternalAPIKeysResponse> {
    return fm.fetchReq<ListInternalAPIKeysRequest, ListInternalAPIKeysResponse>(`/llmariner.users.server.v1.UsersInternalService/ListInternalAPIKeys`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListInternalOrganizations(req: ListInternalOrganizationsRequest, initReq?: fm.InitReq): Promise<ListInternalOrganizationsResponse> {
    return fm.fetchReq<ListInternalOrganizationsRequest, ListInternalOrganizationsResponse>(`/llmariner.users.server.v1.UsersInternalService/ListInternalOrganizations`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListOrganizationUsers(req: ListOrganizationUsersRequest, initReq?: fm.InitReq): Promise<ListOrganizationUsersResponse> {
    return fm.fetchReq<ListOrganizationUsersRequest, ListOrganizationUsersResponse>(`/llmariner.users.server.v1.UsersInternalService/ListOrganizationUsers`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListProjects(req: ListProjectsRequest, initReq?: fm.InitReq): Promise<ListProjectsResponse> {
    return fm.fetchReq<ListProjectsRequest, ListProjectsResponse>(`/llmariner.users.server.v1.UsersInternalService/ListProjects`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListProjectUsers(req: ListProjectUsersRequest, initReq?: fm.InitReq): Promise<ListProjectUsersResponse> {
    return fm.fetchReq<ListProjectUsersRequest, ListProjectUsersResponse>(`/llmariner.users.server.v1.UsersInternalService/ListProjectUsers`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListUsers(req: ListUsersRequest, initReq?: fm.InitReq): Promise<ListUsersResponse> {
    return fm.fetchReq<ListUsersRequest, ListUsersResponse>(`/llmariner.users.server.v1.UsersInternalService/ListUsers`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static CreateUserInternal(req: CreateUserInternalRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
    return fm.fetchReq<CreateUserInternalRequest, GoogleProtobufEmpty.Empty>(`/llmariner.users.server.v1.UsersInternalService/CreateUserInternal`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
}