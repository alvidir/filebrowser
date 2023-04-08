BINARY_NAME=filebrowser
VERSION?=latest
PKG_MANAGER?=dnf
REGISTRY?=docker.io

all: binaries 

binaries: protobuf
ifdef target
	GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -a -installsuffix cgo -o bin/$(target)/$(BINARY_NAME)-$(target) cmd/$(target)/main.go
else
	-GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -a -installsuffix cgo -o bin/grpc/$(BINARY_NAME)-grpc cmd/grpc/main.go
	-GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -a -installsuffix cgo -o bin/rest/$(BINARY_NAME)-rest cmd/rest/main.go
	-GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -a -installsuffix cgo -o bin/agent/$(BINARY_NAME)-agent cmd/agent/main.go
endif

images:
ifdef target
	podman build -t alvidir/$(BINARY_NAME):$(VERSION)-$(target) -f ./container/$(target)/containerfile .
else
	-podman build -t alvidir/$(BINARY_NAME):$(VERSION)-grpc -f ./container/grpc/containerfile .
	-podman build -t alvidir/$(BINARY_NAME):$(VERSION)-rest -f ./container/rest/containerfile .
	-podman build -t alvidir/$(BINARY_NAME):$(VERSION)-agent -f ./container/agent/containerfile .
endif

push-images:
ifdef target
	@podman push alvidir/$(BINARY_NAME):$(VERSION)-$(target)
else
	@-podman push alvidir/$(BINARY_NAME):$(VERSION)-grpc
	@-podman push alvidir/$(BINARY_NAME):$(VERSION)-rest
	@-podman push alvidir/$(BINARY_NAME):$(VERSION)-agent
endif

protobuf: install-deps
	@protoc --proto_path=. --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/*.proto
	@go mod tidy

install-deps:
	-$(PKG_MANAGER) install -y protobuf-compiler
	-$(PKG_MANAGER) install -y openssl-devel
	-$(PKG_MANAGER) install -y libssl-dev
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	
clean:
	@-go clean
	@-rm -rf bin/
	@-rm -rf proto/*.pb.go                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         o
	@-rm -rf secrets/

clean-images:
	@-podman image rm alvidir/$(BINARY_NAME):$(VERSION)-grpc
	@-podman image rm alvidir/$(BINARY_NAME):$(VERSION)-rest
	@-podman image rm alvidir/$(BINARY_NAME):$(VERSION)-agent

test: protobuf
	@go test -v -race ./...

secrets:
	@mkdir -p secrets/
	@openssl ecparam -name prime256v1 -genkey -noout -out secrets/ec_key.pem
	@openssl pkcs8 -topk8 -nocrypt -in secrets/ec_key.pem -out secrets/pkcs8_key.pem
	@cat secrets/pkcs8_key.pem | base64 | tr -d '\n' > secrets/pkcs8_key.base64

deploy:
	@podman-compose -f compose.yaml up -d

undeploy:
	@podman-compose -f compose.yaml down
