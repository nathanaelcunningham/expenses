// @generated by protoc-gen-connect-query v2.1.0 with parameter "target=ts"
// @generated from file auth/v1/auth.proto (package auth.v1, syntax proto3)
/* eslint-disable */

import { AuthService } from "./auth_pb";

/**
 * @generated from rpc auth.v1.AuthService.Register
 */
export const register = AuthService.method.register;

/**
 * @generated from rpc auth.v1.AuthService.Login
 */
export const login = AuthService.method.login;

/**
 * @generated from rpc auth.v1.AuthService.Logout
 */
export const logout = AuthService.method.logout;

/**
 * @generated from rpc auth.v1.AuthService.RefreshSession
 */
export const refreshSession = AuthService.method.refreshSession;

/**
 * @generated from rpc auth.v1.AuthService.ValidateSession
 */
export const validateSession = AuthService.method.validateSession;
