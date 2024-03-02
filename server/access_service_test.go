package server_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"path/filepath"
	"testing"

	"github.com/bufbuild/connect-go"
	"github.com/carlmjohnson/be"
	chap "github.com/srerickson/chaparral"
	chaparralv1 "github.com/srerickson/chaparral/gen/chaparral/v1"
	"github.com/srerickson/chaparral/gen/chaparral/v1/chaparralv1connect"
	"github.com/srerickson/chaparral/internal/testutil"
	"github.com/srerickson/chaparral/server"
	"github.com/srerickson/ocfl-go"
	"github.com/srerickson/ocfl-go/ocflv1"
)

// digest from testdata manifest
const (
	testDigest    = "43a43fe8a8a082d3b5343dfaf2fd0c8b8e370675b1f376e92e9994612c33ea255b11298269d72f797399ebb94edeefe53df243643676548f584fb8603ca53a0f"
	contentLength = 20
)

var _ chaparralv1connect.AccessServiceHandler = (*server.AccessService)(nil)

func TestAccessServiceHandler(t *testing.T) {
	ctx := context.Background()
	testdataDir := filepath.Join("..", "testdata")
	storePath := path.Join("storage-roots", "root-01")
	objectID := "ark:123/abc"
	storeID := "test"
	storeA := testutil.NewStoreTestdata(t, testdataDir)
	mux := server.New(server.WithStorageRoots(storeA))
	srv := httptest.NewTLSServer(mux)
	defer srv.Close()
	httpClient := srv.Client()

	// load fixture for comparison
	storeB, err := ocflv1.GetStore(ctx, ocfl.DirFS(testdataDir), storePath)
	if err != nil {
		t.Fatal("in test setup:", err)
	}
	obj, err := storeB.GetObject(ctx, objectID)
	if err != nil {
		t.Fatal("in test setup:", err)
	}
	expectState := obj.Inventory.Version(0).State
	t.Run("get object version", func(t *testing.T) {
		chap := chaparralv1connect.NewAccessServiceClient(httpClient, srv.URL)
		req := connect.NewRequest(&chaparralv1.GetObjectVersionRequest{
			StorageRootId: storeID,
			ObjectId:      objectID,
		})
		resp, err := chap.GetObjectVersion(ctx, req)
		be.NilErr(t, err)
		got := map[string]string{}
		for d, info := range resp.Msg.State {
			for _, p := range info.Paths {
				got[p] = d
			}
		}
		be.DeepEqual(t, expectState.PathMap(), got)
	})

	t.Run("download by content path", func(t *testing.T) {
		vals := url.Values{
			chap.QueryContentPath: {"inventory.json"},
			chap.QueryObjectID:    {objectID},
			chap.QueryStorageRoot: {storeID},
		}
		u := srv.URL + chap.RouteDownload + "?" + vals.Encode()
		resp, err := httpClient.Get(u)
		be.NilErr(t, err)
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			b, _ := io.ReadAll(resp.Body)
			t.Fatalf("status=%d, resp=%s", resp.StatusCode, string(b))
		}
		// fixture inventory.json size
		be.Equal(t, 744, resp.ContentLength)
	})

	t.Run("download by digest", func(t *testing.T) {
		vals := url.Values{
			chap.QueryDigest:      {testDigest},
			chap.QueryObjectID:    {objectID},
			chap.QueryStorageRoot: {storeID},
		}
		u := srv.URL + chap.RouteDownload + "?" + vals.Encode()
		resp, err := httpClient.Get(u)
		be.NilErr(t, err)
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			b, _ := io.ReadAll(resp.Body)
			t.Fatalf("status=%d, resp=%s", resp.StatusCode, string(b))
		}
		be.Equal(t, contentLength, resp.ContentLength)
	})

	t.Run("head by digest", func(t *testing.T) {
		vals := url.Values{
			chap.QueryDigest:      {testDigest},
			chap.QueryObjectID:    {objectID},
			chap.QueryStorageRoot: {storeID},
		}
		u := srv.URL + chap.RouteDownload + "?" + vals.Encode()
		resp, err := httpClient.Head(u)
		be.NilErr(t, err)
		defer resp.Body.Close()
		be.Equal(t, 200, resp.StatusCode)
		be.Equal(t, contentLength, resp.ContentLength)
	})

	t.Run("download dir", func(t *testing.T) {
		vals := url.Values{
			chap.QueryContentPath: {"v1"},
			chap.QueryObjectID:    {objectID},
			chap.QueryStorageRoot: {storeID},
		}
		u := srv.URL + chap.RouteDownload + "?" + vals.Encode()
		resp, err := httpClient.Get(u)
		be.NilErr(t, err)
		defer resp.Body.Close()
		be.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("download missing content", func(t *testing.T) {
		vals := url.Values{
			chap.QueryContentPath: {"nothing"},
			chap.QueryObjectID:    {objectID},
			chap.QueryStorageRoot: {storeID},
		}
		u := srv.URL + chap.RouteDownload + "?" + vals.Encode()
		resp, err := httpClient.Get(u)
		be.NilErr(t, err)
		defer resp.Body.Close()
		be.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("download missing digest", func(t *testing.T) {
		vals := url.Values{
			chap.QueryDigest:      {"nothing"},
			chap.QueryObjectID:    {objectID},
			chap.QueryStorageRoot: {storeID},
		}
		u := srv.URL + chap.RouteDownload + "?" + vals.Encode()
		resp, err := httpClient.Get(u)
		be.NilErr(t, err)
		defer resp.Body.Close()
		be.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
