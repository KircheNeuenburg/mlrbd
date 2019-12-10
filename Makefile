.POSIX:
.SUFFIXES:
.SUFFIXES: .1 .5 .7 .1.scd .5.scd .7.scd

PREFIX?=/usr
_INSTDIR=$(DESTDIR)$(PREFIX)
BINDIR?=$(_INSTDIR)/bin
GO?=go
GOFLAGS?=

GOSRC!=find . -name '*.go'
GOSRC+=go.mod go.sum

generate:
	$(GO) generate
mlrbd: generate
	$(GO) build -o $@ main.go

all: mlrbd

# Exists in GNUMake but not in NetBSD make and others.
RM?=rm -f

clean:
	$(RM) mlrbd

install: all
	mkdir -m755 -p /etc/mlrbd
	install -m755 mlrbd $(BINDIR)/mlrbd
	install -m644 config/mlrbd.conf /etc/mlrbd/mlrbd.conf

RMDIR_IF_EMPTY:=sh -c '\
if test -d $$0 && ! ls -1qA $$0 | grep -q . ; then \
	rmdir $$0; \
fi'

uninstall:
	$(RM) $(BINDIR)/mlrbd
	$(RM) -r /etc/mlrbd

.DEFAULT_GOAL := all

.PHONY: all clean install uninstall
