package server_test

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"
	"testing"

	"github.com/bufbuild/connect-go"
	"github.com/carlmjohnson/be"
	chapv1 "github.com/srerickson/chaparral/gen/chaparral/v1"
	chapv1connect "github.com/srerickson/chaparral/gen/chaparral/v1/chaparralv1connect"
	"github.com/srerickson/chaparral/internal/testutil"
	"github.com/srerickson/chaparral/server"
	"github.com/srerickson/ocfl-go"
	"golang.org/x/exp/slices"
)

const size = 2_000_000

var _ chapv1connect.CommitServiceHandler = (*server.CommitService)(nil)

func TestCommitServiceCommit(t *testing.T) {
	test := func(t *testing.T, htc *http.Client, url string, store *server.StorageRoot) {
		ctx := context.Background()
		chap := chapv1connect.NewCommitServiceClient(htc, url)
		alg := `sha256`
		newUpResp, err := chap.NewUploader(ctx, connect.NewRequest(&chapv1.NewUploaderRequest{
			DigestAlgorithms: []string{alg},
			Description:      "test commit",
		}))
		be.NilErr(t, err)
		uploaderID := newUpResp.Msg.UploaderId
		uploaderURL := url + newUpResp.Msg.UploadPath
		// upload some files and add digests to new state
		newState := map[string]string{}
		filenames := []string{"a.dat", "b.dat", "c.dat"}
		for _, name := range filenames {
			req, err := http.NewRequest(http.MethodPost, uploaderURL, io.LimitReader(rand.Reader, int64(size)))
			be.NilErr(t, err)
			upResp, err := htc.Do(req)
			be.NilErr(t, err)
			respBody, err := io.ReadAll(upResp.Body)
			be.NilErr(t, err)
			be.NilErr(t, upResp.Body.Close())
			be.Equal(t, http.StatusOK, upResp.StatusCode)
			var upload server.HandleUploadResponse
			be.NilErr(t, json.Unmarshal(respBody, &upload))
			newState[name] = upload.Digests[alg]
		}
		commitReq := &chapv1.CommitRequest{
			StorageRootId:   "test",
			DigestAlgorithm: alg,
			State:           newState,
			Message:         "commit v1",
			User:            &chapv1.User{Name: "Test"},
			ObjectId:        "new-01",
			ContentSource: &chapv1.CommitRequest_Uploader{
				Uploader: &chapv1.CommitRequest_UploaderSource{
					UploaderId: uploaderID,
				},
			},
		}
		_, err = chap.Commit(ctx, connect.NewRequest(commitReq))
		be.NilErr(t, err)
		// check object directly
		obj, err := store.GetObjectState(ctx, "new-01", 0)
		be.NilErr(t, err)
		be.Equal(t, len(filenames), obj.State.LenPaths())
		result, err := store.Validate(ctx)
		be.NilErr(t, err)
		be.NilErr(t, result.Err())
	}
	testutil.RunServiceTest(t, test)
}

func TestCommitServiceUploader(t *testing.T) {
	testutil.RunServiceTest(t, func(t *testing.T, htc *http.Client, url string, store *server.StorageRoot) {
		times := 4 // concurrent uploaders
		wg := sync.WaitGroup{}
		wg.Add(times)
		for i := 0; i < times; i++ {
			go func() {
				defer wg.Done()
				testCommitServiceUploader(t, htc, url, store)
			}()
		}
		wg.Wait()
	})
}

// test creating an uploader, uploading to it, accessing it, and destroying it
func testCommitServiceUploader(t *testing.T, htc *http.Client, baseURL string, store *server.StorageRoot) {
	ctx := context.Background()
	chapClient := chapv1connect.NewCommitServiceClient(htc, baseURL)
	// create new uploader
	alg1 := ocfl.SHA256
	alg2 := ocfl.MD5
	desc := "test uploader"
	newUpResp, err := chapClient.NewUploader(ctx, connect.NewRequest(&chapv1.NewUploaderRequest{
		DigestAlgorithms: []string{alg1, alg2},
		Description:      desc,
	}))
	if !nilErr(t, err) {
		return
	}
	uploaderID := newUpResp.Msg.UploaderId
	uploaderPath := newUpResp.Msg.UploadPath
	// concurrent uploads of 2MB random data.
	times := 3
	wg := sync.WaitGroup{}
	wg.Add(times)
	for i := 0; i < times; i++ {
		go func() {
			defer wg.Done()
			digester := ocfl.NewDigester(alg1)
			body := io.TeeReader(io.LimitReader(rand.Reader, int64(size)), digester)
			req, err := http.NewRequest(http.MethodPost, baseURL+uploaderPath, body)
			if !nilErr(t, err) {
				return
			}
			httpResponse, err := htc.Do(req)
			if !nilErr(t, err) {
				return
			}
			defer httpResponse.Body.Close()
			if !isEqual(t, 200, httpResponse.StatusCode) {
				return
			}
			// check result values
			var uploadResult server.HandleUploadResponse
			err = json.NewDecoder(httpResponse.Body).Decode(&uploadResult)
			if !nilErr(t, err) {
				return
			}
			isEqual(t, size, int(uploadResult.Size))
			isEqual(t, digester.String(), uploadResult.Digests[alg1])
			isTrue(t, uploadResult.Digests[alg2] != "")
			isEqual(t, "", uploadResult.Err)

			// get uploader
			getUpResp, err := chapClient.GetUploader(ctx, connect.NewRequest(&chapv1.GetUploaderRequest{
				UploaderId: uploaderID,
			}))
			if !nilErr(t, err) {
				return
			}
			isTrue(t, !getUpResp.Msg.Created.AsTime().IsZero())
			isEqual(t, uploaderID, getUpResp.Msg.UploaderId)
			isEqual(t, uploaderPath, getUpResp.Msg.UploadPath)
			isEqual(t, desc, getUpResp.Msg.Description)
			isTrue(t, slices.Contains(getUpResp.Msg.DigestAlgorithms, alg1))
			isTrue(t, slices.Contains(getUpResp.Msg.DigestAlgorithms, alg2))
			isTrue(t, slices.ContainsFunc(getUpResp.Msg.Uploads, func(up *chapv1.GetUploaderResponse_Upload) bool {
				if digester.String() != up.Digests[alg1] {
					return false
				}
				if up.Digests[alg2] == "" {
					return false
				}
				if size != int(up.Size) {
					return false
				}
				return true
			}))
		}()
	}
	wg.Wait()
	// list the uploaders
	listUpResp, err := chapClient.ListUploaders(ctx, connect.NewRequest(&chapv1.ListUploadersRequest{}))
	if !nilErr(t, err) {
		return
	}
	isTrue(t, slices.ContainsFunc(listUpResp.Msg.Uploaders, func(item *chapv1.ListUploadersResponse_Item) bool {
		return item.UploaderId == uploaderID
	}))
	_, err = chapClient.DeleteUploader(ctx, connect.NewRequest(&chapv1.DeleteUploaderRequest{
		UploaderId: uploaderID,
	}))
	if !nilErr(t, err) {
		return
	}
	// uploader should be gone
	_, err = chapClient.GetUploader(ctx, connect.NewRequest(&chapv1.GetUploaderRequest{
		UploaderId: uploaderID,
	}))
	isConnectErrCode(t, err, connect.CodeNotFound)
	// files should be gone
	// files, err := fsys.ReadDir(ctx, root)
	// if err != nil {
	// 	// err may be nil if fsys is s3/object store
	// 	isTrue(t, errors.Is(err, fs.ErrNotExist))
	// }
	// isTrue(t, len(files) == 0)
}

func nilErr(t *testing.T, err error) bool {
	t.Helper()
	if err != nil {
		t.Errorf("got: %v", err)
		return false
	}
	return true
}

func isEqual[T comparable](t *testing.T, want, got T) bool {
	t.Helper()
	if want != got {
		t.Errorf("want: %v; got: %v", want, got)
		return false
	}
	return true
}

func isTrue(t testing.TB, value bool) bool {
	t.Helper()
	if !value {
		t.Errorf("got: false")
	}
	return value
}

func isConnectErrCode(t *testing.T, err error, want connect.Code) bool {
	var connErr *connect.Error
	if errors.As(err, &connErr) {
		if got := connErr.Code(); got != want {
			t.Errorf("want: %v; got: %v", want, got)
			return false
		}
		return true
	}
	t.Errorf("want: connect error (%v), got: %v", want, err)
	return false
}
