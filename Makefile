VERSION ?= dev
GOFLAGS ?= -trimpath

.PHONY: build test check browser clean release
build:
	mkdir -p bin
	CGO_ENABLED=0 go build $(GOFLAGS) -ldflags="-s -w -buildid= -X main.version=$(VERSION)" -o bin/horizon-winner ./cmd/horizon-winner
	CGO_ENABLED=0 go build $(GOFLAGS) -ldflags="-s -w -buildid=" -o bin/lab-target ./cmd/lab-target
test:
	go test -race -coverprofile=coverage.out ./...
	go vet ./...
	npm test
	npm run check:js
check:
	go run ./cmd/horizon-winner -config configs/local.example.json -check
	go run ./cmd/horizon-winner -config configs/public.example.json -check
browser:
	npm run test:browser
release:
	./scripts/release.sh "$(VERSION)"
clean:
	go clean
