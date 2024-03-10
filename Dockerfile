FROM sqlc/sqlc:1.25.0 as sqlc
FROM bufbuild/buf:1.28.1 as buf

FROM cgr.dev/chainguard/go:latest as builder
WORKDIR /work
COPY . ./

# add sqlc and buf command to builder
COPY --from=sqlc /workspace/sqlc /usr/local/bin/sqlc
COPY --from=buf /usr/local/bin/buf /usr/local/bin/buf

# code gen
RUN buf lint proto
RUN buf generate proto
RUN sqlc generate -f ./server/chapdb/sqlc.yaml

RUN go mod download
RUN go build -o chaparral ./cmd/chaparral
RUN go build -o chaptoken ./cmd/chaptoken

FROM cgr.dev/chainguard/glibc-dynamic:latest
# FROM cgr.dev/chainguard/glibc-dynamic:latest-dev
COPY --chown=nonroot:nonroot config.yaml /data/config.yaml
COPY --from=builder /work/chaparral chaparral
COPY --from=builder /work/chaptoken chaptoken

EXPOSE 8080

WORKDIR /data
CMD ["/chaparral","-c","config.yaml"]
