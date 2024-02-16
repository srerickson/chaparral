package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"slices"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/srerickson/chaparral"
	chapv1 "github.com/srerickson/chaparral/gen/chaparral/v1"
	chapv1connect "github.com/srerickson/chaparral/gen/chaparral/v1/chaparralv1connect"
	"github.com/srerickson/ocfl-go"
)

type Client struct {
	*http.Client
	baseURL string
	access  chapv1connect.AccessServiceClient
	commit  chapv1connect.CommitServiceClient
}

func NewClient(c *http.Client, baseurl string) *Client {
	return &Client{
		Client:  c,
		baseURL: baseurl,
		access:  chapv1connect.NewAccessServiceClient(c, baseurl),
		commit:  chapv1connect.NewCommitServiceClient(c, baseurl),
	}
}

func (cli Client) NewUploader(ctx context.Context, algs []string, desc string) (up *chaparral.Uploader, err error) {
	req := connect.NewRequest(&chapv1.NewUploaderRequest{
		DigestAlgorithms: algs,
		Description:      desc,
	})
	resp, err := cli.commit.NewUploader(ctx, req)
	if err != nil {
		return
	}
	up = &chaparral.Uploader{
		ID:               resp.Msg.UploaderId,
		UploadPath:       resp.Msg.UploadPath,
		Description:      resp.Msg.Description,
		DigestAlgorithms: resp.Msg.DigestAlgorithms,
		UserID:           resp.Msg.UserId,
	}
	return up, nil
}

func (cli Client) GetUploader(ctx context.Context, id string) (up *chaparral.Uploader, err error) {
	req := &chapv1.GetUploaderRequest{UploaderId: id}
	resp, err := cli.commit.GetUploader(ctx, connect.NewRequest(req))
	if err != nil {
		return up, err
	}
	up = &chaparral.Uploader{
		ID:               resp.Msg.UploaderId,
		UploadPath:       resp.Msg.UploadPath,
		Description:      resp.Msg.Description,
		DigestAlgorithms: resp.Msg.DigestAlgorithms,
		UserID:           resp.Msg.UserId,
		Uploads:          make([]chaparral.Upload, len(resp.Msg.Uploads)),
	}
	for i, u := range resp.Msg.Uploads {
		up.Uploads[i].Digests = u.Digests
		up.Uploads[i].Size = u.Size
	}
	return
}

type UploaderListItem struct {
	ID          string
	Created     time.Time
	UserID      string
	Description string
}

func (cli Client) ListUploaders(ctx context.Context) ([]UploaderListItem, error) {
	req := &chapv1.ListUploadersRequest{}
	resp, err := cli.commit.ListUploaders(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	ids := make([]UploaderListItem, len(resp.Msg.Uploaders))
	for i, up := range resp.Msg.Uploaders {
		ids[i] = UploaderListItem{
			ID:          up.UploaderId,
			Description: up.Description,
			Created:     up.Created.AsTime(),
			UserID:      up.UserId,
		}
	}
	return ids, nil
}

func (cli Client) DeleteUploader(ctx context.Context, id string) error {
	_, err := cli.commit.DeleteUploader(ctx, connect.NewRequest(&chapv1.DeleteUploaderRequest{UploaderId: id}))
	return err
}

func (cli Client) Upload(ctx context.Context, uploadPath string, r io.Reader) (result chaparral.Upload, err error) {
	resp, err := cli.Post(cli.baseURL+uploadPath, "application/octet-stream", r)
	if err != nil {
		return result, fmt.Errorf("during upload: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}()
	byt, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, fmt.Errorf("reading upload response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("unexpected upload response status: %q", resp.Status)
	}
	uploadResp := chaparral.UploadResult{}
	if jsonErr := json.Unmarshal(byt, &uploadResp); jsonErr != nil {
		return result, errors.Join(err, jsonErr)
	}
	if uploadResp.Err != "" {
		err = errors.Join(err, errors.New(uploadResp.Err))
	}
	result.Digests = uploadResp.Digests
	result.Size = uploadResp.Size
	return
}

type Commit struct {
	StorageRootID string
	ObjectID      string
	Version       int
	User          ocfl.User
	Message       string
	State         map[string]string
	Alg           string
}

// valid checks required commit fields
func (commit Commit) valid() error {
	if commit.ObjectID == "" {
		return errors.New("missing required field: 'object id'")
	}
	if commit.Message == "" {
		return errors.New("missing required field: 'message'")
	}
	return nil
}

func (commit *Commit) asProto() *chapv1.CommitRequest {
	return &chapv1.CommitRequest{
		StorageRootId:   commit.StorageRootID,
		ObjectId:        commit.ObjectID,
		Message:         commit.Message,
		Version:         int32(commit.Version),
		User:            &chapv1.User{Name: commit.User.Name, Address: commit.User.Address},
		State:           commit.State,
		DigestAlgorithm: commit.Alg,
	}
}

// CommitFork creates or updates an object using an existing object as its
// content source. If the commit's state is empty, then the source object's
// current version state is used. If srcStore is empty strings,
// commit.GroupID and commit.StorageRootID are used.
func (cli Client) CommitFork(ctx context.Context, commit *Commit, srcStore, srcObj string) error {
	if err := commit.valid(); err != nil {
		return err
	}
	objSource := &chapv1.CommitRequest_ObjectSource{
		StorageRootId: srcStore,
		ObjectId:      srcObj,
	}
	req := commit.asProto()
	req.ContentSource = &chapv1.CommitRequest_Object{Object: objSource}
	if _, err := cli.commit.Commit(ctx, connect.NewRequest(req)); err != nil {
		return err
	}
	return nil
}

func (cli Client) CommitUploader(ctx context.Context, commit *Commit, up *chaparral.Uploader) error {
	if err := commit.valid(); err != nil {
		return err
	}
	req := commit.asProto()
	if up != nil {
		req.ContentSource = &chapv1.CommitRequest_Uploader{
			Uploader: &chapv1.CommitRequest_UploaderSource{
				UploaderId: up.ID,
			},
		}
	}
	_, err := cli.commit.Commit(ctx, connect.NewRequest(req))
	if err != nil {
		// TODO: return dirty state and uploader id somehow
		return err
	}
	return nil
}

func objectVersionFromProto(proto *chapv1.GetObjectVersionResponse) *chaparral.ObjectVersion {
	state := &chaparral.ObjectVersion{
		StorageRootID:   proto.StorageRootId,
		ObjectID:        proto.ObjectId,
		Spec:            proto.Spec,
		Version:         int(proto.Version),
		DigestAlgorithm: proto.DigestAlgorithm,
		Head:            int(proto.Head),
		Message:         proto.Message,
		State:           chaparral.Manifest{},
	}
	for digest, info := range proto.State {
		state.State[digest] = chaparral.FileInfo{
			Size:   info.Size,
			Paths:  info.Paths,
			Fixity: info.Fixity,
		}
	}
	if proto.Created != nil {
		state.Created = proto.Created.AsTime()
	}
	if proto.User != nil {
		state.User = &ocfl.User{Name: proto.User.Name, Address: proto.User.Address}
	}
	return state
}

func (cli Client) GetObjectVersion(ctx context.Context, storeID string, objectID string, ver int) (*chaparral.ObjectVersion, error) {
	req := &chapv1.GetObjectVersionRequest{
		StorageRootId: storeID,
		ObjectId:      objectID,
		Version:       int32(ver),
	}
	resp, err := cli.access.GetObjectVersion(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	state := objectVersionFromProto(resp.Msg)
	return state, nil
}

func (cli Client) DeleteObject(ctx context.Context, storeID string, objectID string) error {
	req := &chapv1.DeleteObjectRequest{
		StorageRootId: storeID,
		ObjectId:      objectID,
	}
	_, err := cli.commit.DeleteObject(ctx, connect.NewRequest(req))
	return err
}

// Download
func (cli Client) GetContent(ctx context.Context, storeID, objectID, digest, contentPath string) (io.ReadCloser, error) {
	u := cli.baseURL + chaparral.RouteDownload
	vals := url.Values{
		chaparral.QueryStorageRoot: {storeID},
		chaparral.QueryObjectID:    {objectID},
		chaparral.QueryContentPath: {contentPath},
		chaparral.QueryDigest:      {digest},
	}
	resp, err := cli.Client.Get(u + "?" + vals.Encode())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server response (%s): %s", resp.Status, msg)
	}
	return resp.Body, nil
}

// UploadStage uploads content files in stage that are used in the stage's state. Content for digests
// already present in the uploader's Digests list are not uploaded
func (cli Client) UploadStage(ctx context.Context, up *chaparral.Uploader, stage *Stage, excludeDigests ...string) error {
	if !slices.Contains(up.DigestAlgorithms, stage.Alg) {
		return fmt.Errorf("stage and uploader use different digest algorithms")
	}
	for _, digest := range stage.State {
		if slices.Index(excludeDigests, digest) > 0 {
			// explicitly ignored
			continue
		}
		if slices.ContainsFunc(up.Uploads, func(existing chaparral.Upload) bool {
			return existing.Digests[stage.Alg] == digest
		}) {
			// already uploaded
			continue
		}
		staged := stage.Content[digest]
		if len(staged) == 0 {
			// is that an error?
			continue
		}
		var f fs.File
		var err error
		f, err = os.Open(staged[0])
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = cli.Upload(ctx, up.UploadPath, f)
		if err != nil {
			return fmt.Errorf("uploading %s: %w", staged[0], err)
		}
	}
	return nil
}

func IsNotFound(err error) bool {
	var connErr *connect.Error
	if errors.As(err, &connErr) && connErr.Code() == connect.CodeNotFound {
		return true
	}
	return false
}
