import * as fm from "../../fetch.pb";
import * as LlmarinerUsersServerV1User_manager_service from "./user_manager_service.pb";
export declare class UsersInternalService {
    static ListInternalAPIKeys(req: LlmarinerUsersServerV1User_manager_service.ListInternalAPIKeysRequest, initReq?: fm.InitReq): Promise<LlmarinerUsersServerV1User_manager_service.ListInternalAPIKeysResponse>;
    static ListInternalOrganizations(req: LlmarinerUsersServerV1User_manager_service.ListInternalOrganizationsRequest, initReq?: fm.InitReq): Promise<LlmarinerUsersServerV1User_manager_service.ListInternalOrganizationsResponse>;
    static ListOrganizationUsers(req: LlmarinerUsersServerV1User_manager_service.ListOrganizationUsersRequest, initReq?: fm.InitReq): Promise<LlmarinerUsersServerV1User_manager_service.ListOrganizationUsersResponse>;
    static ListProjects(req: LlmarinerUsersServerV1User_manager_service.ListProjectsRequest, initReq?: fm.InitReq): Promise<LlmarinerUsersServerV1User_manager_service.ListProjectsResponse>;
    static ListProjectUsers(req: LlmarinerUsersServerV1User_manager_service.ListProjectUsersRequest, initReq?: fm.InitReq): Promise<LlmarinerUsersServerV1User_manager_service.ListProjectUsersResponse>;
}
