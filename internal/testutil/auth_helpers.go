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

var (
	// key used for JWS signing/validation in tests
	key *rsa.PrivateKey

	AnonUser = server.AuthUser{}
	// canned users for testing
	MemberUser = server.AuthUser{
		ID:    "test-member",
		Email: "test-member@testing.com",
		Name:  "Test Member",
		Roles: []string{server.RoleMember}}
	ManagerUser = server.AuthUser{
		ID:    "test-manager",
		Email: "test-manager@testing.com",
		Name:  "Test Manager",
		Roles: []string{server.RoleManager}}
	AdminUser = server.AuthUser{
		ID:    "test-admin",
		Email: "test-admin@testing.com",
		Name:  "Test Admin",
		Roles: []string{server.RoleAdmin}}

	// canned permissions used in testing
	AuthorizeAll = server.Roles{
		// anyone can do anythong
		server.RoleDefault: server.RolePermissions{
			"*": []string{"*"},
		},
	}
	AuthorizeNone = server.Roles{}

	// default
	AuthorizeDefaults = server.DefaultRoles("test")
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

func AuthUserFunc() server.AuthUserFunc { return server.DefaultAuthUserFunc(&testKey().PublicKey) }

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
