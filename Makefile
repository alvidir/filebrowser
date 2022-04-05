VERSION=0.1.0
PROJECT=filebrowser
REPO=alvidir

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