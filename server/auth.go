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
	ReadAction   string = "read"
	CommitAction string = "write"
	DeleteAction string = "delete"
	AdminAction  string = "administer"

	rolePrefix = "chaparral"

	// built-in user roles

	// The DefaultRole can be used to assign permissions to all users, even
	// un-authenticated ones. The default role is attached to users implicitly.
	// It doesn't need to be included in the user roles.
	DefaultRole = rolePrefix + ":default"
	MemberRole  = rolePrefix + ":member"
	ManagerRole = rolePrefix + ":manager"
	AdminRole   = rolePrefix + ":admin"
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
			newLogger.Debug("setting user role for request")
			newCtx := CtxWithAuthUser(CtxWithLogger(r.Context(), newLogger), user)
			next.ServeHTTP(w, r.WithContext(newCtx))
		}
		return fn
	}
}

// Authorizer is an interface used by types that can perform authorziation
// for requests.
// TODO: simplify this interface by removing user arg from methods; get
// the user from the ctx.
type Authorizer interface {
	// RootActionAllowed returns true if the user is allowed to perform action
	// on the resource with the given root_id.
	RootActionAllowed(ctx context.Context, user *AuthUser, action, root_id string) bool
	// ActionAllowed return true if the user has permission to perform action
	// on at least one resource.
	ActionAllowed(ctx context.Context, user *AuthUser, action string) bool
}

// Permissions is a map of roles to permissions. It implements the Authorizer
// interface.
type Permissions map[string][]RolePermission

// RootActionAllowed returns true if the user has a role with a permission
// allowing the action on the resource with the given group and root ids.
func (p Permissions) RootActionAllowed(_ context.Context, user *AuthUser, action, root string) bool {
	roles := []string{DefaultRole}
	if user != nil {
		roles = append(roles, user.Roles...)
	}
	return slices.ContainsFunc(roles, func(r string) bool {
		return slices.ContainsFunc(p[r], func(rp RolePermission) bool {
			return rp.allowRootAction(action, root)
		})
	})
}

func (p Permissions) ActionAllowed(_ context.Context, user *AuthUser, action string) bool {
	roles := []string{DefaultRole}
	if user != nil {
		roles = append(roles, user.Roles...)
	}
	return slices.ContainsFunc(roles, func(r string) bool {
		return slices.ContainsFunc(p[r], func(rp RolePermission) bool {
			return rp.allowAction(action)
		})
	})
}

type RolePermission struct {
	Actions       []string `json:"actions"`
	StorageRootID string   `json:"storage_root_id"`
}

func (p RolePermission) allowAction(action string) bool {
	return slices.ContainsFunc(p.Actions, func(a string) bool {
		return a == "*" || a == action
	})
}

func (p RolePermission) allowRootAction(action, root string) bool {
	if !p.allowAction(action) {
		return false
	}
	return (p.StorageRootID == "*" || p.StorageRootID == root)
}

// DefaultPermissions returns the default server Permissions.
func DefaultPermissions() Permissions {
	return Permissions{
		// No access for un-authenticated users
		DefaultRole: []RolePermission{},
		// members can read objects in the default storage root
		MemberRole: []RolePermission{
			{Actions: []string{ReadAction}},
		},
		// managers can read, commit, and delete objects in the default storage
		// root
		ManagerRole: []RolePermission{
			// storage root
			{Actions: []string{ReadAction, CommitAction, DeleteAction}},
		},
		// admins can do anything to objects in any storage root
		AdminRole: []RolePermission{
			{Actions: []string{"*"}, StorageRootID: "*"},
		},
	}
}
