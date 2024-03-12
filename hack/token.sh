PEM_FILE='hack/data/auth.pem'
USER_ID="1"
USER_EMAIL="nobody@nothing.never"
USER_NAME="test user"
USER_ROLES="chaparral_admin"

# generate signed client bearer token
echo "client bearer token:"
go run ./cmd/chaptoken -key "$PEM_FILE" -id "$USER_ID" -email "$USER_EMAIL" -name "$USER_NAME" -roles "$USER_ROLES" -exp 1
