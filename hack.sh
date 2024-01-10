TOKEN_KEY='hack/server/chaparral.pem'
TOKEN_USERID="0"
TOKEN_EMAIL="nobody@nothing.never"
TOKEN_NAME="test user"
TOKEN_ROLES="chaparral:member"

openssl genrsa -out $TOKEN_KEY

echo "client bearer token:"
go run ./cmd/chaptoken -key $TOKEN_KEY -id $TOKEN_USERID -email $TOKEN_EMAIL -name $NAME -roles $TOKEN_ROLES -exp 1

echo "starting server..."
# start a local server
CHAPARRAL_AUTH_PEM=$TOKEN_KEY \
CHAPARRAL_BACKEND=file://hack/server \
CHAPARRAL_DB=hack/server/chaparral.sqlite3 \
CHAPARRAL_LISTEN="127.0.0.1:8080" \
CHAPARRAL_DEBUG=true \
go run ./server/cmd/chaparral -c hack/server/roots.yaml