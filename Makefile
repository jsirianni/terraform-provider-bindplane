ifeq (, $(shell which docker))
    $(error "No docker in $(PATH)")
endif

VERSION := $(shell cat main.go | grep "const version" | cut -c 17- | tr -d '"')

$(shell mkdir -p artifacts)

build: clean fmt
	$(info building bindplane-terraform-provider ${VERSION})

	@docker build \
	    --no-cache \
	    --build-arg version=0.1.0 \
	    -t btp:0.1.0 .

	@docker create -ti --name artifacts btp:0.1.0 bash && \
	    docker cp artifacts:/src/bindplane-terraform-provider/terraform-provider-bindplane_linux_amd64_v0.1.0.zip artifacts/terraform-provider-bindplane_linux_amd64_v0.1.0.zip && \
	    docker cp artifacts:/src/bindplane-terraform-provider/terraform-provider-bindplane_darwin_amd64_v0.1.0.zip artifacts/terraform-provider-bindplane_darwin_amd64_v0.1.0.zip && \
	    docker cp artifacts:/src/bindplane-terraform-provider/terraform-provider-bindplane_windows_amd64_v0.1.0.zip artifacts/terraform-provider-bindplane_windows_amd64_v0.1.0.zip && \
	    docker cp artifacts:/src/bindplane-terraform-provider/SHA256SUMS artifacts/SHA256SUMS

	# cleanup
	@docker rm -fv artifacts &> /dev/null

test:
	go test ./...

lint:
	golint ./...

fmt:
	go fmt ./...

quick:
	$(shell env CGO_ENABLED=0 go build -a -ldflags '-w -extldflags "-static"')

clean:
	$(shell rm -rf artifacts/*)
