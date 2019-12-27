NAME       := jobs-queue
DESTDIR    := /opt
INSTALLDIR := $(DESTDIR)/$(NAME)

ifeq ($(GITHUB_REF),)
GIT_VER    := $(shell git describe --abbrev=7 --always --tags)-$(shell git rev-parse --abbrev-ref HEAD)-$(shell date +%Y%m%d)
else
GIT_VER    := $(shell basename $(GITHUB_REF))-$(shell date +%Y%m%d)
endif
LDFLAGS    := -ldflags "-X main.version=$(GIT_VER)"

.PHONY: lint clean archive install
lint:
	find ./app/ -type f -name '*.go' | xargs gofmt -l -e
	go vet -mod=vendor ./app/...
	$(shell go env GOPATH)/bin/golint ./app/...
	go test -mod=vendor ./app/...

bin/$(NAME):
	go build -mod=vendor -v $(LDFLAGS) -o $@ ./app

clean:
	rm -rf bin/$(NAME) tmp/$(NAME)

release: clean dist
	make DESTDIR=./tmp install
	tar -cvzf dist/$(NAME)_$(GIT_VER)_x86-64.tar.gz --owner=0 --group=0 -C ./tmp $(NAME)

$(INSTALLDIR) dist tmp:
	mkdir -p $@

install: $(INSTALLDIR) bin/$(NAME)
	install -m 0755 bin/$(NAME) $(INSTALLDIR)
	install -m 0600 config/config.dist.yaml $(INSTALLDIR)/config.yaml
