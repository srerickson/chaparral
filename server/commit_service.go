package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/bufbuild/connect-go"
	chaparralv1 "github.com/srerickson/chaparral/gen/chaparral/v1"
	"github.com/srerickson/chaparral/gen/chaparral/v1/chaparralv1connect"
	"github.com/srerickson/chaparral/server/internal/lock"
	"github.com/srerickson/chaparral/server/uploader"
	"github.com/srerickson/ocfl-go"
	"github.com/srerickson/ocfl-go/ocflv1"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	CommitServiceName = chaparralv1connect.CommitServiceName
	RouteUpload       = `/` + CommitServiceName + "/" + "upload"
	QueryUploaderID   = "uploader"
	QueryStorageRoot  = "storage_root"
)

var (
	ErrDigestAlgorithm = errors.New("invalid digest algorithm")
)

// CommitService implements chaparral.v1.CommitService
type CommitService struct {
	*chaparral
}

func (s *CommitService) Handler() (string, http.Handler) {
	opts := []connect.HandlerOption{}
	if s.auth != nil {
		opts = append(opts, connect.WithInterceptors(s.AuthorizeInterceptor()))
	}
	route, handle := chaparralv1connect.NewCommitServiceHandler(s, opts...)
	// new handler that includes upload handler
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == RouteUpload && r.Method == http.MethodPost {
			s.HandleUpload(w, r)
			return
		}
		handle.ServeHTTP(w, r)
	})
	return route, fn
}

// Commit is used to create or update OCFL objects
func (s *CommitService) Commit(ctx context.Context, req *connect.Request[chaparralv1.CommitRequest]) (*connect.Response[chaparralv1.CommitResponse], error) {
	commitCtx := context.WithoutCancel(ctx)
	authUser := AuthUserFromCtx(ctx)
	logger := LoggerFromCtx(ctx).With(
		QueryStorageRoot, req.Msg.StorageRootId,
		QueryObjectID, req.Msg.ObjectId,
		"user_id", authUser.ID,
	)
	store, err := s.storageRoot(req.Msg.StorageRootId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if req.Msg.ObjectId == "" {
		err := errors.New("missing required 'object_id' value")
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	commitAlg := req.Msg.DigestAlgorithm
	if commitAlg != ocfl.SHA256 && commitAlg != ocfl.SHA512 {
		err := fmt.Errorf("digest algorithm must be %s or %s", ocfl.SHA512, ocfl.SHA256)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if req.Msg.State == nil {
		err := errors.New("missing required 'state' value")
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if req.Msg.Message == "" {
		err := errors.New("missing required 'message' value")
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if req.Msg.User == nil || req.Msg.User.Name == "" {
		if authUser.Name == "" {
			err := errors.New("missing required 'user name' value")
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		req.Msg.User = &chaparralv1.User{Name: authUser.Name, Address: authUser.Email}
	}
	// prepare commit: handle different content source types
	stage := ocfl.NewStage(commitAlg)
	stage.State, err = PathMap(req.Msg.State).DigestMap()
	if err != nil {
		err := fmt.Errorf("commit request includes invalid object state: %w", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	switch src := req.Msg.ContentSource.(type) {
	case *chaparralv1.CommitRequest_Uploader:
		// commit content from uploader
		uploaderID := src.Uploader.UploaderId
		logger = logger.With(
			"uploader_id", src.Uploader.UploaderId,
		)
		logger.Debug("commit from uploader")
		upper, err := s.uploadMgr.GetUploader(ctx, uploaderID)
		if err != nil {
			err = fmt.Errorf("getting uploader %q: %w", src.Uploader.UploaderId, err)
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		defer func() {
			if err := upper.Close(commitCtx); err != nil {
				logger.Error(err.Error())
			}
		}()
		if !slices.Contains(upper.Config().Algs, commitAlg) {
			err = fmt.Errorf("uploader doesn't provide digest algorithm used by commit: %s", commitAlg)
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		// apply the uploader's contents to the stage
		stage.SetFS(upper.Root())
		for _, up := range upper.Uploads() {
			if err := stage.UnsafeAddPathAs(up.Name, "", up.Digests); err != nil {
				return nil, connect.NewError(connect.CodeInternal, err)
			}
		}
	case *chaparralv1.CommitRequest_Object:
		// commit content from another object
		logger.Debug("commit from existing object state",
			"src_store_id", src.Object.StorageRootId,
			"src_object", src.Object.ObjectId,
		)
		if req.Msg.StorageRootId == src.Object.StorageRootId &&
			req.Msg.ObjectId == src.Object.ObjectId {
			err := errors.New("can't use self as object content source")
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		srcStore, err := s.storageRoot(src.Object.StorageRootId)
		if err != nil {
			err := fmt.Errorf("unknown storage root for source object state: %s", src.Object.StorageRootId)
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		srcObj, err := srcStore.GetObjectState(ctx, src.Object.ObjectId, 0)
		if err != nil {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("in source content: %w", err))
		}
		defer srcObj.Close()
		if srcAlg := srcObj.Alg; srcAlg != commitAlg {
			err = fmt.Errorf("commit declares %s, but source object was created with %s", commitAlg, srcAlg)
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		stage.SetFS(srcObj.FS, srcObj.Path)
		if err = stage.UnsafeSetManifestFixty(srcObj.Manifest, srcObj.Fixity); err != nil {
			err = fmt.Errorf("building new stage manifest from source object: %w", err)
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}
	commitOpts := []ocflv1.CommitOption{
		ocflv1.WithMessage(req.Msg.Message),
		ocflv1.WithUser(*UserFromProto(req.Msg.User)),
		ocflv1.WithLogger(logger.WithGroup("ocfl-go")),
	}
	if req.Msg.Version > 0 {
		commitOpts = append(commitOpts, ocflv1.WithHEAD(int(req.Msg.Version)))
	}
	logger.Debug("finalizing commit")
	if err := store.Commit(commitCtx, req.Msg.ObjectId, stage, commitOpts...); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	resp := &chaparralv1.CommitResponse{}
	return connect.NewResponse(resp), nil
}

// DeleteObject permanently deletes an existing OCFL object.
func (s *CommitService) DeleteObject(ctx context.Context, req *connect.Request[chaparralv1.DeleteObjectRequest]) (*connect.Response[chaparralv1.DeleteObjectResponse], error) {
	store, err := s.storageRoot(req.Msg.StorageRootId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	noCancel := context.WithoutCancel(ctx)
	if err := store.DeleteObject(noCancel, req.Msg.ObjectId); err != nil {
		if errors.Is(err, lock.ErrCapacity) {
			// can't make more object locks
			return nil, connect.NewError(connect.CodeResourceExhausted, err)
		}
		if errors.Is(err, lock.ErrWriteLock) {
			// object is already being deleted or committed
			return nil, connect.NewError(connect.CodeFailedPrecondition, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	resp := &chaparralv1.DeleteObjectResponse{}
	return connect.NewResponse(resp), nil
}

func (s *CommitService) NewUploader(ctx context.Context, req *connect.Request[chaparralv1.NewUploaderRequest]) (*connect.Response[chaparralv1.NewUploaderResponse], error) {
	logger := LoggerFromCtx(ctx)
	user := AuthUserFromCtx(ctx)
	if !slices.ContainsFunc(req.Msg.DigestAlgorithms, func(alg string) bool {
		return alg == ocfl.SHA256 || alg == ocfl.SHA512
	}) {
		err := errors.New("uploader must include sha256 or sha512 digest algorithms")
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	for _, alg := range req.Msg.DigestAlgorithms {
		if ocfl.NewDigester(alg) == nil {
			err := fmt.Errorf("%w: %q", ErrDigestAlgorithm, alg)
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
	}
	uploaderConfig := &uploader.Config{
		Description: req.Msg.Description,
		UserID:      user.ID,
		Algs:        req.Msg.DigestAlgorithms,
	}
	if s.uploadMgr == nil {
		err := errors.New("the storage root does not allow uploading")
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	id, err := s.uploadMgr.NewUploader(ctx, uploaderConfig)
	if err != nil {
		logger.Error(err.Error())
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	newUp, err := s.uploadMgr.GetUploader(ctx, id)
	if err != nil {
		logger.Error(err.Error())
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	defer func() {
		if err := newUp.Close(context.WithoutCancel(ctx)); err != nil {
			logger.Error(err.Error())
		}
	}()
	logger.Debug("new uploader", "uploader_id", id)
	config := newUp.Config()
	resp := &chaparralv1.NewUploaderResponse{
		UploaderId:       id,
		UserId:           config.UserID,
		Description:      config.Description,
		DigestAlgorithms: config.Algs,
		UploadPath:       uploadPath(id),
		Created:          timestamppb.New(newUp.Created()),
	}
	return connect.NewResponse(resp), nil
}

func (s *CommitService) GetUploader(ctx context.Context, req *connect.Request[chaparralv1.GetUploaderRequest]) (*connect.Response[chaparralv1.GetUploaderResponse], error) {
	logger := LoggerFromCtx(ctx)
	upper, err := s.uploadMgr.GetUploader(ctx, req.Msg.UploaderId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	defer func() {
		if err := upper.Close(context.WithoutCancel(ctx)); err != nil {
			logger.Error(err.Error())
		}
	}()
	config := upper.Config()
	resp := &chaparralv1.GetUploaderResponse{
		UploaderId:       req.Msg.UploaderId,
		Created:          timestamppb.New(upper.Created()),
		Description:      config.Description,
		DigestAlgorithms: config.Algs,
		UserId:           config.UserID,
		UploadPath:       uploadPath(req.Msg.UploaderId),
	}
	uploads := upper.Uploads()
	resp.Uploads = make([]*chaparralv1.GetUploaderResponse_Upload, len(uploads))
	for i := range uploads {
		resp.Uploads[i] = &chaparralv1.GetUploaderResponse_Upload{
			Size:    uploads[i].Size,
			Digests: uploads[i].Digests,
		}
	}
	return connect.NewResponse(resp), nil
}

func (s *CommitService) ListUploaders(ctx context.Context, req *connect.Request[chaparralv1.ListUploadersRequest]) (*connect.Response[chaparralv1.ListUploadersResponse], error) {
	logger := LoggerFromCtx(ctx)
	ids, err := s.uploadMgr.UploaderIDs(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	resp := &chaparralv1.ListUploadersResponse{
		Uploaders: make([]*chaparralv1.ListUploadersResponse_Item, len(ids)),
	}
	for i, id := range ids {
		upper, err := s.uploadMgr.GetUploader(ctx, id)
		if err != nil {
			// the uploader may have been deleted between UploaderIDs() and
			// here. If so, skipt it.
			if errors.Is(err, uploader.ErrUploaderDelete) || errors.Is(err, uploader.ErrUploaderNotFound) {
				continue
			}
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		config := upper.Config()
		resp.Uploaders[i] = &chaparralv1.ListUploadersResponse_Item{
			UploaderId:  id,
			Created:     timestamppb.New(upper.Created()),
			Description: config.Description,
			UserId:      config.UserID,
		}
		if err := upper.Close(context.WithoutCancel(ctx)); err != nil {
			logger.ErrorContext(ctx, err.Error())
		}
	}
	return connect.NewResponse(resp), nil

}

// DeleteUploader deletes the uploader created with NewUploader and all files
// uploaded to it. Delete will fail if the uploader is being used, either
// because files are being uploaded to it or because it is being used for a
// commit.
func (s *CommitService) DeleteUploader(ctx context.Context, req *connect.Request[chaparralv1.DeleteUploaderRequest]) (*connect.Response[chaparralv1.DeleteUploaderResponse], error) {
	logger := LoggerFromCtx(ctx).With(QueryUploaderID, req.Msg.UploaderId)
	// don't cancel context if client disconnects.
	noCancelCtx := context.WithoutCancel(ctx)
	upper, err := s.uploadMgr.GetUploader(ctx, req.Msg.UploaderId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	defer func() {
		if err := upper.Close(noCancelCtx); err != nil {
			logger.Error(err.Error())
		}
	}()
	if err := upper.Delete(noCancelCtx); err != nil {
		return nil, connect.NewError(connect.CodeAborted, err)
	}
	resp := &chaparralv1.DeleteUploaderResponse{}
	return connect.NewResponse(resp), nil
}

// HandleUploadResponse is the response body for an upload request.
type HandleUploadResponse struct {
	uploader.Upload
	Err string `json:"error,omitempty"`
}

// Handler for file uploads.
func (s *CommitService) HandleUpload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := AuthUserFromCtx(ctx)
	logger := LoggerFromCtx(ctx)
	var result HandleUploadResponse
	defer func() {
		if err := json.NewEncoder(w).Encode(&result); err != nil {
			logger.Error(err.Error())
		}
	}()
	uploaderID := r.URL.Query().Get(QueryUploaderID)
	if s.auth != nil && !s.auth.ActionAllowed(ctx, &user, CommitAction) {
		w.WriteHeader(http.StatusUnauthorized)
		result.Err = "you don't have permission to upload files"
		return
	}
	upper, err := s.uploadMgr.GetUploader(ctx, uploaderID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		result.Err = fmt.Sprintf("uploader %q: %s", uploaderID, err.Error())
		return
	}
	defer func() {
		noCancel := context.WithoutCancel(ctx)
		if err := upper.Close(noCancel); err != nil {
			logger.Error(err.Error())
		}
	}()
	upload, err := upper.Write(ctx, r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		result.Err = err.Error()
		logger.Error(result.Err)
		return
	}
	result.Digests = upload.Digests
	result.Size = upload.Size
}

// AuthIntercept is middleware that does authorization for all grpc/connect-go
// requests to the commit service. Note that auth for the upload handler is done
// in handler itself.
func (s *CommitService) AuthorizeInterceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		if s.auth == nil {
			return next
		}
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if req.Spec().IsClient {
				// just for server side
				return next(ctx, req)
			}
			user := AuthUserFromCtx(ctx)
			var ok bool
			switch msg := req.Any().(type) {
			case *chaparralv1.CommitRequest:
				ok = s.auth.RootActionAllowed(ctx, &user, CommitAction, msg.StorageRootId)
				if !ok {
					break
				}
				// check permission to read source object (if commit uses one)
				if obj, isObj := msg.ContentSource.(*chaparralv1.CommitRequest_Object); isObj {
					ok = s.auth.RootActionAllowed(ctx, &user, ReadAction, obj.Object.StorageRootId)
				}
			case *chaparralv1.DeleteObjectRequest:
				ok = s.auth.RootActionAllowed(ctx, &user, DeleteAction, msg.StorageRootId)
			case *chaparralv1.NewUploaderRequest:
				ok = s.auth.ActionAllowed(ctx, &user, CommitAction)
			case *chaparralv1.DeleteUploaderRequest:
				ok = s.auth.ActionAllowed(ctx, &user, CommitAction)
			case *chaparralv1.GetUploaderRequest:
				ok = s.auth.ActionAllowed(ctx, &user, CommitAction)
			case *chaparralv1.ListUploadersRequest:
				ok = s.auth.ActionAllowed(ctx, &user, CommitAction)
			}
			if !ok {
				return nil, connect.NewError(connect.CodePermissionDenied, errors.New("API key insufficient permission"))
			}
			return next(ctx, req)
		}
	}
}

func uploadPath(uploadID string) string {
	params := url.Values{QueryUploaderID: {uploadID}}
	return RouteUpload + "?" + params.Encode()
}

// TODO: put this in ocfl-go
type PathMap map[string]string

// DigestMap returns a new DigestMap with contets from pm.
func (pm PathMap) DigestMap() (ocfl.DigestMap, error) {
	dm := map[string][]string{}
	for pth, dig := range pm {
		dm[dig] = append(dm[dig], pth)
	}
	return ocfl.NewDigestMap(dm)
}
