package testutil

import (
	"crypto/rand"
	"crypto/rsa"
	_ "embed"
	"fmt"
	"net/http"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/srerickson/chaparral/server"
)

const (
	issuer = "chaparral-test"

	// roles used for testing
	RoleMember  = issuer + ":member"
	RoleManager = issuer + ":manager"
	RoleAdmin   = issuer + ":admin"
)

var (
	// key used for JWS signing/validation in tests
	key *rsa.PrivateKey

	AnonUser = server.AuthUser{}
	// canned users for testing
	MemberUser = server.AuthUser{
		ID:    "test-member",
		Email: "test-member@testing.com",
		Name:  "Test Member",
		Roles: []string{RoleMember}}
	ManagerUser = server.AuthUser{
		ID:    "test-manager",
		Email: "test-manager@testing.com",
		Name:  "Test Manager",
		Roles: []string{RoleManager}}
	AdminUser = server.AuthUser{
		ID:    "test-admin",
		Email: "test-admin@testing.com",
		Name:  "Test Admin",
		Roles: []string{RoleAdmin}}

	// canned permissions used in testing
	AuthorizeAll  = server.RolePermissions{Default: server.Permissions{"*": []string{"*"}}}
	AuthorizeNone = server.RolePermissions{}

	// default
	AuthorizeDefaults = DefaultRoles("test")
)

// DefaultRoles is a set of role permissions used in testing
func DefaultRoles(defaultRoot string) server.RolePermissions {
	return server.RolePermissions{
		// No access for un-authenticated users
		Default: server.Permissions{},
		Roles: map[string]server.Permissions{
			// members can read objects in the default storage root
			RoleMember: {
				server.ActionReadObject: []string{server.AuthResource(defaultRoot, "*")},
			},
			// managers can read, commit, and delete objects in the default storage
			// root
			RoleManager: {
				server.ActionReadObject:   []string{server.AuthResource(defaultRoot, "*")},
				server.ActionCommitObject: []string{server.AuthResource(defaultRoot, "*")},
				server.ActionDeleteObject: []string{server.AuthResource(defaultRoot, "*")},
			},
			// admins can do anything to objects in any storage root
			RoleAdmin: {"*": []string{"*"}},
		},
	}
}

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

func AuthUserFunc() server.AuthUserFunc { return server.JWSAuthFunc(&testKey().PublicKey) }

// SetUserToken modifies the client to include a bearer token
// for the given user. The token is signed with testKey.
func SetUserToken(cli *http.Client, user server.AuthUser) {
	if cli.Transport == nil {
		cli.Transport = http.DefaultTransport
	}
	if existing, ok := cli.Transport.(*bearerTokenTransport); ok {
		existing.Token = authUserToken(user)
		return
	}
	cli.Transport = &bearerTokenTransport{
		Token: authUserToken(user),
		Base:  cli.Transport,
	}
}

// authUserToken generates a token for the given user signed with the test kßey.
func authUserToken(user server.AuthUser) string {
	key := testKey()
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: key}, (&jose.SignerOptions{}).WithType("JWT"))
	if err != nil {
		panic(fmt.Errorf("user token signing: %v", err))
	}
	now := time.Now()
	token := server.AuthToken{
		User: user,
		Claims: jwt.Claims{
			Issuer:    issuer,
			Subject:   user.ID,
			Audience:  jwt.Audience{issuer},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-1 * time.Hour)),
			Expiry:    jwt.NewNumericDate(now.Add(1 * time.Hour)),
		},
	}
	encToken, err := jwt.Signed(signer).Claims(token).Serialize()
	if err != nil {
		panic(fmt.Errorf("user token signing: %v", err))
	}
	return encToken
}

type bearerTokenTransport struct {
	Token string
	Base  http.RoundTripper
}

func (t *bearerTokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.Token)
	return t.Base.RoundTrip(req)
}
