# chaparral

Chaparral is an API service for storing and accessing versioned
datasets. It uses the [Oxford Common File Layout](https://ocfl.io) (OCFL)
specification for storing, versioning, and validating content. File system
and S3 backends are supported.

> [!Warning] 
> This project is in early development. Expect bugs and breaking changes.


## Server

The server is distributed as a container image: [srerickson/chaparral](https://hub.docker.com/repository/docker/srerickson/chaparral/general)).
See [`config.yaml`](config.yaml) for an example configuration

## API 

Chaparral's server API is defined using protocol buffers/gRPC and implemented
using [connect-go](https://github.com/connectrpc/connect-go). It supports both
gRPC and http/1.1 requests. Documentation is available through the [Buf schema
registry](https://buf.build/srerickson/chaparral/docs/main:chaparral.v1). 

## Building Images

Images are built with `ko`

```sh
 VERSION=0.x.y KO_DOCKER_REPO=srerickson ko -B build ./cmd/chaparral
```

## Authorization

Chaparral can use signed web tokens for authentication by setting the
`CHAPARRAL_PUBKEY_FILE` environment variable or the `pubkey_file` config value
to the path of a PEM-encoded RSA public key.

Generate a new RSA key pair:

```sh
# generate a new key
$ openssl genrsa -out auth.pem 2048

# export public key
$openssl pkey -in auth.pem -pubout > auth-pub.pem
```

## About the name

> Chaparral is a shrubland plant community found primarily in California, in
> southern Oregon and in the northern portion of the Baja California Peninsula
> in Mexico. It is shaped by a Mediterranean climate (mild wet winters and hot
> dry summers) and infrequent, high-intensity crown fires.
> ([Wikipedia](https://en.wikipedia.org/wiki/Chaparral))
