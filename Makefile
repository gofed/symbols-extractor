#                                                         -*- coding: utf-8 -*-
# File:    ./Makefile
# Author:  Jan Chaloupka, <jchaloup AT redhat DOT com>
#          Jiří Kučera, <jkucera AT redhat DOT com>
# Stamp:   N/A
# Project: Symbols Extractor
# Version: See VERSION.
# License: See LICENSE.
# Brief:   Makefile.
#

# Commands & tools:
GO_CMD           := go
GO_ENV           := $(GO_CMD) env
GO_BUILD         := $(GO_CMD) build
GO_BUILD_FLAGS   :=
GO_INSTALL       := $(GO_CMD) install
GO_INSTALL_FLAGS :=
GO_TEST          := $(GO_CMD) test
GO_TEST_FLAGS    :=
GLOG_FLAGS       :=
ECHO             := echo
PASS             := true
RM               := rm
RM_FLAGS         := -f

# Params:
# - verbocity level:
ifeq (not$(V)set, notset)
V := 0
endif
V_AUX := $(V)
ifeq ($(V_AUX), 2)
GLOG_FLAGS += -stderrthreshold=INFO
V_AUX := 1
endif
ifeq ($(V_AUX), 1)
GO_BUILD_FLAGS += -v
GO_INSTALL_FLAGS += -v
GO_TEST_FLAGS  += -v
V_AUX := 0
endif
# - logging:
ifeq (not$(L)set, notset)
L := 0
endif
ifeq ($(L), 0)
WLOG :=
RMLOG := $(PASS)
else
WLOG := >>$(MAKECMDGOALS).log 2>&1
RMLOG := $(RM) $(RM_FLAGS) $(MAKECMDGOALS).log
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
define AddProductX =
$$(eval $$(call _AddProductX,$$(strip $(1)),$$(strip $(2))))
endef
define AddProductD =
$$(eval $$(call _AddProductD,$$(strip $(1)),$$(strip $(2))))
endef
define _AddProductX =
$(1) := $(2)
TRASH += $$(GOPATH)/bin/$$($(1))
TRASH += $$(GOPATH)/bin/$$($(1)).exe
endef
define _AddProductD =
$(1) := $(2)
TRASH += $$($(1))/*
endef

$(eval $(call AddProductX, UPDATE,       update       ))
$(eval $(call AddProductX, EXTRACT,      extract      ))
$(eval $(call AddProductD, SYMBOLTABLES, symboltables ))

# Remove old log:
$(shell $(RMLOG))

# Dependencies:
UPDATE_DEPS  := ./cmd/update/update.go \
                ./pkg/updater/testgen/init.go \
                ./pkg/updater/command.go \
                ./pkg/updater/errors.go \
                ./pkg/updater/help.go \
                ./Makefile
EXTRACT_DEPS := ./cmd/extract/extract.go \
                ./pkg/parser/alloctable/global/global.go \
                ./pkg/parser/alloctable/table.go \
                ./pkg/parser/expression/expression_parser.go \
                ./pkg/parser/file/file_parser.go \
                ./pkg/parser/statement/statement_parser.go \
                ./pkg/parser/symboltable/global/global.go \
                ./pkg/parser/symboltable/stack/stack.go \
                ./pkg/parser/symboltable/Ctable.go \
                ./pkg/parser/symboltable/table.go \
                ./pkg/parser/symboltable/types.go \
                ./pkg/parser/type/type_parser.go \
                ./pkg/parser/types/types.go \
                ./pkg/parser/parser.go \
                ./pkg/types/internal_types.go \
                ./pkg/types/types.go \
                ./Makefile

.PHONY: all help goenv install test gen extract clean

all:
	$(MAKE) install V=$(V) L=$(L)

help:
	@$(ECHO) "Usage: $(MAKE) <target> [params]"
	@$(ECHO) "where <target> is one of"
	@$(ECHO) "    help    - print this screen"
	@$(ECHO) "    goenv   - print the values of Go's environment variables"
	@$(ECHO) "    install - build this project and install binaries to"
	@$(ECHO) "              $(GOPATH)/bin (default)"
	@$(ECHO) "    test    - run the testsuite on this project"
	@$(ECHO) "    gen     - generate ./pkg/types/types.go"
	@$(ECHO) "    extract - run Symbols Extractor"
	@$(ECHO) "    clean   - remove built products"
	@$(ECHO) ""
	@$(ECHO) "[params] refers to a list of space separated parameters"
	@$(ECHO) "of the form KEY=VALUE. The so far supported parameters are"
	@$(ECHO) ""
	@$(ECHO) "    V=number    - set the verbocity level; possible values"
	@$(ECHO) "                  of verbocity level are:"
	@$(ECHO) "                    0 - be quite (default)"
	@$(ECHO) "                    1 - be verbose, no logging"
	@$(ECHO) "                    2 - be verbose, log info messages"
	@$(ECHO) "    L=bool      - if L differs from 0 or empty value, stdout"
	@$(ECHO) "                  and stderr are redirected to the log file"
	@$(ECHO) "                  named '<target>.log'"
	@$(ECHO) "    PKG=path    - path to a package to be extracted"
	@$(ECHO) "                  (relative to GOPATH)"
	@$(ECHO) "    GOROOT=path - set the Go's GOROOT; the default value is"
	@$(ECHO) "                  taken from '$(GO_ENV) GOROOT'"
	@$(ECHO) "    GOPATH=path - set the Go's GOPATH; the default value is"
	@$(ECHO) "                  the current working directory"
	@$(ECHO) ""

goenv:
	@$(GO_ENV) $(WLOG)

install: $(GOPATH)/bin/$(UPDATE) $(GOPATH)/bin/$(EXTRACT)

test:
#	$(GO_TEST) $(GO_TEST_FLAGS) $(PROJECT_ROOT)/pkg/parser $(GLOG_FLAGS) \
#            $(WLOG)
	$(GO_TEST) $(GO_TEST_FLAGS) $(PROJECT_ROOT)/pkg/parser/file \
            $(GLOG_FLAGS) $(WLOG)
	$(GO_TEST) $(GO_TEST_FLAGS) $(PROJECT_ROOT)/pkg/types $(WLOG)
	$(GO_TEST) $(GO_TEST_FLAGS) $(PROJECT_ROOT)/pkg/parser/expression \
            $(GLOG_FLAGS) $(WLOG)
	$(GO_TEST) $(GO_TEST_FLAGS) $(PROJECT_ROOT)/pkg/parser/statement \
            $(GLOG_FLAGS) $(WLOG)

gen:
	./gentypes.sh

ifeq (not$(PKG)set, notset)
extract:
	@$(ECHO) "PKG is not set"
else
extract: $(GOPATH)/bin/$(EXTRACT)
	$(GOPATH)/bin/$(EXTRACT) \
            --package-path $(GOPATH)/$(PKG) \
            --symbol-table-dir $(SYMBOLTABLES) \
            --cgo-symbols-path cgo/cgo.yml \
            $(GLOG_FLAGS) $(WLOG)
endif

clean:
	$(RM) $(RM_FLAGS) $(TRASH)

$(GOPATH)/bin/$(UPDATE): $(UPDATE_DEPS)
	$(GO_INSTALL) $(GO_INSTALL_FLAGS) $(PROJECT_ROOT)/cmd/$(UPDATE) $(WLOG)

$(GOPATH)/bin/$(EXTRACT): $(EXTRACT_DEPS)
	$(GO_INSTALL) $(GO_INSTALL_FLAGS) $(PROJECT_ROOT)/cmd/$(EXTRACT) \
            $(WLOG)
