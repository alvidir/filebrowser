install: go-install
	sudo dnf install openssl-devel

go-install:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

proto:
	protoc --proto_path=. --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/*.proto

	go mod tidy

build:
	podman build -t alvidir/filebrowser:latest-grpc -f ./container/grpc/containerfile .
	podman build -t alvidir/filebrowser:latest-rest -f ./container/rest/containerfile .
	podman build -t alvidir/filebrowser:latest-agent -f ./container/agent/containerfile .

setup:
	mkdir -p .ssh/

	openssl ecparam -name prime256v1 -genkey -noout -out .ssh/ec_key.pem
	openssl pkcs8 -topk8 -nocrypt -in .ssh/ec_key.pem -out .ssh/pkcs8_key.pem
	
	cat .ssh/pkcs8_key.pem | base64 | tr -d '\n' > .ssh/pkcs8_key.base64
	
deploy:
	podman-compose -f compose.yaml up -d

undeploy:
	podman-compose -f compose.yaml down

follow:
	podman logs --follow --names $(srv)

test:
	go test -v -race ./...