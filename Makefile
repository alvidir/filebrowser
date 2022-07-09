VERSION=0.1.0
PROJECT=filebrowser
REPO=alvidir
REMOTE=docker.io

install:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

proto:
	protoc --proto_path=proto --go_out=proto --go_opt=paths=source_relative \
		--go-grpc_out=proto --go-grpc_opt=paths=source_relative \
		proto/*.proto

	go mod tidy

release: build push

build:
	podman build -t ${REPO}/${PROJECT}:${VERSION} -f ./container/filebrowser/containerfile .
	podman build -t ${REPO}/${PROJECT}:${VERSION}-mq-users -f ./container/mq-users/containerfile .

push:
	podman tag localhost/${REPO}/${PROJECT}:${VERSION} ${REMOTE}/${REPO}/${PROJECT}:${VERSION}
	podman push ${REMOTE}/${REPO}/${PROJECT}:${VERSION}
	podman tag localhost/${REPO}/${PROJECT}:${VERSION}-mq-users ${REMOTE}/${REPO}/${PROJECT}:${VERSION}-mq-users
	podman push ${REMOTE}/${REPO}/${PROJECT}:${VERSION}-mq-users

deploy:
	podman-compose -f compose.yaml up -d

follow:
	podman logs --follow --names filebrowser-server

undeploy:
	podman-compose -f compose.yaml down

run:
	go run cmd/filebrowser/main.go

all-mqworkers:
	go run cmd/mq-users/main.go

test:
	go test -v -race ./...