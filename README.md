# chaparral

Chaparral is client/server application for storing and accessing versioned
datasets. It uses the [Oxford Common File Layout](https://ocfl.io) (OCFL)
specification for storing, versioning, and validating content. File system
and S3 backends are supported.

> [!Warning] 
> This project is in early development. Expect bugs and breaking changes.


## Server

The server is distributed as a container image: [srerickson/chaparral](https://hub.docker.com/repository/docker/srerickson/chaparral/general)

## API 

Chaparral's server API is defined using protocol buffers/gRPC and implemented
using [connect-go](https://github.com/connectrpc/connect-go). It supports both
gRPC and http/1.1 requests. Documentation is available through the [Buf schema
registry](https://buf.build/srerickson/chaparral/docs/main:chaparral.v1). 

## CLI

This repo includes a command line client called [`chap`](cmd/chap), which supports push/pull
operations for creating, updating, and deleting objects.

## About the name

> Chaparral is a shrubland plant community found primarily in California, in
> southern Oregon and in the northern portion of the Baja California Peninsula
> in Mexico. It is shaped by a Mediterranean climate (mild wet winters and hot
> dry summers) and infrequent, high-intensity crown fires.
> ([Wikipedia](https://en.wikipedia.org/wiki/Chaparral))
