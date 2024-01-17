package testutil

import (
	"crypto/rand"
	"crypto/rsa"
	_ "embed"
	"fmt"
	"net/http"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/srerickson/chaparral/server"
)

var (
	// key used for JWS signing/validation in tests
	key *rsa.PrivateKey

	// canned users for testing
	MemberUser = server.AuthUser{
		ID:    "test-member",
		Email: "test-member@testing.com",
		Name:  "Test Member",
		Roles: []string{server.MemberRole}}
	ManagerUser = server.AuthUser{
		ID:    "test-manager",
		Email: "test-manager@testing.com",
		Name:  "Test Manager",
		Roles: []string{server.ManagerRole}}
	AdminUser = server.AuthUser{
		ID:    "test-admin",
		Email: "test-admin@testing.com",
		Name:  "Test Admin",
		Roles: []string{server.AdminRole}}

	// canned permissions used in testing
	AllowAll = server.Permissions{
		// anyone can do anythong
		server.DefaultRole: []server.RolePermission{
			{Actions: []string{"*"}, StorageRootID: "*"},
		},
	}
	AllowNone = server.Permissions(nil)
)

func testKey() *rsa.PrivateKey {
	if key != nil {
		return key
	}
	var err error
	key, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	return key
}

// AuthorizeClient modifies the client to include a bearer token
// for the given user. The token is signed with testKey.
func authorizeClient(cli *http.Client, user server.AuthUser) {
	if cli.Transport == nil {
		cli.Transport = http.DefaultTransport
	}
	cli.Transport = &bearerTokenTransport{
		Token: AuthUserToken(user),
		Base:  cli.Transport,
	}
}

// AuthUserToken generates a token for the given user signed with the test kßey.
func AuthUserToken(user server.AuthUser) string {
	key := testKey()
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: key}, (&jose.SignerOptions{}).WithType("JWT"))
	if err != nil {
		panic(fmt.Errorf("user token signing: %v", err))
	}
	token := server.AuthToken{
		User: user,
		Claims: jwt.Claims{
			Issuer:    "chaparral-test",
			Subject:   user.ID,
			Audience:  jwt.Audience{"chaparral-test"},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			Expiry:    jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}
	encToken, err := jwt.Signed(signer).Claims(token).CompactSerialize()
	if err != nil {
		panic(fmt.Errorf("user token signing: %v", err))
	}
	return encToken
}

// Permissions used for testing

type bearerTokenTransport struct {
	Token string
	Base  http.RoundTripper
}

func (t *bearerTokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.Token)
	return t.Base.RoundTrip(req)
}
