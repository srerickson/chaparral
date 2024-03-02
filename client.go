package chaparral

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/bufbuild/connect-go"
	chapv1 "github.com/srerickson/chaparral/gen/chaparral/v1"
	chapv1connect "github.com/srerickson/chaparral/gen/chaparral/v1/chaparralv1connect"
	"github.com/srerickson/ocfl-go"
)

// http routes for upload/download
const (
	RouteDownload = "/" + chapv1connect.AccessServiceName + "/" + "download"
	RouteUpload   = `/` + chapv1connect.CommitServiceName + "/" + "upload"

	QueryDigest      = "digest"
	QueryContentPath = "content_path"
	QueryObjectID    = "object_id"
	QueryUploaderID  = "uploader"
	QueryStorageRoot = "storage_root"
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

type Commit struct {
	To             ObjectRef
	Version        int
	User           ocfl.User
	Message        string
	State          map[string]string
	Alg            string
	ContentSources []any
}

func commitAsProto(c *Commit) *chapv1.CommitRequest {
	rq := &chapv1.CommitRequest{
		StorageRootId:   c.To.StorageRootID,
		ObjectId:        c.To.ID,
		Message:         c.Message,
		Version:         int32(c.Version),
		User:            &chapv1.User{Name: c.User.Name, Address: c.User.Address},
		State:           c.State,
		DigestAlgorithm: c.Alg,
		ContentSources:  []*chapv1.CommitRequest_ContentSourceItem{},
	}
	for i := range c.ContentSources {
		newSrc := &chapv1.CommitRequest_ContentSourceItem{}
		switch src := c.ContentSources[i].(type) {
		case ObjectRef:
			newSrc.Item = &chapv1.CommitRequest_ContentSourceItem_Object{
				Object: &chapv1.CommitRequest_ObjectSource{
					StorageRootId: src.StorageRootID,
					ObjectId:      src.ID,
				},
			}
		case UploaderRef:
			newSrc.Item = &chapv1.CommitRequest_ContentSourceItem_Uploader{
				Uploader: &chapv1.CommitRequest_UploaderSource{
					UploaderId: src.ID,
				},
			}
		}
		if newSrc.Item != nil {
			rq.ContentSources = append(rq.ContentSources, newSrc)
		}
	}
	return rq
}

type ObjectRef struct {
	StorageRootID string `json:"storage_root_id"`
	ID            string `json:"object_id"`
}

type UploaderRef struct {
	ID string `json:"uploader_id"`
}

// CommitFork creates or updates an object using an existing object as its
// content source. If the commit's state is empty, then the source object's
// current version state is used. If srcStore is empty strings,
// commit.GroupID and commit.StorageRootID are used.
func (cli Client) Commit(ctx context.Context, commit *Commit) error {
	req := commitAsProto(commit)
	_, err := cli.commit.Commit(ctx, connect.NewRequest(req))
	return err
}

// ObjectVersion corresponds to GetObjectVersionResponse proto
type ObjectVersion struct {
	ObjectRef
	Spec            string
	Version         int
	Head            int
	DigestAlgorithm string
	State           Manifest
	Message         string
	User            *ocfl.User
	Created         time.Time
}

func objectVersionFromProto(proto *chapv1.GetObjectVersionResponse) *ObjectVersion {
	state := &ObjectVersion{
		ObjectRef: ObjectRef{
			StorageRootID: proto.StorageRootId,
			ID:            proto.ObjectId,
		},
		Spec:            proto.Spec,
		Version:         int(proto.Version),
		DigestAlgorithm: proto.DigestAlgorithm,
		Head:            int(proto.Head),
		Message:         proto.Message,
		State:           manifestFromProto(proto.State),
	}
	if proto.Created != nil {
		state.Created = proto.Created.AsTime()
	}
	if proto.User != nil {
		state.User = &ocfl.User{Name: proto.User.Name, Address: proto.User.Address}
	}
	return state
}

// Manifest maps digest to FileInfo
type Manifest map[string]FileInfo

func manifestFromProto(proto map[string]*chapv1.FileInfo) Manifest {
	man := Manifest{}
	for digest, info := range proto {
		man[digest] = FileInfo{
			Size:   info.Size,
			Paths:  info.Paths,
			Fixity: info.Fixity,
		}
	}
	return man
}

func (m Manifest) DigestMap() ocfl.DigestMap {
	dm := map[string][]string{}
	for d, info := range m {
		dm[d] = info.Paths
	}
	return ocfl.DigestMap(dm)
}

func (m Manifest) PathMap() ocfl.PathMap {
	pm := map[string]string{}
	for d, info := range m {
		for _, p := range info.Paths {
			pm[p] = d
		}
	}
	return pm
}

type FileInfo struct {
	Size   int64          // number of bytes
	Paths  []string       // sorted slice of path names
	Fixity ocfl.DigestSet // other digests associated witht the content
}

func (cli Client) GetObjectVersion(ctx context.Context, storeID string, objectID string, ver int) (*ObjectVersion, error) {
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

// ObjectManifest corresponds to GetObjectManifestResponse proto
type ObjectManifest struct {
	ObjectRef
	Path            string
	Spec            string
	DigestAlgorithm string
	Manifest        Manifest
}

func objectManifestFromProto(proto *chapv1.GetObjectManifestResponse) *ObjectManifest {
	obj := &ObjectManifest{
		ObjectRef:       ObjectRef{ID: proto.ObjectId, StorageRootID: proto.StorageRootId},
		Path:            proto.Path,
		Spec:            proto.Spec,
		DigestAlgorithm: proto.DigestAlgorithm,
		Manifest:        manifestFromProto(proto.Manifest),
	}
	return obj
}

func (cli Client) GetObjectManifest(ctx context.Context, storeID string, objectID string) (*ObjectManifest, error) {
	req := &chapv1.GetObjectManifestRequest{
		StorageRootId: storeID,
		ObjectId:      objectID,
	}
	resp, err := cli.access.GetObjectManifest(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return objectManifestFromProto(resp.Msg), nil
}

type Content struct {
	io.ReadCloser
	Size int64
}

// Download
func (cli Client) GetContent(ctx context.Context, storeID, objectID, digest string) (*Content, error) {
	u := cli.baseURL + RouteDownload
	vals := url.Values{
		QueryStorageRoot: {storeID},
		QueryObjectID:    {objectID},
		QueryDigest:      {digest},
	}
	resp, err := cli.Client.Get(u + "?" + vals.Encode())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server response (%s): %s", resp.Status, msg)
	}
	return &Content{
		ReadCloser: resp.Body,
		Size:       resp.ContentLength,
	}, nil
}

func (cli Client) GetContentPath(ctx context.Context, storeID, objectID, contentPath string) (*Content, error) {
	u := cli.baseURL + RouteDownload
	vals := url.Values{
		QueryStorageRoot: {storeID},
		QueryObjectID:    {objectID},
		QueryContentPath: {contentPath},
	}
	resp, err := cli.Client.Get(u + "?" + vals.Encode())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server response (%s): %s", resp.Status, msg)
	}
	return &Content{
		ReadCloser: resp.Body,
		Size:       resp.ContentLength,
	}, nil
}

func (cli Client) DeleteObject(ctx context.Context, storeID string, objectID string) error {
	req := &chapv1.DeleteObjectRequest{
		StorageRootId: storeID,
		ObjectId:      objectID,
	}
	_, err := cli.commit.DeleteObject(ctx, connect.NewRequest(req))
	return err
}

type Uploader struct {
	UploaderRef
	UploadPath       string    `json:"upload_path"`
	DigestAlgorithms []string  `json:"digest_algorithms"`
	Description      string    `json:"description"`
	Created          time.Time `json:"created"`
	UserID           string    `json:"user_id"`
	Uploads          []Upload  `json:"uploads,omitempty"`
}

func (cli Client) NewUploader(ctx context.Context, algs []string, desc string) (up *Uploader, err error) {
	req := connect.NewRequest(&chapv1.NewUploaderRequest{
		DigestAlgorithms: algs,
		Description:      desc,
	})
	resp, err := cli.commit.NewUploader(ctx, req)
	if err != nil {
		return
	}
	up = &Uploader{
		UploaderRef:      UploaderRef{ID: resp.Msg.UploaderId},
		UploadPath:       resp.Msg.UploadPath,
		Description:      resp.Msg.Description,
		DigestAlgorithms: resp.Msg.DigestAlgorithms,
		UserID:           resp.Msg.UserId,
	}
	return up, nil
}

func (cli Client) GetUploader(ctx context.Context, id string) (up *Uploader, err error) {
	req := &chapv1.GetUploaderRequest{UploaderId: id}
	resp, err := cli.commit.GetUploader(ctx, connect.NewRequest(req))
	if err != nil {
		return up, err
	}
	up = &Uploader{
		UploaderRef:      UploaderRef{ID: resp.Msg.UploaderId},
		UploadPath:       resp.Msg.UploadPath,
		Description:      resp.Msg.Description,
		DigestAlgorithms: resp.Msg.DigestAlgorithms,
		UserID:           resp.Msg.UserId,
		Uploads:          make([]Upload, len(resp.Msg.Uploads)),
	}
	for i, u := range resp.Msg.Uploads {
		up.Uploads[i].Digests = u.Digests
		up.Uploads[i].Size = u.Size
	}
	return
}

type UploaderListItem struct {
	UploaderRef
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
			UploaderRef: UploaderRef{ID: up.UploaderId},
			Description: up.Description,
			Created:     up.Created.AsTime(),
			UserID:      up.UserId,
		}
	}
	return ids, nil
}

func (cli Client) DeleteUploader(ctx context.Context, id string) error {
	req := &chapv1.DeleteUploaderRequest{UploaderId: id}
	_, err := cli.commit.DeleteUploader(ctx, connect.NewRequest(req))
	return err
}

type Upload struct {
	Size    int64          `json:"size"`
	Digests ocfl.DigestSet `json:"digests"`
}

func (cli Client) Upload(ctx context.Context, uploadPath string, r io.Reader) (result Upload, err error) {
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
	uploadResp := Upload{}
	if jsonErr := json.Unmarshal(byt, &uploadResp); jsonErr != nil {
		return result, errors.Join(err, jsonErr)
	}
	result.Digests = uploadResp.Digests
	result.Size = uploadResp.Size
	return
}

// // UploadStage uploads content files in stage that are used in the stage's state. Content for digests
// // already present in the uploader's Digests list are not uploaded
// func (cli Client) UploadStage(ctx context.Context, up *Uploader, stage *Stage, excludeDigests ...string) error {
// 	if !slices.Contains(up.DigestAlgorithms, stage.Alg) {
// 		return fmt.Errorf("stage and uploader use different digest algorithms")
// 	}
// 	for _, digest := range stage.State {
// 		if slices.Index(excludeDigests, digest) > 0 {
// 			// explicitly ignored
// 			continue
// 		}
// 		if slices.ContainsFunc(up.Uploads, func(existing Upload) bool {
// 			return existing.Digests[stage.Alg] == digest
// 		}) {
// 			// already uploaded
// 			continue
// 		}
// 		staged := stage.Content[digest]
// 		if len(staged) == 0 {
// 			// is that an error?
// 			continue
// 		}
// 		var f fs.File
// 		var err error
// 		f, err = os.Open(staged[0])
// 		if err != nil {
// 			return err
// 		}
// 		defer f.Close()
// 		_, err = cli.Upload(ctx, up.UploadPath, f)
// 		if err != nil {
// 			return fmt.Errorf("uploading %s: %w", staged[0], err)
// 		}
// 	}
// 	return nil
// }

func IsNotFound(err error) bool {
	var connErr *connect.Error
	if errors.As(err, &connErr) && connErr.Code() == connect.CodeNotFound {
		return true
	}
	return false
}
