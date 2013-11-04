PROGRAM_NAME := tasktogo
VERSION = $(shell git describe --dirty=+)

GOCOMPILER = go build
GOFLAGS	+= -ldflags "-X main.Version $(VERSION)"

PREFIX=/usr

.PHONY: all install clean

all: $(PROGRAM_NAME)

$(PROGRAM_NAME): $(wildcard *.go)
	$(GOCOMPILER) $(GOFLAGS)

install: all
	install -m 0755 $(PROGRAM_NAME) $(PREFIX)/bin

clean:
	@- $(RM) $(PROGRAM_NAME)
