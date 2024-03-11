# start a local server
echo "starting server..."
CHAPARRAL_BACKEND=file://hack/data \
CHAPARRAL_PUBKEY_FILE='hack/data/pub.pem'  \
CHAPARRAL_DB=hack/data/chaparral.sqlite3 \
CHAPARRAL_LISTEN=":8080" \
CHAPARRAL_DEBUG=true \
go run ./cmd/chaparral -c config.yaml
