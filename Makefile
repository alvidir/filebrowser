VERSION=0.1.0
PROJECT=filebrowser
REPO=alvidir

build:
	podman build -t ${REPO}/${PROJECT}:${VERSION} -f ./container/filebrowser/containerfile .

deploy:
	podman-compose -f container-compose.yaml up --remove-orphans
	# delete -d in order to see output logs

undeploy:
	podman-compose -f container-compose.yaml down

run:
	go run cmd/filebrowser/main.go

test:
	go test -v -race ./...