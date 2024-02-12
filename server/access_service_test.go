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
	chaparralv1 "github.com/srerickson/chaparral/gen/chaparral/v1"
	"github.com/srerickson/chaparral/gen/chaparral/v1/chaparralv1connect"
	"github.com/srerickson/chaparral/internal/testutil"
	"github.com/srerickson/chaparral/server"
	"github.com/srerickson/ocfl-go"
	"github.com/srerickson/ocfl-go/ocflv1"
)

var _ chaparralv1connect.AccessServiceHandler = (*server.AccessService)(nil)

func TestAccessServiceHandler(t *testing.T) {
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
	ctx := context.Background()
	storeB, err := ocflv1.GetStore(ctx, ocfl.DirFS(testdataDir), storePath)
	if err != nil {
		t.Fatal("in test setup:", err)
	}
	obj, err := storeB.GetObject(ctx, objectID)
	if err != nil {
		t.Fatal("in test setup:", err)
	}
	expectState := obj.Inventory.Version(0).State
	t.Run("GetObjectState()", func(t *testing.T) {
		chap := chaparralv1connect.NewAccessServiceClient(httpClient, srv.URL)
		ctx := context.Background()
		req := connect.NewRequest(&chaparralv1.GetObjectStateRequest{
			StorageRootId: storeID,
			ObjectId:      objectID,
		})
		resp, err := chap.GetObjectState(ctx, req)
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
			server.QueryContentPath: {"inventory.json"},
			server.QueryObjectID:    {objectID},
			server.QueryStorageRoot: {storeID},
		}
		u := srv.URL + server.RouteDownload + "?" + vals.Encode()
		resp, err := httpClient.Get(u)
		if err != nil {
			t.Fatal("http client error:", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			b, _ := io.ReadAll(resp.Body)
			t.Log(string(b))
			t.Fatalf("Get(%q): status=%d", u, resp.StatusCode)
		}
	})

	t.Run("download by digest", func(t *testing.T) {
		vals := url.Values{
			server.QueryDigest:      {"43a43fe8a8a082d3b5343dfaf2fd0c8b8e370675b1f376e92e9994612c33ea255b11298269d72f797399ebb94edeefe53df243643676548f584fb8603ca53a0f"},
			server.QueryObjectID:    {objectID},
			server.QueryStorageRoot: {storeID},
		}
		u := srv.URL + server.RouteDownload + "?" + vals.Encode()
		resp, err := httpClient.Get(u)
		if err != nil {
			t.Fatal("http client error:", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			b, _ := io.ReadAll(resp.Body)
			t.Log(string(b))
			t.Fatalf("Get(%q): status=%d", u, resp.StatusCode)
		}
	})

	t.Run("download dir", func(t *testing.T) {
		vals := url.Values{
			server.QueryContentPath: {"v1"},
			server.QueryObjectID:    {objectID},
			server.QueryStorageRoot: {storeID},
		}
		u := srv.URL + server.RouteDownload + "?" + vals.Encode()
		resp, err := httpClient.Get(u)
		be.NilErr(t, err)
		defer resp.Body.Close()
		be.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("download missing content", func(t *testing.T) {
		vals := url.Values{
			server.QueryContentPath: {"nothing"},
			server.QueryObjectID:    {objectID},
			server.QueryStorageRoot: {storeID},
		}
		u := srv.URL + server.RouteDownload + "?" + vals.Encode()
		resp, err := httpClient.Get(u)
		be.NilErr(t, err)
		defer resp.Body.Close()
		be.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("download missing digest", func(t *testing.T) {
		vals := url.Values{
			server.QueryDigest:      {"nothing"},
			server.QueryObjectID:    {objectID},
			server.QueryStorageRoot: {storeID},
		}
		u := srv.URL + server.RouteDownload + "?" + vals.Encode()
		resp, err := httpClient.Get(u)
		be.NilErr(t, err)
		defer resp.Body.Close()
		be.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
