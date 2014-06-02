PROGRAM_NAME := tasktogo
VERSION := $(shell git describe --dirty=+)

ifndef GOCOMPILER
GOCOMPILER = go build $(GOFLAGS)
endif
GOFLAGS	+= -ldflags "-X main.Version $(VERSION)"

# If the prefix is not yet defined, define it here.
ifndef prefix
prefix = /usr/local
endif

.PHONY: all install clean

all: $(PROGRAM_NAME) man

$(PROGRAM_NAME): $(wildcard *.go)
	$(GOCOMPILER) -o $(PROGRAM_NAME)

# Compile man page sources in `doc` to `man`.
man: doc/tasktogo.1
	test -d man || mkdir -p man
	gzip -c doc/tasktogo.1 > man/tasktogo.1.gz

install: all
	test -d $(prefix)/bin || mkdir -p $(prefix)/bin
	test -d $(prefix)/share/man/man1 || mkdir -p $(prefix)/share/man/man1
	install -m 0755 $(PROGRAM_NAME) $(prefix)/bin
	install -m 0644 man/tasktogo.1.gz $(prefix)/share/man/man1

clean:
	- rm -rf $(PROGRAM_NAME)
	- rm -rf man
