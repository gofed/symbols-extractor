#!/bin/sh

python gen-types.py | gofmt > pkg/types/types.go
