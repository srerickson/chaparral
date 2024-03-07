//

package server

import (
	"context"
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

	permSep = "::"
)

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

// JWSAuthFunc returns an Authentication func that looks
// for a jwt bearer token signed with the public key.
func JWSAuthFunc(pubkey any) AuthUserFunc {
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
		payload, err := sig.Verify(pubkey)
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

// RolePermissions is a map of role names to Permissions. It implements the Authorizer
// interface.
type RolePermissions struct {
	// Default permissions that apply to all users and un-authenticated requests
	Default Permissions            `json:"default"`
	Roles   map[string]Permissions `json:"roles"`
}

func AuthResource(parts ...string) string {
	return strings.Join(parts, permSep)
}

// Allowed returns true if the user associated with the context has a role with a permission
// allowing the action on the resource. If resource is '*', Allowed returns true if
// the if the action is allowed for any resource.
func (r RolePermissions) Allowed(ctx context.Context, action string, resource string) bool {
	user := AuthUserFromCtx(ctx)
	if r.Default.allow(action, resource) {
		return true
	}
	return slices.ContainsFunc(user.Roles, func(role string) bool {
		perm, ok := r.Roles[role]
		if !ok {
			return false
		}
		return perm.allow(action, resource)
	})
}

// Permissions maps actions to resources for which the action is allowed.
type Permissions map[string][]string

func (p Permissions) allow(action string, resource string) bool {
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
