PROGRAM_NAME := tasktogo
VERSION = $(shell git describe --dirty=+)

GOCOMPILER = go build
GOFLAGS	+= -ldflags "-X main.Version $(VERSION)"

PREFIX = /usr

.PHONY: all install clean

all: $(PROGRAM_NAME) man

$(PROGRAM_NAME): $(wildcard *.go)
	$(GOCOMPILER) $(GOFLAGS)

# Compile man page sources in `doc` to `man`.
man: doc/tasktogo.1
	test -d man || mkdir man
	gzip -c doc/tasktogo.1 > man/tasktogo.1.gz

install: all
	install -m 0755 $(PROGRAM_NAME) $(PREFIX)/bin
	install -m 0644 man/tasktogo.1.gz $(PREFIX)/share/man/man1

clean:
	- rm -rf $(PROGRAM_NAME)
	- rm -rf man
