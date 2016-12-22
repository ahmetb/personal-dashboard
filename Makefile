all: binaries

binaries: check-env clean
	set -euo pipefail; set -x; \
	cmds=$$(go list ./cmd/...); \
	for d in $$cmds; do \
		GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
			govvv build -a -tags netgo -ldflags '-w' -o "./bin/$$(basename $$d)" "$$d"; \
	done;

clean: check-env
	go clean
	rm -rf "$(GOPATH)/bin" "./bin"

check-env:
ifndef GOPATH
	$(error GOPATH is unset)
endif

docker-image: binaries
	docker build -t personal-dashboard . 
