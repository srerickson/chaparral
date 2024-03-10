package server_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/carlmjohnson/be"
	chap "github.com/srerickson/chaparral"
	chaparralv1 "github.com/srerickson/chaparral/gen/chaparral/v1"
	"github.com/srerickson/chaparral/gen/chaparral/v1/chaparralv1connect"
	"github.com/srerickson/chaparral/internal/testutil"
	"github.com/srerickson/chaparral/server"
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
	objectID := "ark:123/abc"
	storeID := "test"
	store := testutil.NewStoreTestdata(t, testdataDir)
	mux := server.New(server.WithStorageRoots(store),
		server.WithAuthorizer(testutil.AuthorizeDefaults),
		server.WithAuthUserFunc(testutil.AuthUserFunc()))
	srv := httptest.NewTLSServer(mux)
	defer srv.Close()
	httpClient := srv.Client()

	// load fixture for comparison
	t.Run("get object version", func(t *testing.T) {
		testutil.SetUserToken(httpClient, testutil.ManagerUser)
		chap := chaparralv1connect.NewAccessServiceClient(httpClient, srv.URL)
		req := connect.NewRequest(&chaparralv1.GetObjectVersionRequest{
			StorageRootId: storeID,
			ObjectId:      objectID,
		})
		resp, err := chap.GetObjectVersion(ctx, req)
		be.NilErr(t, err)
		be.Equal(t, storeID, resp.Msg.StorageRootId)
		be.Equal(t, objectID, resp.Msg.ObjectId)
		be.Equal(t, "a_file.txt", resp.Msg.State[testDigest].Paths[0])
		be.Equal(t, "A Person", resp.Msg.User.Name)
		be.Equal(t, "mailto:a_person@example.org", resp.Msg.User.Address)
		be.Equal(t, "An version with one file", resp.Msg.Message)
		be.Equal(t, "sha512", resp.Msg.DigestAlgorithm)
		be.Equal(t, int32(1), resp.Msg.Head)
		be.Equal(t, int32(1), resp.Msg.Version)
		be.Equal(t, testutil.Must(time.Parse(time.RFC3339, "2019-01-01T02:03:04Z")), resp.Msg.Created.AsTime())
		be.Equal(t, "1.0", resp.Msg.Spec)

		t.Run("unauthorized", func(t *testing.T) {
			testutil.SetUserToken(httpClient, testutil.AnonUser)
			_, err := chap.GetObjectVersion(ctx, req)
			be.True(t, err != nil)
			var conErr *connect.Error
			be.True(t, errors.As(err, &conErr))
			be.Equal(t, connect.CodePermissionDenied, conErr.Code())
		})
	})

	t.Run("get object manifest", func(t *testing.T) {
		testutil.SetUserToken(httpClient, testutil.ManagerUser)
		chap := chaparralv1connect.NewAccessServiceClient(httpClient, srv.URL)
		req := connect.NewRequest(&chaparralv1.GetObjectManifestRequest{
			StorageRootId: storeID,
			ObjectId:      objectID,
		})
		resp, err := chap.GetObjectManifest(ctx, req)
		be.NilErr(t, err)
		be.Equal(t, storeID, resp.Msg.StorageRootId)
		be.Equal(t, objectID, resp.Msg.ObjectId)
		be.Equal(t, "v1/content/a_file.txt", resp.Msg.Manifest[testDigest].Paths[0])
		be.Equal(t, "sha512", resp.Msg.DigestAlgorithm)
		be.Equal(t, "1.0", resp.Msg.Spec)

		t.Run("unauthorized", func(t *testing.T) {
			testutil.SetUserToken(httpClient, testutil.AnonUser)
			_, err := chap.GetObjectManifest(ctx, req)
			be.True(t, err != nil)
			var conErr *connect.Error
			be.True(t, errors.As(err, &conErr))
			be.Equal(t, connect.CodePermissionDenied, conErr.Code())
		})
	})

	t.Run("download by content path", func(t *testing.T) {
		testutil.SetUserToken(httpClient, testutil.ManagerUser)
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
		testutil.SetUserToken(httpClient, testutil.ManagerUser)
		vals := url.Values{
			chap.QueryDigest:      {testDigest},
			chap.QueryObjectID:    {objectID},
			chap.QueryStorageRoot: {storeID},
		}
		u := srv.URL + chap.RouteDownload + "?" + vals.Encode()
		resp, err := httpClient.Get(u)
		be.NilErr(t, err)
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Fatalf("status=%d, resp=%s", resp.StatusCode, string(b))
		}
		be.Equal(t, contentLength, resp.ContentLength)

		t.Run("unauthorized", func(t *testing.T) {
			testutil.SetUserToken(httpClient, testutil.AnonUser)
			resp, err := httpClient.Get(u)
			be.NilErr(t, err)
			be.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			resp.Body.Close()
		})
	})

	t.Run("head by digest", func(t *testing.T) {
		testutil.SetUserToken(httpClient, testutil.ManagerUser)
		vals := url.Values{
			chap.QueryDigest:      {testDigest},
			chap.QueryObjectID:    {objectID},
			chap.QueryStorageRoot: {storeID},
		}
		u := srv.URL + chap.RouteDownload + "?" + vals.Encode()
		resp, err := httpClient.Head(u)
		be.NilErr(t, err)
		resp.Body.Close()
		be.Equal(t, 200, resp.StatusCode)
		be.Equal(t, contentLength, resp.ContentLength)
	})

	t.Run("download dir", func(t *testing.T) {
		testutil.SetUserToken(httpClient, testutil.ManagerUser)
		vals := url.Values{
			chap.QueryContentPath: {"v1"},
			chap.QueryObjectID:    {objectID},
			chap.QueryStorageRoot: {storeID},
		}
		u := srv.URL + chap.RouteDownload + "?" + vals.Encode()
		resp, err := httpClient.Get(u)
		be.NilErr(t, err)
		resp.Body.Close()
		be.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("download missing content", func(t *testing.T) {
		testutil.SetUserToken(httpClient, testutil.ManagerUser)
		vals := url.Values{
			chap.QueryContentPath: {"nothing"},
			chap.QueryObjectID:    {objectID},
			chap.QueryStorageRoot: {storeID},
		}
		u := srv.URL + chap.RouteDownload + "?" + vals.Encode()
		resp, err := httpClient.Get(u)
		be.NilErr(t, err)
		resp.Body.Close()
		be.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("download missing digest", func(t *testing.T) {
		testutil.SetUserToken(httpClient, testutil.ManagerUser)
		vals := url.Values{
			chap.QueryDigest:      {"nothing"},
			chap.QueryObjectID:    {objectID},
			chap.QueryStorageRoot: {storeID},
		}
		u := srv.URL + chap.RouteDownload + "?" + vals.Encode()
		resp, err := httpClient.Get(u)
		be.NilErr(t, err)
		resp.Body.Close()
		be.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
