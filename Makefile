#                                                         -*- coding: utf-8 -*-
# File:    ./Makefile
# Authors: Jan Chaloupka
#          Jiří Kučera
# Created: N/A
# Project: Go Fedora Packaging Tooling
# Version: See VERSION.
# Brief:   Makefile.
#

# Commands & tools:
GO_CMD         := go
GO_ENV         := $(GO_CMD) env
GO_BUILD       := $(GO_CMD) build
GO_BUILD_FLAGS :=
GO_TEST        := $(GO_CMD) test
GO_TEST_FLAGS  :=
GLOG_FLAGS     :=
ECHO           := echo
RM             := rm
RM_FLAGS       := -f

# Params:
ifeq (not$(V)set, notset)
V := 0
endif

# Apply params:
# - verbocity level:
V_AUX := $(V)
ifeq ($(V_AUX), 2)
GLOG_FLAGS += -stderrthreshold=INFO
V_AUX := 1
endif
ifeq ($(V_AUX), 1)
GO_BUILD_FLAGS += -v
GO_TEST_FLAGS  += -v
V_AUX := 0
endif

# Setup Go's environment:
# - if not set, get GOROOT from `go env GOROOT`:
ifeq (not$(GOROOT)set, notset)
GOROOT := $(shell $(GO_ENV) GOROOT)
endif
# - if not set, GOPATH will be the current working directory:
ifeq (not$(GOPATH)set, notset)
GOPATH := $(CURDIR)
endif
# GOPATH must be absolute
override GOPATH := $(abspath $(GOPATH))

export GOROOT
export GOPATH

# Locations:
PROJECT_ROOT   := github.com/gofed/symbols-extractor

# Products:
define AddProduct =
$$(eval $$(call _AddProduct,$$(strip $(1)),$$(strip $(2))))
endef
define _AddProduct =
$(1) := $(2)
TRASH += $$($(1))
endef

$(eval $(call AddProduct, EXTRACT,      extract        ))
$(eval $(call AddProduct, SYMBOLTABLES, symboltables/* ))

.PHONY: all help goenv build test gen clean

all:
	$(MAKE) build V=$(V)

help:
	@$(ECHO) "Usage: $(MAKE) <target> [params]"
	@$(ECHO) "where <target> is one of"
	@$(ECHO) "    help  - print this screen"
	@$(ECHO) "    goenv - print the values of Go's environment variables"
	@$(ECHO) "    build - build this project (default)"
	@$(ECHO) "    test  - run the testsuite on this project"
	@$(ECHO) "    gen   - generate ./pkg/types/types.go"
	@$(ECHO) "    clean - remove built products"
	@$(ECHO) ""
	@$(ECHO) "[params] refers to a list of space separated parameters"
	@$(ECHO) "of the form KEY=VALUE. The so far supported parameters are"
	@$(ECHO) ""
	@$(ECHO) "    V=number    - set the verbocity level; possible values"
	@$(ECHO) "                  of verbocity level are:"
	@$(ECHO) "                    0 - be quite (default);"
	@$(ECHO) "                    1 - be verbose, no logging;"
	@$(ECHO) "                    2 - be verbose, log info messages."
	@$(ECHO) "                  This parameter takes its influence to"
	@$(ECHO) "                  'build' and 'test' targets only;"
	@$(ECHO) "    GOROOT=path - set the Go's GOROOT; the default value is"
	@$(ECHO) "                  taken from '$(GO_ENV) GOROOT';"
	@$(ECHO) "    GOPATH=path - set the Go's GOPATH; the default value is"
	@$(ECHO) "                  the current working directory."
	@$(ECHO) ""

goenv:
	@$(GO_ENV)

build:
	$(GO_BUILD) $(GO_BUILD_FLAGS) -o $(EXTRACT) $(PROJECT_ROOT)/cmd

test:
	#$(GO_TEST) $(GO_TEST_FLAGS) $(PROJECT_ROOT)/pkg/parser $(GLOG_FLAGS)
	$(GO_TEST) $(GO_TEST_FLAGS) $(PROJECT_ROOT)/pkg/parser/file \
            $(GLOG_FLAGS)
	$(GO_TEST) $(GO_TEST_FLAGS) $(PROJECT_ROOT)/pkg/types
	$(GO_TEST) $(GO_TEST_FLAGS) $(PROJECT_ROOT)/pkg/parser/expression \
            $(GLOG_FLAGS)
	$(GO_TEST) $(GO_TEST_FLAGS) $(PROJECT_ROOT)/pkg/parser/statement \
            $(GLOG_FLAGS)

gen:
	./gentypes.sh

clean:
	$(RM) $(RM_FLAGS) $(TRASH)
