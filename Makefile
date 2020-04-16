PROJECTNAME = locker

# paths
PREFIX = /usr/local

# Go related variables.
GOBASE := $(shell pwd)
GOCMD := $(GOBASE)/cmd/locker
GOPATH := $(GOBASE)/vendor:$(GOBASE)
GOBIN := $(GOBASE)/bin

LDFLAGS=-trimpath -mod=readonly

all: options build

## options: print build options
options:
	@echo "$(PROJECTNAME) build options:"
	@echo "GOBASE  = $(GOBASE)"
	@echo "GOPATH  = $(GOPATH)"
	@echo "GOBIN   = $(GOBIN)"
	@echo "LDFLAGS = $(LDFLAGS)"

## get: Install missing dependencies. e.g; make get get=github.com/foo/bar
get:
	cd $(GOCMD) && GOPATH=$(GOPATH) GOBIN=$(GOBIN) go get $(get)
	chmod u+w -R $(GOBASE)/vendor

## build: Compile the binary, place it in ./bin
build: bin/locker

## install: install the executable to /usr/local/bin and place the config files in their appropriate locations
install:
	@-[ ! -f bin/locker ] && $(MAKE) build
	mkdir -p $(DESTDIR)$(PREFIX)/bin
	install -Dm755 bin/$(PROJECTNAME) $(DESTDIR)$(PREFIX)/bin/$(PROJECTNAME)
	mkdir -p $(DESTDIR)/etc/$(PROJECTNAME)
	install -Dm644 seccomp/seccomp_default.json -t $(DESTDIR)/etc/$(PROJECTNAME)
	mkdir -p $(DESTDIR)/var/lib/$(PROJECTNAME)
	echo {} > $(DESTDIR)/var/lib/$(PROJECTNAME)/images.json
	chmod 644 $(DESTDIR)/var/lib/$(PROJECTNAME)/images.json

## uninstall: removes the executable from /usr/local/bin and delete the config files
uninstall:
	rm -rf $(DESTDIR)$(PREFIX)/bin/$(PROJECTNAME)\
	    	$(DESTDIR)/etc/$(PROJECTNAME)\
		$(DESTDIR)/var/lib/$(PROJECTNAME)

## exec: Run given command, wrapped with custom GOPATH. e.g; make exec run="go test ./..."
exec:
	GOPATH=$(GOPATH) GOBIN=$(GOBIN) $(run)

## clean: Clean build files.
clean:
	rm -rf $(GOBIN)
	rm -rf $(GOBASE)/vendor

bin/locker: get
	@echo "Building binary..."
	cd $(GOCMD) && GOPATH=$(GOPATH) GOBIN=$(GOBIN) go build $(LDFLAGS) -o $(GOBIN)/$(PROJECTNAME)

go-clean:
	@echo "Cleaning build cache"
	GOPATH=$(GOPATH) GOBIN=$(GOBIN) go clean

## help: print this help message
help: Makefile
	@echo "Choose a command run in make:"
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: all options get build install uninstall clean go-clean help
