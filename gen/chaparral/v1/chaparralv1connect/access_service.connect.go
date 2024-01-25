// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: chaparral/v1/access_service.proto

package chaparralv1connect

import (
	context "context"
	errors "errors"
	connect_go "github.com/bufbuild/connect-go"
	v1 "github.com/srerickson/chaparral/gen/chaparral/v1"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect_go.IsAtLeastVersion0_1_0

const (
	// AccessServiceName is the fully-qualified name of the AccessService service.
	AccessServiceName = "chaparral.v1.AccessService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// AccessServiceGetObjectStateProcedure is the fully-qualified name of the AccessService's
	// GetObjectState RPC.
	AccessServiceGetObjectStateProcedure = "/chaparral.v1.AccessService/GetObjectState"
	// AccessServiceGetObjectManifestProcedure is the fully-qualified name of the AccessService's
	// GetObjectManifest RPC.
	AccessServiceGetObjectManifestProcedure = "/chaparral.v1.AccessService/GetObjectManifest"
)

// AccessServiceClient is a client for the chaparral.v1.AccessService service.
type AccessServiceClient interface {
	// GetObjectState returns details about the logical state of an OCFL object
	// version.
	GetObjectState(context.Context, *connect_go.Request[v1.GetObjectStateRequest]) (*connect_go.Response[v1.GetObjectStateResponse], error)
	// GetObjectManifest returns digests, sizes, and fixity information for all
	// content associated with an object across all its versions.
	GetObjectManifest(context.Context, *connect_go.Request[v1.GetObjectManifestRequest]) (*connect_go.Response[v1.GetObjectManifestResponse], error)
}

// NewAccessServiceClient constructs a client for the chaparral.v1.AccessService service. By
// default, it uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses,
// and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the
// connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewAccessServiceClient(httpClient connect_go.HTTPClient, baseURL string, opts ...connect_go.ClientOption) AccessServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &accessServiceClient{
		getObjectState: connect_go.NewClient[v1.GetObjectStateRequest, v1.GetObjectStateResponse](
			httpClient,
			baseURL+AccessServiceGetObjectStateProcedure,
			opts...,
		),
		getObjectManifest: connect_go.NewClient[v1.GetObjectManifestRequest, v1.GetObjectManifestResponse](
			httpClient,
			baseURL+AccessServiceGetObjectManifestProcedure,
			opts...,
		),
	}
}

// accessServiceClient implements AccessServiceClient.
type accessServiceClient struct {
	getObjectState    *connect_go.Client[v1.GetObjectStateRequest, v1.GetObjectStateResponse]
	getObjectManifest *connect_go.Client[v1.GetObjectManifestRequest, v1.GetObjectManifestResponse]
}

// GetObjectState calls chaparral.v1.AccessService.GetObjectState.
func (c *accessServiceClient) GetObjectState(ctx context.Context, req *connect_go.Request[v1.GetObjectStateRequest]) (*connect_go.Response[v1.GetObjectStateResponse], error) {
	return c.getObjectState.CallUnary(ctx, req)
}

// GetObjectManifest calls chaparral.v1.AccessService.GetObjectManifest.
func (c *accessServiceClient) GetObjectManifest(ctx context.Context, req *connect_go.Request[v1.GetObjectManifestRequest]) (*connect_go.Response[v1.GetObjectManifestResponse], error) {
	return c.getObjectManifest.CallUnary(ctx, req)
}

// AccessServiceHandler is an implementation of the chaparral.v1.AccessService service.
type AccessServiceHandler interface {
	// GetObjectState returns details about the logical state of an OCFL object
	// version.
	GetObjectState(context.Context, *connect_go.Request[v1.GetObjectStateRequest]) (*connect_go.Response[v1.GetObjectStateResponse], error)
	// GetObjectManifest returns digests, sizes, and fixity information for all
	// content associated with an object across all its versions.
	GetObjectManifest(context.Context, *connect_go.Request[v1.GetObjectManifestRequest]) (*connect_go.Response[v1.GetObjectManifestResponse], error)
}

// NewAccessServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewAccessServiceHandler(svc AccessServiceHandler, opts ...connect_go.HandlerOption) (string, http.Handler) {
	accessServiceGetObjectStateHandler := connect_go.NewUnaryHandler(
		AccessServiceGetObjectStateProcedure,
		svc.GetObjectState,
		opts...,
	)
	accessServiceGetObjectManifestHandler := connect_go.NewUnaryHandler(
		AccessServiceGetObjectManifestProcedure,
		svc.GetObjectManifest,
		opts...,
	)
	return "/chaparral.v1.AccessService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case AccessServiceGetObjectStateProcedure:
			accessServiceGetObjectStateHandler.ServeHTTP(w, r)
		case AccessServiceGetObjectManifestProcedure:
			accessServiceGetObjectManifestHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedAccessServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedAccessServiceHandler struct{}

func (UnimplementedAccessServiceHandler) GetObjectState(context.Context, *connect_go.Request[v1.GetObjectStateRequest]) (*connect_go.Response[v1.GetObjectStateResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("chaparral.v1.AccessService.GetObjectState is not implemented"))
}

func (UnimplementedAccessServiceHandler) GetObjectManifest(context.Context, *connect_go.Request[v1.GetObjectManifestRequest]) (*connect_go.Response[v1.GetObjectManifestResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("chaparral.v1.AccessService.GetObjectManifest is not implemented"))
}
