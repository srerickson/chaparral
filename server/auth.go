//

package server

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
)

const (
	// actions
	ActionReadObject   = "read_object"
	ActionCommitObject = "commit_object"
	ActionDeleteObject = "delete_object"
	// ActionAdminister   = "administer"

	rolePrefix = "chaparral"

	// built-in user roles

	// RoleDefault can be used to assign permissions to all users, even
	// un-authenticated ones. The default role is attached to users implicitly.
	// It doesn't need to be included in the user's list of roles.
	RoleDefault = rolePrefix + ":default"

	RoleMember  = rolePrefix + ":member"
	RoleManager = rolePrefix + ":manager"
	RoleAdmin   = rolePrefix + ":admin"

	permSep = "::"
)

// var pkenv = strings.ToUpper(rolePrefix) + "_JWK"

type userCtxKey struct{}

type AuthUser struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Email string   `json:"email"`
	Roles []string `json:"roles"`
}

func (u AuthUser) Empty() bool {
	return u.ID == ""
}

func CtxWithAuthUser(ctx context.Context, user AuthUser) context.Context {
	return context.WithValue(ctx, userCtxKey{}, user)
}

func AuthUserFromCtx(ctx context.Context) AuthUser {
	u, _ := ctx.Value(userCtxKey{}).(AuthUser)
	return u
}

// AuthUserFunc returns the AuthUser for the request. The AuthUser may be
// empty.
type AuthUserFunc func(*http.Request) (AuthUser, error)

// AuthToken is the JWT bearer token used to authenticate users.
type AuthToken struct {
	jwt.Claims
	User AuthUser `json:"chaparral"`
}

// DefaultAuthUserFunc returns an Authentication func that looks
// for a signed JWT bearer token.
func DefaultAuthUserFunc(pub *rsa.PublicKey) AuthUserFunc {
	auth := func(r *http.Request) (user AuthUser, err error) {
		authHeader := r.Header.Get("Authorization")
		_, encToken, _ := strings.Cut(authHeader, " ")
		if encToken == "" {
			// no header token
			return
		}
		sig, err := jose.ParseSigned(encToken, []jose.SignatureAlgorithm{jose.RS256, jose.RS512})
		if err != nil {
			err = fmt.Errorf("parsing auth token: %w", err)
			return
		}
		payload, err := sig.Verify(pub)
		if err != nil {
			err = fmt.Errorf("auth token signature verification failed: %w", err)
			return
		}
		var token AuthToken
		err = json.Unmarshal(payload, &token)
		if err != nil {
			err = fmt.Errorf("couldn't unmarshal authtoken: %w", err)
			return
		}
		// TODO: validate issuer, subject, etc.
		expected := jwt.Expected{}
		if err = token.ValidateWithLeeway(expected, 2*time.Minute); err != nil {
			err = fmt.Errorf("authentication token has invalid claims: %w", err)
			return
		}
		user = token.User
		return
	}
	return auth
}

func AuthUserMiddleware(authFn AuthUserFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		var fn http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
			logger := LoggerFromCtx(r.Context())
			user, err := authFn(r)
			if err != nil {
				logger.Error("during auth:" + err.Error())
				w.WriteHeader(http.StatusInternalServerError)
			}
			newLogger := logger.With("user_roles", strings.Join(user.Roles, ","))
			newCtx := CtxWithAuthUser(CtxWithLogger(r.Context(), newLogger), user)
			next.ServeHTTP(w, r.WithContext(newCtx))
		}
		return fn
	}
}

// Authorizer is an interface used by types that can perform authorziation
// for requests.
type Authorizer interface {
	// Allowed returns true if the user is allowed to perform action
	// on the resource with the given root_id.
	Allowed(ctx context.Context, action string, resources string) bool
}

// Roles is a map of role names to RolePermissions. It implements the Authorizer
// interface.
type Roles map[string]RolePermissions

func AuthResource(parts ...string) string {
	return strings.Join(parts, permSep)
}

// Allowed returns true if the user associated with the context has a role with a permission
// allowing the action on the resource. If resource is '*', Allowed returns true if
// the if the action is allowed for any resource.
func (r Roles) Allowed(ctx context.Context, action string, resource string) bool {
	user := AuthUserFromCtx(ctx)
	roles := append(user.Roles, RoleDefault)
	return slices.ContainsFunc(roles, func(role string) bool {
		perm, ok := r[role]
		if !ok {
			return false
		}
		return perm.allow(action, resource)
	})
}

type RolePermissions map[string][]string

func (p RolePermissions) allow(action string, resource string) bool {
	for _, act := range []string{action, "*"} {
		ok := slices.ContainsFunc(p[act], func(okResource string) bool {
			return resousrceMatch(resource, okResource)
		})
		if ok {
			return true
		}
	}
	return false
}

func resousrceMatch(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	if a == "*" || b == "*" || a == b {
		return true
	}
	if !strings.Contains(a, permSep) {
		return false
	}
	aParts := strings.Split(a, permSep)
	bParts := strings.Split(b, permSep)
	if len(aParts) != len(bParts) {
		return false
	}
	for i := range aParts {
		if !resousrceMatch(aParts[i], bParts[i]) {
			return false
		}
	}
	return true
}

// DefaultRoles returns the default server Permissions.
// the "chaparral::member" role can access the default storage
// root.
func DefaultRoles(defaultRoot string) Roles {
	return Roles{
		// No access for un-authenticated users
		RoleDefault: RolePermissions{},
		// members can read objects in the default storage root
		RoleMember: RolePermissions{
			ActionReadObject: []string{AuthResource(defaultRoot, "*")},
		},
		// managers can read, commit, and delete objects in the default storage
		// root
		RoleManager: RolePermissions{
			ActionReadObject:   []string{AuthResource(defaultRoot, "*")},
			ActionCommitObject: []string{AuthResource(defaultRoot, "*")},
			ActionDeleteObject: []string{AuthResource(defaultRoot, "*")},
		},
		// admins can do anything to objects in any storage root
		RoleAdmin: RolePermissions{
			"*": []string{"*"},
		},
	}
}
