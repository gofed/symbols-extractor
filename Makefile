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
# - bump up GOROOT:
ifeq (not$(GOROOT)set, notset)
GOROOT := $(shell $(GO_ENV) GOROOT)
export GOROOT
endif
# - set GOPATH to be the current working directory; GOPATH must be absolute:
ifeq (not$(GOPATH)set, notset)
GOPATH := $(CURDIR)
export GOPATH
endif

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
	@$(ECHO) "of the form KEY=VALUE. The list of so far supported"
	@$(ECHO) "parameters is"
	@$(ECHO) ""
	@$(ECHO) "    V=n   - set the verbocity level; possible values of"
	@$(ECHO) "            verbocity level ('n') are:"
	@$(ECHO) "              0 - be quite (default);"
	@$(ECHO) "              1 - be verbose, no logging;"
	@$(ECHO) "              2 - be verbose, log info messages."
	@$(ECHO) "            This parameter takes its influence to 'build'"
	@$(ECHO) "            and 'test' targets only."
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
