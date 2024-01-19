SQLC='sqlc/sqlc:1.25.0'
BUF='bufbuild/buf:1.28.1'
CONTAINER_BIN=docker
RUN_CONTAINER=$(CONTAINER_BIN) run --rm --security-opt label=disable
MINIO_CONTAINER=minio-chaparral-test

.PHONY: proto sqcl gen minio-start minio-stop go-test-s3 lint

test: proto sqlc minio-start go-test-s3 minio-stop

minio-start:
	$(RUN_CONTAINER) -d --name $(MINIO_CONTAINER)\
		-p 9000:9000 \
		-v "$(shell pwd)/testdata/minio:/data" \
		quay.io/minio/minio server /data

minio-stop:
	$(CONTAINER_BIN) stop $(MINIO_CONTAINER)

proto:
	rm -rf ./gen/*
	$(RUN_CONTAINER) -v "$(shell pwd):/workspace" \
		--workdir /workspace bufbuild/buf lint proto
	$(RUN_CONTAINER) -v "$(shell pwd):/workspace" \
		--workdir /workspace bufbuild/buf generate proto
		
sqlc:
	rm -rf ./server/chapdb/sqlite_gen/*
	$(RUN_CONTAINER) -v "$(shell pwd):/src" \
		--workdir /src $(SQLC) generate -f server/chapdb/sqlc.yaml

gen: proto sqlc

lint:
	$(RUN_CONTAINER) -v $(shell pwd):/app -w /app golangci/golangci-lint:latest golangci-lint run -v

go-test-s3:
	CHAPARRAL_TEST_S3="http://localhost:9000" go test -race -count=10 -parallel=5 ./...
