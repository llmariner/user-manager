/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/
import * as fm from "../../fetch.pb";
export class UsersInternalService {
    static ListInternalAPIKeys(req, initReq) {
        return fm.fetchReq(`/llmoperator.users.server.v1.UsersInternalService/ListInternalAPIKeys`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static ListInternalOrganizations(req, initReq) {
        return fm.fetchReq(`/llmoperator.users.server.v1.UsersInternalService/ListInternalOrganizations`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static ListOrganizationUsers(req, initReq) {
        return fm.fetchReq(`/llmoperator.users.server.v1.UsersInternalService/ListOrganizationUsers`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static ListProjects(req, initReq) {
        return fm.fetchReq(`/llmoperator.users.server.v1.UsersInternalService/ListProjects`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static ListProjectUsers(req, initReq) {
        return fm.fetchReq(`/llmoperator.users.server.v1.UsersInternalService/ListProjectUsers`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
}
