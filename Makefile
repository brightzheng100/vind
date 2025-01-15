
# Variables
BINARY_PATH 	:= ./bin
BINARY_NAME 	:= vind
VERSION_TAG 	:= $(shell git describe --abbrev=0 --tags)
VERSION_COMMIT	:= $(shell git rev-parse --short HEAD)
VERSION_DATE	:= $(shell date +%Y-%m-%dT%H:%M:%SZ)
GOFLAGS 		:= -ldflags="-X 'github.com/brightzheng100/vind/cmd.version=${VERSION_TAG}' -X 'github.com/brightzheng100/vind/cmd.commit=${VERSION_COMMIT}' -X 'github.com/brightzheng100/vind/cmd.date=${VERSION_DATE}'"

.PHONY: build
build:
	# MacOS
	GOARCH=amd64 GOOS=darwin go build ${GOFLAGS} -o ${BINARY_PATH}/${BINARY_NAME}_${VERSION_TAG}_darwin_amd64 main.go
	GOARCH=arm64 GOOS=darwin go build ${GOFLAGS} -o ${BINARY_PATH}/${BINARY_NAME}_${VERSION_TAG}_darwin_arm64 main.go
	# Linux
	GOARCH=amd64 GOOS=linux go build ${GOFLAGS} -o ${BINARY_PATH}/${BINARY_NAME}_${VERSION_TAG}_linux_amd64 main.go
	GOARCH=arm64 GOOS=linux go build ${GOFLAGS} -o ${BINARY_PATH}/${BINARY_NAME}_${VERSION_TAG}_linux_arm64 main.go
	# Windows
	#GOARCH=amd64 GOOS=windows go build ${GOFLAGS} -o ${BINARY_PATH}/${BINARY_NAME}_${VERSION_TAG}_windows_amd64 main.go

.PHONY: clean
clean:
	go clean
	rm ${BINARY_PATH}/${BINARY_NAME}_${VERSION_TAG}_darwin_*
	rm ${BINARY_PATH}/{BINARY_NAME}_${VERSION_TAG}_linux_*
	rm ${BINARY_PATH}/${BINARY_NAME}_${VERSION_TAG}_windows_*

.PHONY: test
test:
	go test ./...

.PHONY: test_coverage
test_coverage:
	go test ./... -coverprofile=coverage.out

.PHONY: dep
dep:
	go mod download

.PHONY: vet
vet:
	go vet

.PHONY: lint
lint:
	# Install it by: 
	# curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
	#	sh -s -- -b $(go env GOPATH)/bin v1.61.0
	golangci-lint run --enable-all
