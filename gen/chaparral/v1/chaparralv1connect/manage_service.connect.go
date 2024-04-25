// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: chaparral/v1/manage_service.proto

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
	// ManageServiceName is the fully-qualified name of the ManageService service.
	ManageServiceName = "chaparral.v1.ManageService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// ManageServiceStreamObjectRootsProcedure is the fully-qualified name of the ManageService's
	// StreamObjectRoots RPC.
	ManageServiceStreamObjectRootsProcedure = "/chaparral.v1.ManageService/StreamObjectRoots"
	// ManageServiceSyncObjectProcedure is the fully-qualified name of the ManageService's SyncObject
	// RPC.
	ManageServiceSyncObjectProcedure = "/chaparral.v1.ManageService/SyncObject"
)

// ManageServiceClient is a client for the chaparral.v1.ManageService service.
type ManageServiceClient interface {
	// StreamObjectRoots scans an OCFL storage root and returns a stream
	// of OCFL oobject root details to the caller.
	StreamObjectRoots(context.Context, *connect_go.Request[v1.StreamObjectRootsRequest]) (*connect_go.ServerStreamForClient[v1.StreamObjectRootsResponse], error)
	// SyncObject updates chaparral's internal metadata index to reflect the
	// actual state of an OCFL object. If the object is not found, any
	// references to object in the index is removed.
	SyncObject(context.Context, *connect_go.Request[v1.SyncObjectRequest]) (*connect_go.Response[v1.SyncObjectResponse], error)
}

// NewManageServiceClient constructs a client for the chaparral.v1.ManageService service. By
// default, it uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses,
// and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the
// connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewManageServiceClient(httpClient connect_go.HTTPClient, baseURL string, opts ...connect_go.ClientOption) ManageServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &manageServiceClient{
		streamObjectRoots: connect_go.NewClient[v1.StreamObjectRootsRequest, v1.StreamObjectRootsResponse](
			httpClient,
			baseURL+ManageServiceStreamObjectRootsProcedure,
			opts...,
		),
		syncObject: connect_go.NewClient[v1.SyncObjectRequest, v1.SyncObjectResponse](
			httpClient,
			baseURL+ManageServiceSyncObjectProcedure,
			opts...,
		),
	}
}

// manageServiceClient implements ManageServiceClient.
type manageServiceClient struct {
	streamObjectRoots *connect_go.Client[v1.StreamObjectRootsRequest, v1.StreamObjectRootsResponse]
	syncObject        *connect_go.Client[v1.SyncObjectRequest, v1.SyncObjectResponse]
}

// StreamObjectRoots calls chaparral.v1.ManageService.StreamObjectRoots.
func (c *manageServiceClient) StreamObjectRoots(ctx context.Context, req *connect_go.Request[v1.StreamObjectRootsRequest]) (*connect_go.ServerStreamForClient[v1.StreamObjectRootsResponse], error) {
	return c.streamObjectRoots.CallServerStream(ctx, req)
}

// SyncObject calls chaparral.v1.ManageService.SyncObject.
func (c *manageServiceClient) SyncObject(ctx context.Context, req *connect_go.Request[v1.SyncObjectRequest]) (*connect_go.Response[v1.SyncObjectResponse], error) {
	return c.syncObject.CallUnary(ctx, req)
}

// ManageServiceHandler is an implementation of the chaparral.v1.ManageService service.
type ManageServiceHandler interface {
	// StreamObjectRoots scans an OCFL storage root and returns a stream
	// of OCFL oobject root details to the caller.
	StreamObjectRoots(context.Context, *connect_go.Request[v1.StreamObjectRootsRequest], *connect_go.ServerStream[v1.StreamObjectRootsResponse]) error
	// SyncObject updates chaparral's internal metadata index to reflect the
	// actual state of an OCFL object. If the object is not found, any
	// references to object in the index is removed.
	SyncObject(context.Context, *connect_go.Request[v1.SyncObjectRequest]) (*connect_go.Response[v1.SyncObjectResponse], error)
}

// NewManageServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewManageServiceHandler(svc ManageServiceHandler, opts ...connect_go.HandlerOption) (string, http.Handler) {
	manageServiceStreamObjectRootsHandler := connect_go.NewServerStreamHandler(
		ManageServiceStreamObjectRootsProcedure,
		svc.StreamObjectRoots,
		opts...,
	)
	manageServiceSyncObjectHandler := connect_go.NewUnaryHandler(
		ManageServiceSyncObjectProcedure,
		svc.SyncObject,
		opts...,
	)
	return "/chaparral.v1.ManageService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case ManageServiceStreamObjectRootsProcedure:
			manageServiceStreamObjectRootsHandler.ServeHTTP(w, r)
		case ManageServiceSyncObjectProcedure:
			manageServiceSyncObjectHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedManageServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedManageServiceHandler struct{}

func (UnimplementedManageServiceHandler) StreamObjectRoots(context.Context, *connect_go.Request[v1.StreamObjectRootsRequest], *connect_go.ServerStream[v1.StreamObjectRootsResponse]) error {
	return connect_go.NewError(connect_go.CodeUnimplemented, errors.New("chaparral.v1.ManageService.StreamObjectRoots is not implemented"))
}

func (UnimplementedManageServiceHandler) SyncObject(context.Context, *connect_go.Request[v1.SyncObjectRequest]) (*connect_go.Response[v1.SyncObjectResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("chaparral.v1.ManageService.SyncObject is not implemented"))
}
