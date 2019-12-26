NAME       := jobs-queue
DESTDIR    := /opt
INSTALLDIR := $(DESTDIR)/$(NAME)

GIT_VER    := $(shell git describe --abbrev=7 --always --tags)-$(shell git rev-parse --abbrev-ref HEAD)-$(shell date +%Y%m%d)
LDFLAGS    := -ldflags "-X main.version=$(GIT_VER)"

.PHONY: lint
lint:
	find ./app/ -type f -name '*.go' | xargs gofmt -l -e
	go vet -mod=vendor ./app/...
	$(shell go env GOPATH)/bin/golint ./app/...
	go test -mod=vendor ./app/...

bin/$(NAME):
	go build -mod=vendor -v $(LDFLAGS) -o bin/$(NAME) ./app

.PHONY: clean
clean:
	rm -f bin/$(NAME)
	rm -f install/*.retry
	rm -f dist/*.tar.gz

.PHONY: doc
doc:
	godoc -http :6060

$(INSTALLDIR) dist tmp:
	mkdir -p $@

.PHONY: install
install: $(INSTALLDIR) bin/$(NAME)
	install -m 0755 bin/$(NAME) $(INSTALLDIR)
	install -m 0600 config/config.dist.yaml $(INSTALLDIR)/config.yaml

.PHONY: archive
archive: clean dist
	make DESTDIR=./tmp install
	tar -cvzf dist/$(NAME)_$(GIT_VER)_amd64.tar.gz --owner=0 --group=0 -C ./tmp $(NAME)
