#!/bin/sh

python gen-types.py
gofmt -w pkg/types/types.go
gofmt -w pkg/parser/symboltable/symbol.go
