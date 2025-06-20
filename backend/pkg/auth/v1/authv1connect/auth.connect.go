// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: auth/v1/auth.proto

package authv1connect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	v1 "expenses-backend/pkg/auth/v1"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect.IsAtLeastVersion1_13_0

const (
	// AuthServiceName is the fully-qualified name of the AuthService service.
	AuthServiceName = "auth.v1.AuthService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// AuthServiceRegisterProcedure is the fully-qualified name of the AuthService's Register RPC.
	AuthServiceRegisterProcedure = "/auth.v1.AuthService/Register"
	// AuthServiceLoginProcedure is the fully-qualified name of the AuthService's Login RPC.
	AuthServiceLoginProcedure = "/auth.v1.AuthService/Login"
	// AuthServiceLogoutProcedure is the fully-qualified name of the AuthService's Logout RPC.
	AuthServiceLogoutProcedure = "/auth.v1.AuthService/Logout"
	// AuthServiceRefreshSessionProcedure is the fully-qualified name of the AuthService's
	// RefreshSession RPC.
	AuthServiceRefreshSessionProcedure = "/auth.v1.AuthService/RefreshSession"
	// AuthServiceValidateSessionProcedure is the fully-qualified name of the AuthService's
	// ValidateSession RPC.
	AuthServiceValidateSessionProcedure = "/auth.v1.AuthService/ValidateSession"
)

// AuthServiceClient is a client for the auth.v1.AuthService service.
type AuthServiceClient interface {
	Register(context.Context, *connect.Request[v1.RegisterRequest]) (*connect.Response[v1.RegisterResponse], error)
	Login(context.Context, *connect.Request[v1.LoginRequest]) (*connect.Response[v1.LoginResponse], error)
	Logout(context.Context, *connect.Request[v1.LogoutRequest]) (*connect.Response[v1.LogoutResponse], error)
	RefreshSession(context.Context, *connect.Request[v1.RefreshSessionRequest]) (*connect.Response[v1.RefreshSessionResponse], error)
	ValidateSession(context.Context, *connect.Request[v1.ValidateSessionRequest]) (*connect.Response[v1.ValidateSessionResponse], error)
}

// NewAuthServiceClient constructs a client for the auth.v1.AuthService service. By default, it uses
// the Connect protocol with the binary Protobuf Codec, asks for gzipped responses, and sends
// uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the connect.WithGRPC() or
// connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewAuthServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) AuthServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	authServiceMethods := v1.File_auth_v1_auth_proto.Services().ByName("AuthService").Methods()
	return &authServiceClient{
		register: connect.NewClient[v1.RegisterRequest, v1.RegisterResponse](
			httpClient,
			baseURL+AuthServiceRegisterProcedure,
			connect.WithSchema(authServiceMethods.ByName("Register")),
			connect.WithClientOptions(opts...),
		),
		login: connect.NewClient[v1.LoginRequest, v1.LoginResponse](
			httpClient,
			baseURL+AuthServiceLoginProcedure,
			connect.WithSchema(authServiceMethods.ByName("Login")),
			connect.WithClientOptions(opts...),
		),
		logout: connect.NewClient[v1.LogoutRequest, v1.LogoutResponse](
			httpClient,
			baseURL+AuthServiceLogoutProcedure,
			connect.WithSchema(authServiceMethods.ByName("Logout")),
			connect.WithClientOptions(opts...),
		),
		refreshSession: connect.NewClient[v1.RefreshSessionRequest, v1.RefreshSessionResponse](
			httpClient,
			baseURL+AuthServiceRefreshSessionProcedure,
			connect.WithSchema(authServiceMethods.ByName("RefreshSession")),
			connect.WithClientOptions(opts...),
		),
		validateSession: connect.NewClient[v1.ValidateSessionRequest, v1.ValidateSessionResponse](
			httpClient,
			baseURL+AuthServiceValidateSessionProcedure,
			connect.WithSchema(authServiceMethods.ByName("ValidateSession")),
			connect.WithClientOptions(opts...),
		),
	}
}

// authServiceClient implements AuthServiceClient.
type authServiceClient struct {
	register        *connect.Client[v1.RegisterRequest, v1.RegisterResponse]
	login           *connect.Client[v1.LoginRequest, v1.LoginResponse]
	logout          *connect.Client[v1.LogoutRequest, v1.LogoutResponse]
	refreshSession  *connect.Client[v1.RefreshSessionRequest, v1.RefreshSessionResponse]
	validateSession *connect.Client[v1.ValidateSessionRequest, v1.ValidateSessionResponse]
}

// Register calls auth.v1.AuthService.Register.
func (c *authServiceClient) Register(ctx context.Context, req *connect.Request[v1.RegisterRequest]) (*connect.Response[v1.RegisterResponse], error) {
	return c.register.CallUnary(ctx, req)
}

// Login calls auth.v1.AuthService.Login.
func (c *authServiceClient) Login(ctx context.Context, req *connect.Request[v1.LoginRequest]) (*connect.Response[v1.LoginResponse], error) {
	return c.login.CallUnary(ctx, req)
}

// Logout calls auth.v1.AuthService.Logout.
func (c *authServiceClient) Logout(ctx context.Context, req *connect.Request[v1.LogoutRequest]) (*connect.Response[v1.LogoutResponse], error) {
	return c.logout.CallUnary(ctx, req)
}

// RefreshSession calls auth.v1.AuthService.RefreshSession.
func (c *authServiceClient) RefreshSession(ctx context.Context, req *connect.Request[v1.RefreshSessionRequest]) (*connect.Response[v1.RefreshSessionResponse], error) {
	return c.refreshSession.CallUnary(ctx, req)
}

// ValidateSession calls auth.v1.AuthService.ValidateSession.
func (c *authServiceClient) ValidateSession(ctx context.Context, req *connect.Request[v1.ValidateSessionRequest]) (*connect.Response[v1.ValidateSessionResponse], error) {
	return c.validateSession.CallUnary(ctx, req)
}

// AuthServiceHandler is an implementation of the auth.v1.AuthService service.
type AuthServiceHandler interface {
	Register(context.Context, *connect.Request[v1.RegisterRequest]) (*connect.Response[v1.RegisterResponse], error)
	Login(context.Context, *connect.Request[v1.LoginRequest]) (*connect.Response[v1.LoginResponse], error)
	Logout(context.Context, *connect.Request[v1.LogoutRequest]) (*connect.Response[v1.LogoutResponse], error)
	RefreshSession(context.Context, *connect.Request[v1.RefreshSessionRequest]) (*connect.Response[v1.RefreshSessionResponse], error)
	ValidateSession(context.Context, *connect.Request[v1.ValidateSessionRequest]) (*connect.Response[v1.ValidateSessionResponse], error)
}

// NewAuthServiceHandler builds an HTTP handler from the service implementation. It returns the path
// on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewAuthServiceHandler(svc AuthServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	authServiceMethods := v1.File_auth_v1_auth_proto.Services().ByName("AuthService").Methods()
	authServiceRegisterHandler := connect.NewUnaryHandler(
		AuthServiceRegisterProcedure,
		svc.Register,
		connect.WithSchema(authServiceMethods.ByName("Register")),
		connect.WithHandlerOptions(opts...),
	)
	authServiceLoginHandler := connect.NewUnaryHandler(
		AuthServiceLoginProcedure,
		svc.Login,
		connect.WithSchema(authServiceMethods.ByName("Login")),
		connect.WithHandlerOptions(opts...),
	)
	authServiceLogoutHandler := connect.NewUnaryHandler(
		AuthServiceLogoutProcedure,
		svc.Logout,
		connect.WithSchema(authServiceMethods.ByName("Logout")),
		connect.WithHandlerOptions(opts...),
	)
	authServiceRefreshSessionHandler := connect.NewUnaryHandler(
		AuthServiceRefreshSessionProcedure,
		svc.RefreshSession,
		connect.WithSchema(authServiceMethods.ByName("RefreshSession")),
		connect.WithHandlerOptions(opts...),
	)
	authServiceValidateSessionHandler := connect.NewUnaryHandler(
		AuthServiceValidateSessionProcedure,
		svc.ValidateSession,
		connect.WithSchema(authServiceMethods.ByName("ValidateSession")),
		connect.WithHandlerOptions(opts...),
	)
	return "/auth.v1.AuthService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case AuthServiceRegisterProcedure:
			authServiceRegisterHandler.ServeHTTP(w, r)
		case AuthServiceLoginProcedure:
			authServiceLoginHandler.ServeHTTP(w, r)
		case AuthServiceLogoutProcedure:
			authServiceLogoutHandler.ServeHTTP(w, r)
		case AuthServiceRefreshSessionProcedure:
			authServiceRefreshSessionHandler.ServeHTTP(w, r)
		case AuthServiceValidateSessionProcedure:
			authServiceValidateSessionHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedAuthServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedAuthServiceHandler struct{}

func (UnimplementedAuthServiceHandler) Register(context.Context, *connect.Request[v1.RegisterRequest]) (*connect.Response[v1.RegisterResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("auth.v1.AuthService.Register is not implemented"))
}

func (UnimplementedAuthServiceHandler) Login(context.Context, *connect.Request[v1.LoginRequest]) (*connect.Response[v1.LoginResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("auth.v1.AuthService.Login is not implemented"))
}

func (UnimplementedAuthServiceHandler) Logout(context.Context, *connect.Request[v1.LogoutRequest]) (*connect.Response[v1.LogoutResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("auth.v1.AuthService.Logout is not implemented"))
}

func (UnimplementedAuthServiceHandler) RefreshSession(context.Context, *connect.Request[v1.RefreshSessionRequest]) (*connect.Response[v1.RefreshSessionResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("auth.v1.AuthService.RefreshSession is not implemented"))
}

func (UnimplementedAuthServiceHandler) ValidateSession(context.Context, *connect.Request[v1.ValidateSessionRequest]) (*connect.Response[v1.ValidateSessionResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("auth.v1.AuthService.ValidateSession is not implemented"))
}
