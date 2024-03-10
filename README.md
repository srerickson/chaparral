# chaparral

Chaparral is an API service for storing and accessing versioned
datasets. It uses the [Oxford Common File Layout](https://ocfl.io) (OCFL)
specification for storing, versioning, and validating content. File system
and S3 backends are supported.

> [!Warning] 
> This project is in early development. Expect bugs and breaking changes.


## Server

The server is distributed as a container image
([srerickson/chaparral](https://hub.docker.com/repository/docker/srerickson/chaparral/general)).
To run chaparral on :8080 using `hack/data` for persistence:

```sh
docker run --rm -v $(pwd)/hack/data:/data -p 8080:8080 srerickson/chaparral:latest
```

On first run, a new default OCFL storage root is initialized if one doesn't
exist. In addition, the server will create a new sqlite3 database for internal
state and an RSA key for signing auth tokens.

See [`config.yaml`](config.yaml) for configuration options.

## API 

Chaparral's server API is defined using protocol buffers/gRPC and implemented
using [connect-go](https://github.com/connectrpc/connect-go). It supports both
gRPC and http/1.1 requests. Documentation is available through the [Buf schema
registry](https://buf.build/srerickson/chaparral/docs/main:chaparral.v1). 

## About the name

> Chaparral is a shrubland plant community found primarily in California, in
> southern Oregon and in the northern portion of the Baja California Peninsula
> in Mexico. It is shaped by a Mediterranean climate (mild wet winters and hot
> dry summers) and infrequent, high-intensity crown fires.
> ([Wikipedia](https://en.wikipedia.org/wiki/Chaparral))
