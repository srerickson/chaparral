# start a local server
echo "starting server..."
CHAPARRAL_AUTH_PEM='hack/data/chaparral.pem'  \
CHAPARRAL_BACKEND=file://hack/data \
CHAPARRAL_DB=hack/data/chaparral.sqlite3 \
CHAPARRAL_LISTEN="127.0.0.1:8080" \
CHAPARRAL_DEBUG=true \
go run ./server/cmd/chaparral -c config.yaml