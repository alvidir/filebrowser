VERSION=0.1.0
PROJECT=filebrowser
REPO=alvidir

install:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

proto:
	protoc --proto_path=proto --go_out=proto --go_opt=paths=source_relative \
		--go-grpc_out=proto --go-grpc_opt=paths=source_relative \
		proto/*.proto

	go mod tidy

build:
	podman build -t ${REPO}/${PROJECT}:${VERSION} -f ./container/filebrowser/containerfile .

deploy:
	podman-compose -f compose.yaml up --remove-orphans -d

undeploy:
	podman-compose -f compose.yaml down

run:
	go run cmd/filebrowser/main.go

test:
	go test -v -race ./...