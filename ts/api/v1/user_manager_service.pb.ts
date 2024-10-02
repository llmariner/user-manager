/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../../fetch.pb"
import * as GoogleProtobufEmpty from "../../google/protobuf/empty.pb"

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

export type Organization = {
  id?: string
  title?: string
  createdAt?: string
}

export type ProjectUser = {
  userId?: string
  projectId?: string
  organizationId?: string
  role?: ProjectRole
}

export type Project = {
  id?: string
  title?: string
  kubernetesNamespace?: string
  organizationId?: string
  createdAt?: string
}

export type CreateAPIKeyRequest = {
  name?: string
  projectId?: string
  organizationId?: string
}

export type ListAPIKeysRequest = {
  projectId?: string
  organizationId?: string
}

export type ListAPIKeysResponse = {
  object?: string
  data?: APIKey[]
}

export type DeleteAPIKeyRequest = {
  id?: string
  projectId?: string
  organizationId?: string
}

export type DeleteAPIKeyResponse = {
  id?: string
  object?: string
  deleted?: boolean
}

export type CreateOrganizationRequest = {
  title?: string
}

export type ListOrganizationsRequest = {
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

export class UsersService {
  static CreateAPIKey(req: CreateAPIKeyRequest, initReq?: fm.InitReq): Promise<APIKey> {
    return fm.fetchReq<CreateAPIKeyRequest, APIKey>(`/v1/organizations/${req["organizationId"]}/projects/${req["projectId"]}/api_keys`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListAPIKeys(req: ListAPIKeysRequest, initReq?: fm.InitReq): Promise<ListAPIKeysResponse> {
    return fm.fetchReq<ListAPIKeysRequest, ListAPIKeysResponse>(`/v1/organizations/${req["organizationId"]}/projects/${req["projectId"]}/api_keys?${fm.renderURLSearchParams(req, ["organizationId", "projectId"])}`, {...initReq, method: "GET"})
  }
  static DeleteAPIKey(req: DeleteAPIKeyRequest, initReq?: fm.InitReq): Promise<DeleteAPIKeyResponse> {
    return fm.fetchReq<DeleteAPIKeyRequest, DeleteAPIKeyResponse>(`/v1/organizations/${req["organizationId"]}/projects/${req["projectId"]}/api_keys/${req["id"]}`, {...initReq, method: "DELETE"})
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
}