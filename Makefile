default: install

generate:
	go generate ./...

build:
	go build -trimpath -ldflags "-s -w" -o terraform-provider-sftpgo

install:
	go install .

test:
	go test -count=1 -parallel=4 ./...

testacc:
	TF_ACC=1 go test -count=1 -parallel=4 -timeout 10m -v ./...
