BUILDDIR ?= build

.PHONY: build
build: $(BUILDDIR)/image-clone-controller

.PHONY: clean
clean:
	rm -f $(BUILDDIR)/*

.PHONY: deps
deps:
	go mod download

$(BUILDDIR)/image-clone-controller: deps
	go build -o $(BUILDDIR)/image-clone-controller ./cmd/image-clone-controller

.PHONY: test
test:
	docker run --rm -d --name registry -p 5000:5000 registry:2
	go test -race ./...
	docker stop registry
