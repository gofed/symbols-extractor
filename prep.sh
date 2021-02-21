#!/bin/sh

# copy-paste all internal bits from go and change the import path prefixes
# so cmd/go/internal/load.PackagesAndErrors can be invoked
# TODO(jchaloup): find a way how to extract package info without importing go's internal code
rm -rf cmd/go/internal
mkdir -p cmd/go
cp -r $GOROOT/src/cmd/go/internal/ cmd/go/internal
rm -rf cmd/internal
cp -r $GOROOT/src/cmd/internal/ cmd/internal
rm -rf internal
cp -r $GOROOT/src/internal/ internal

for file in $(find cmd/go -iname "*.go"); do
  sed -i "s/\"cmd\/go\/internal\//\"github\.com\/gofed\/symbols-extractor\/cmd\/go\/internal\//g" $file
  sed -i "s/\"cmd\/internal\//\"github\.com\/gofed\/symbols-extractor\/cmd\/internal\//g" $file
  sed -i "s/\"internal\//\"github\.com\/gofed\/symbols-extractor\/internal\//g" $file
done

for file in $(find cmd/internal -iname "*.go"); do
  sed -i "s/\"cmd\/go\/internal\//\"github\.com\/gofed\/symbols-extractor\/cmd\/go\/internal\//g" $file
  sed -i "s/\"cmd\/internal\//\"github\.com\/gofed\/symbols-extractor\/cmd\/internal\//g" $file
  sed -i "s/\"internal\//\"github\.com\/gofed\/symbols-extractor\/internal\//g" $file
done

rm -rf vendor/golang.org/x
mkdir -p vendor/golang.org
cp -r $GOROOT/src/cmd/vendor/golang.org/x vendor/golang.org/x
