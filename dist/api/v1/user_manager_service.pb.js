/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/
import * as fm from "../../fetch.pb";
export var OrganizationRole;
(function (OrganizationRole) {
    OrganizationRole["ORGANIZATION_ROLE_UNSPECIFIED"] = "ORGANIZATION_ROLE_UNSPECIFIED";
    OrganizationRole["ORGANIZATION_ROLE_OWNER"] = "ORGANIZATION_ROLE_OWNER";
    OrganizationRole["ORGANIZATION_ROLE_READER"] = "ORGANIZATION_ROLE_READER";
})(OrganizationRole || (OrganizationRole = {}));
export var ProjectRole;
(function (ProjectRole) {
    ProjectRole["PROJECT_ROLE_UNSPECIFIED"] = "PROJECT_ROLE_UNSPECIFIED";
    ProjectRole["PROJECT_ROLE_OWNER"] = "PROJECT_ROLE_OWNER";
    ProjectRole["PROJECT_ROLE_MEMBER"] = "PROJECT_ROLE_MEMBER";
})(ProjectRole || (ProjectRole = {}));
export class UsersService {
    static CreateAPIKey(req, initReq) {
        return fm.fetchReq(`/v1/api_keys`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static ListAPIKeys(req, initReq) {
        return fm.fetchReq(`/v1/api_keys?${fm.renderURLSearchParams(req, [])}`, Object.assign(Object.assign({}, initReq), { method: "GET" }));
    }
    static DeleteAPIKey(req, initReq) {
        return fm.fetchReq(`/v1/api_keys/${req["id"]}`, Object.assign(Object.assign({}, initReq), { method: "DELETE" }));
    }
    static UpdateAPIKey(req, initReq) {
        return fm.fetchReq(`/v1/api_keys/${req["apiKeyId"]}`, Object.assign(Object.assign({}, initReq), { method: "PATCH", body: JSON.stringify(req["api_key"]) }));
    }
    static CreateProjectAPIKey(req, initReq) {
        return fm.fetchReq(`/v1/organizations/${req["organizationId"]}/projects/${req["projectId"]}/api_keys`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static ListProjectAPIKeys(req, initReq) {
        return fm.fetchReq(`/v1/organizations/${req["organizationId"]}/projects/${req["projectId"]}/api_keys?${fm.renderURLSearchParams(req, ["organizationId", "projectId"])}`, Object.assign(Object.assign({}, initReq), { method: "GET" }));
    }
    static DeleteProjectAPIKey(req, initReq) {
        return fm.fetchReq(`/v1/organizations/${req["organizationId"]}/projects/${req["projectId"]}/api_keys/${req["id"]}`, Object.assign(Object.assign({}, initReq), { method: "DELETE" }));
    }
    static CreateOrganization(req, initReq) {
        return fm.fetchReq(`/v1/organizations`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static ListOrganizations(req, initReq) {
        return fm.fetchReq(`/v1/organizations?${fm.renderURLSearchParams(req, [])}`, Object.assign(Object.assign({}, initReq), { method: "GET" }));
    }
    static DeleteOrganization(req, initReq) {
        return fm.fetchReq(`/v1/organizations/${req["id"]}`, Object.assign(Object.assign({}, initReq), { method: "DELETE" }));
    }
    static CreateOrganizationUser(req, initReq) {
        return fm.fetchReq(`/v1/organizations/${req["organizationId"]}/users`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static ListOrganizationUsers(req, initReq) {
        return fm.fetchReq(`/v1/organizations/${req["organizationId"]}/users?${fm.renderURLSearchParams(req, ["organizationId"])}`, Object.assign(Object.assign({}, initReq), { method: "GET" }));
    }
    static DeleteOrganizationUser(req, initReq) {
        return fm.fetchReq(`/v1/organizations/${req["organizationId"]}/users/${req["userId"]}`, Object.assign(Object.assign({}, initReq), { method: "DELETE" }));
    }
    static CreateProject(req, initReq) {
        return fm.fetchReq(`/v1/organizations/${req["organizationId"]}/projects`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static ListProjects(req, initReq) {
        return fm.fetchReq(`/v1/organizations/${req["organizationId"]}/projects?${fm.renderURLSearchParams(req, ["organizationId"])}`, Object.assign(Object.assign({}, initReq), { method: "GET" }));
    }
    static DeleteProject(req, initReq) {
        return fm.fetchReq(`/v1/organizations/${req["organizationId"]}/projects/${req["id"]}`, Object.assign(Object.assign({}, initReq), { method: "DELETE" }));
    }
    static CreateProjectUser(req, initReq) {
        return fm.fetchReq(`/v1/organizations/${req["organizationId"]}/projects/${req["projectId"]}/users`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static ListProjectUsers(req, initReq) {
        return fm.fetchReq(`/v1/organizations/${req["organizationId"]}/projects/${req["projectId"]}/users?${fm.renderURLSearchParams(req, ["organizationId", "projectId"])}`, Object.assign(Object.assign({}, initReq), { method: "GET" }));
    }
    static DeleteProjectUser(req, initReq) {
        return fm.fetchReq(`/v1/organizations/${req["organizationId"]}/projects/${req["projectId"]}/users/${req["userId"]}`, Object.assign(Object.assign({}, initReq), { method: "DELETE" }));
    }
    static GetUserSelf(req, initReq) {
        return fm.fetchReq(`/v1/users:getSelf?${fm.renderURLSearchParams(req, [])}`, Object.assign(Object.assign({}, initReq), { method: "GET" }));
    }
}
export class UsersInternalService {
    static ListInternalAPIKeys(req, initReq) {
        return fm.fetchReq(`/llmariner.users.server.v1.UsersInternalService/ListInternalAPIKeys`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static ListInternalOrganizations(req, initReq) {
        return fm.fetchReq(`/llmariner.users.server.v1.UsersInternalService/ListInternalOrganizations`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static ListOrganizationUsers(req, initReq) {
        return fm.fetchReq(`/llmariner.users.server.v1.UsersInternalService/ListOrganizationUsers`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static ListProjects(req, initReq) {
        return fm.fetchReq(`/llmariner.users.server.v1.UsersInternalService/ListProjects`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static ListProjectUsers(req, initReq) {
        return fm.fetchReq(`/llmariner.users.server.v1.UsersInternalService/ListProjectUsers`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static CreateUserInternal(req, initReq) {
        return fm.fetchReq(`/llmariner.users.server.v1.UsersInternalService/CreateUserInternal`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
}
