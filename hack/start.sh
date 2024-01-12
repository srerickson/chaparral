# server startup script for testing

TOKEN_KEY='hack/data/chaparral.pem' # will be created
TOKEN_USERID="0"
TOKEN_EMAIL="nobody@nothing.never"
TOKEN_NAME="test user"
TOKEN_ROLES="chaparral:member"

# generate signing key
openssl genrsa -out $TOKEN_KEY

# generate signed client bearer token
echo "client bearer token:"
go run ./cmd/chaptoken -key $TOKEN_KEY -id $TOKEN_USERID -email $TOKEN_EMAIL -name $NAME -roles $TOKEN_ROLES -exp 1

# start a local server
echo "starting server..."
CHAPARRAL_AUTH_PEM=$TOKEN_KEY \
CHAPARRAL_BACKEND=file://hack/data \
CHAPARRAL_DB=hack/data/chaparral.sqlite3 \
CHAPARRAL_LISTEN="127.0.0.1:8080" \
CHAPARRAL_DEBUG=true \
go run ./server/cmd/chaparral -c hack/config.yaml