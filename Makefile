CONTAINER_BIN=docker
RUN_CONTAINER=$(CONTAINER_BIN) run --rm --security-opt label=disable
MINIO_CONTAINER=minio-chaparral-test

.PHONY: proto minio-start minio-stop go-test-s3

test: proto minio-start go-test-s3 minio-stop

minio-start:
	$(RUN_CONTAINER) -d --name $(MINIO_CONTAINER)\
		-p 9000:9000 \
		-v "$(shell pwd)/testdata/minio:/data" \
		quay.io/minio/minio server /data

minio-stop:
	$(CONTAINER_BIN) stop $(MINIO_CONTAINER)

clean: 
	rm -rf ./gen/*

proto: clean
	$(RUN_CONTAINER) -v "$(shell pwd):/workspace" \
		--workdir /workspace bufbuild/buf lint proto
	$(RUN_CONTAINER) -v "$(shell pwd):/workspace" \
		--workdir /workspace bufbuild/buf generate proto
		
sqlc:
	sqlc generate -f server/chapdb/sqlc.yaml

go-test-s3:
	CHAPARRAL_TEST_S3="http://localhost:9000" go test -race -count=10 -parallel=5 ./...

# build:
# 	go build -o ./cmd/ocfld ./cmd/ocfld
# 	go build -o ./cmd/oxctl ./cmd/oxctl