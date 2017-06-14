#!/usr/bin/make -f 

GO ?= $(shell which go)
ULOGD_SRC ?= /tmp
CGO_CFLAGS ?= "-I$(ULOGD_SRC)/include -I$(ULOGD_SRC)"

sources = main.go plugin/*.go plugin/plugin.c resolver/*.go

# we build two versions of the plugin (actually, two very similar plugins) from the same sources
objects = ${OBJDIR}/filter_DIEGOINSTANCE.so

all: clean $(objects)

$(OBJDIR)/filter_DIEGOINSTANCE.so : $(sources)
	CGO_CFLAGS=$(CGO_CFLAGS) $(GO) build -tags diego -o $@ -buildmode=c-shared .

clean:
	rm -f $(objects)
