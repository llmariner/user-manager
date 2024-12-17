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
  createdAt?: string
  user?: User
  organization?: Organization
  project?: Project
  organizationRole?: OrganizationRole
  projectRole?: ProjectRole
}

export type User = {
  id?: string
  internalId?: string
}

export type OrganizationUser = {
  userId?: string
  internalUserId?: string
  organizationId?: string
  role?: OrganizationRole
}

export type OrganizationSummary = {
  projectCount?: number
  userCount?: number
}

export type Organization = {
  id?: string
  title?: string
  createdAt?: string
  summary?: OrganizationSummary
}

export type ProjectUser = {
  userId?: string
  projectId?: string
  organizationId?: string
  role?: ProjectRole
}

export type ProjectSummary = {
  userCount?: number
}

export type Project = {
  id?: string
  title?: string
  kubernetesNamespace?: string
  organizationId?: string
  createdAt?: string
  summary?: ProjectSummary
}

export type CreateAPIKeyRequest = {
  name?: string
  projectId?: string
  organizationId?: string
}

export type ListProjectAPIKeysRequest = {
  projectId?: string
  organizationId?: string
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
  projectId?: string
  organizationId?: string
}

export type DeleteAPIKeyResponse = {
  id?: string
  object?: string
  deleted?: boolean
}

export type UpdateAPIKeyRequest = {
  apiKey?: APIKey
  updateMask?: GoogleProtobufField_mask.FieldMask
}

export type CreateOrganizationRequest = {
  title?: string
}

export type ListOrganizationsRequest = {
  includeSummary?: boolean
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
  organizationId?: string
  userId?: string
  role?: OrganizationRole
}

export type ListOrganizationUsersRequest = {
  organizationId?: string
}

export type ListOrganizationUsersResponse = {
  users?: OrganizationUser[]
}

export type DeleteOrganizationUserRequest = {
  organizationId?: string
  userId?: string
}

export type DeleteOrganizationUserResponse = {
  id?: string
  object?: string
  deleted?: boolean
}

export type CreateProjectRequest = {
  title?: string
  organizationId?: string
  kubernetesNamespace?: string
}

export type ListProjectsRequest = {
  organizationId?: string
  includeSummary?: boolean
}

export type ListProjectsResponse = {
  projects?: Project[]
}

export type DeleteProjectRequest = {
  organizationId?: string
  id?: string
}

export type DeleteProjectResponse = {
  id?: string
  object?: string
  deleted?: boolean
}

export type CreateProjectUserRequest = {
  organizationId?: string
  projectId?: string
  userId?: string
  role?: ProjectRole
}

export type ListProjectUsersRequest = {
  organizationId?: string
  projectId?: string
}

export type ListProjectUsersResponse = {
  users?: ProjectUser[]
}

export type DeleteProjectUserRequest = {
  organizationId?: string
  projectId?: string
  userId?: string
}

export type DeleteProjectUserResponse = {
  id?: string
  object?: string
  deleted?: boolean
}

export type GetUserSelfRequest = {
}

export type InternalAPIKey = {
  apiKey?: APIKey
  tenantId?: string
}

export type ListInternalAPIKeysRequest = {
}

export type ListInternalAPIKeysResponse = {
  apiKeys?: InternalAPIKey[]
}

export type InternalOrganization = {
  organization?: Organization
  tenantId?: string
}

export type ListInternalOrganizationsRequest = {
}

export type ListInternalOrganizationsResponse = {
  organizations?: InternalOrganization[]
}

export type CreateUserInternalRequest = {
  tenantId?: string
  title?: string
  userId?: string
  kubernetesNamespac?: string
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
    return fm.fetchReq<UpdateAPIKeyRequest, APIKey>(`/v1/api_keys`, {...initReq, method: "PATCH", body: JSON.stringify(req)})
  }
  static CreateProjectAPIKey(req: CreateAPIKeyRequest, initReq?: fm.InitReq): Promise<APIKey> {
    return fm.fetchReq<CreateAPIKeyRequest, APIKey>(`/v1/organizations/${req["organizationId"]}/projects/${req["projectId"]}/api_keys`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListProjectAPIKeys(req: ListProjectAPIKeysRequest, initReq?: fm.InitReq): Promise<ListAPIKeysResponse> {
    return fm.fetchReq<ListProjectAPIKeysRequest, ListAPIKeysResponse>(`/v1/organizations/${req["organizationId"]}/projects/${req["projectId"]}/api_keys?${fm.renderURLSearchParams(req, ["organizationId", "projectId"])}`, {...initReq, method: "GET"})
  }
  static DeleteProjectAPIKey(req: DeleteProjectAPIKeyRequest, initReq?: fm.InitReq): Promise<DeleteAPIKeyResponse> {
    return fm.fetchReq<DeleteProjectAPIKeyRequest, DeleteAPIKeyResponse>(`/v1/organizations/${req["organizationId"]}/projects/${req["projectId"]}/api_keys/${req["id"]}`, {...initReq, method: "DELETE"})
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
    return fm.fetchReq<CreateOrganizationUserRequest, OrganizationUser>(`/v1/organizations/${req["organizationId"]}/users`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListOrganizationUsers(req: ListOrganizationUsersRequest, initReq?: fm.InitReq): Promise<ListOrganizationUsersResponse> {
    return fm.fetchReq<ListOrganizationUsersRequest, ListOrganizationUsersResponse>(`/v1/organizations/${req["organizationId"]}/users?${fm.renderURLSearchParams(req, ["organizationId"])}`, {...initReq, method: "GET"})
  }
  static DeleteOrganizationUser(req: DeleteOrganizationUserRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
    return fm.fetchReq<DeleteOrganizationUserRequest, GoogleProtobufEmpty.Empty>(`/v1/organizations/${req["organizationId"]}/users/${req["userId"]}`, {...initReq, method: "DELETE"})
  }
  static CreateProject(req: CreateProjectRequest, initReq?: fm.InitReq): Promise<Project> {
    return fm.fetchReq<CreateProjectRequest, Project>(`/v1/organizations/${req["organizationId"]}/projects`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListProjects(req: ListProjectsRequest, initReq?: fm.InitReq): Promise<ListProjectsResponse> {
    return fm.fetchReq<ListProjectsRequest, ListProjectsResponse>(`/v1/organizations/${req["organizationId"]}/projects?${fm.renderURLSearchParams(req, ["organizationId"])}`, {...initReq, method: "GET"})
  }
  static DeleteProject(req: DeleteProjectRequest, initReq?: fm.InitReq): Promise<DeleteProjectResponse> {
    return fm.fetchReq<DeleteProjectRequest, DeleteProjectResponse>(`/v1/organizations/${req["organizationId"]}/projects/${req["id"]}`, {...initReq, method: "DELETE"})
  }
  static CreateProjectUser(req: CreateProjectUserRequest, initReq?: fm.InitReq): Promise<ProjectUser> {
    return fm.fetchReq<CreateProjectUserRequest, ProjectUser>(`/v1/organizations/${req["organizationId"]}/projects/${req["projectId"]}/users`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListProjectUsers(req: ListProjectUsersRequest, initReq?: fm.InitReq): Promise<ListProjectUsersResponse> {
    return fm.fetchReq<ListProjectUsersRequest, ListProjectUsersResponse>(`/v1/organizations/${req["organizationId"]}/projects/${req["projectId"]}/users?${fm.renderURLSearchParams(req, ["organizationId", "projectId"])}`, {...initReq, method: "GET"})
  }
  static DeleteProjectUser(req: DeleteProjectUserRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
    return fm.fetchReq<DeleteProjectUserRequest, GoogleProtobufEmpty.Empty>(`/v1/organizations/${req["organizationId"]}/projects/${req["projectId"]}/users/${req["userId"]}`, {...initReq, method: "DELETE"})
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
  static CreateUserInternal(req: CreateUserInternalRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
    return fm.fetchReq<CreateUserInternalRequest, GoogleProtobufEmpty.Empty>(`/llmariner.users.server.v1.UsersInternalService/CreateUserInternal`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
}