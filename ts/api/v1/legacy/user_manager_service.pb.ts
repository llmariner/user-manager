/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../../../fetch.pb"
import * as LlmarinerUsersServerV1User_manager_service from "../user_manager_service.pb"
export class UsersInternalService {
  static ListInternalAPIKeys(req: LlmarinerUsersServerV1User_manager_service.ListInternalAPIKeysRequest, initReq?: fm.InitReq): Promise<LlmarinerUsersServerV1User_manager_service.ListInternalAPIKeysResponse> {
    return fm.fetchReq<LlmarinerUsersServerV1User_manager_service.ListInternalAPIKeysRequest, LlmarinerUsersServerV1User_manager_service.ListInternalAPIKeysResponse>(`/llmoperator.users.server.v1.UsersInternalService/ListInternalAPIKeys`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListInternalOrganizations(req: LlmarinerUsersServerV1User_manager_service.ListInternalOrganizationsRequest, initReq?: fm.InitReq): Promise<LlmarinerUsersServerV1User_manager_service.ListInternalOrganizationsResponse> {
    return fm.fetchReq<LlmarinerUsersServerV1User_manager_service.ListInternalOrganizationsRequest, LlmarinerUsersServerV1User_manager_service.ListInternalOrganizationsResponse>(`/llmoperator.users.server.v1.UsersInternalService/ListInternalOrganizations`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListOrganizationUsers(req: LlmarinerUsersServerV1User_manager_service.ListOrganizationUsersRequest, initReq?: fm.InitReq): Promise<LlmarinerUsersServerV1User_manager_service.ListOrganizationUsersResponse> {
    return fm.fetchReq<LlmarinerUsersServerV1User_manager_service.ListOrganizationUsersRequest, LlmarinerUsersServerV1User_manager_service.ListOrganizationUsersResponse>(`/llmoperator.users.server.v1.UsersInternalService/ListOrganizationUsers`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListProjects(req: LlmarinerUsersServerV1User_manager_service.ListProjectsRequest, initReq?: fm.InitReq): Promise<LlmarinerUsersServerV1User_manager_service.ListProjectsResponse> {
    return fm.fetchReq<LlmarinerUsersServerV1User_manager_service.ListProjectsRequest, LlmarinerUsersServerV1User_manager_service.ListProjectsResponse>(`/llmoperator.users.server.v1.UsersInternalService/ListProjects`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListProjectUsers(req: LlmarinerUsersServerV1User_manager_service.ListProjectUsersRequest, initReq?: fm.InitReq): Promise<LlmarinerUsersServerV1User_manager_service.ListProjectUsersResponse> {
    return fm.fetchReq<LlmarinerUsersServerV1User_manager_service.ListProjectUsersRequest, LlmarinerUsersServerV1User_manager_service.ListProjectUsersResponse>(`/llmoperator.users.server.v1.UsersInternalService/ListProjectUsers`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
}