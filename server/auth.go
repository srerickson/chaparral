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

	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
)

const (
	// actions
	ActionReadObject   = "read_object"
	ActionCommitObject = "commit_object"
	ActionDeleteObject = "delete_object"
	ActionUpload       = "upload_files"
	// ActionAdminister   = "administer"

	rolePrefix = "chaparral"

	// built-in user roles

	// The RoleDefault can be used to assign permissions to all users, even
	// un-authenticated ones. The default role is attached to users implicitly.
	// It doesn't need to be included in the user roles.
	RoleDefault = rolePrefix + ":default"
	RoleMember  = rolePrefix + ":member"
	RoleManager = rolePrefix + ":manager"
	RoleAdmin   = rolePrefix + ":admin"
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
		sig, err := jose.ParseSigned(encToken)
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

// Permissions is a map of roles to permissions. It implements the Authorizer
// interface.
type Permissions map[string]Permission

// RootActionAllowed returns true if the user has a role with a permission
// allowing the action on the resource with the given group and root ids.
func (p Permissions) Allowed(ctx context.Context, action string, resource string) bool {
	user := AuthUserFromCtx(ctx)
	roles := append(user.Roles, RoleDefault)
	return slices.ContainsFunc(roles, func(r string) bool {
		perm, ok := p[r]
		if !ok {
			return false
		}
		return perm.allow(action, resource)
	})
}

// func (p Permissions) ActionAllowed(_ context.Context, user *AuthUser, action string) bool {
// 	roles := []string{DefaultRole}
// 	if user != nil {
// 		roles = append(roles, user.Roles...)
// 	}
// 	return slices.ContainsFunc(roles, func(r string) bool {
// 		return slices.ContainsFunc(p[r], func(rp RolePermission) bool {
// 			return rp.allowAction(action)
// 		})
// 	})
// }

type Permission map[string][]string

func (p Permission) allow(action string, resource string) bool {
	for _, act := range []string{action, "*"} {
		ok := slices.ContainsFunc(p[act], func(permitResource string) bool {
			return resource == permitResource
		})
		if ok {
			return true
		}
	}
	return false
}

// DefaultPermissions returns the default server Permissions.
func DefaultPermissions() Permissions {
	return Permissions{
		// No access for un-authenticated users
		RoleDefault: Permission{},
		// members can read objects in the default storage root
		RoleMember: Permission{
			ActionReadObject: []string{},
		},
		// managers can read, commit, and delete objects in the default storage
		// root
		RoleManager: Permission{
			ActionReadObject:   []string{},
			ActionCommitObject: []string{},
			ActionDeleteObject: []string{},
		},
		// admins can do anything to objects in any storage root
		RoleAdmin: Permission{
			"*": []string{"*"},
		},
	}
}
