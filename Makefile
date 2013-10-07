PROGRAM_NAME := tasktogo
GOCOMPILER = go build
GOFLAGS	+= -ldflags "-X main.Version $(shell git describe --dirty=+)"

.PHONY: all clean deps

all: $(PROGRAM_NAME)

$(PROGRAM_NAME): $(wildcard *.go)
	$(GOCOMPILER) $(GOFLAGS)

clean:
	@- $(RM) $(PROGRAM_NAME)
