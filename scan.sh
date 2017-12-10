#!/bin/bash

# pushd /usr/lib/golang/src 1>/dev/null 2>/dev/null && gofed inspect -p

error=0
mkdir -p symboltables
for package in $(cat gopackages); do
	if [ "${package:0:1}" == "#" ]; then
		echo "Skipping $package"
		continue
	fi	
	#echo "./extract --package-path=$package --symbol-table-dir symboltables --cgo-symbols-path cgo/cgo.yml 1>/tmp/log.log 2>&1"
	ok=$(./extract --package-path=$package --symbol-table-dir symboltables --cgo-symbols-path cgo/cgo.yml 1>/dev/null 2>&1 && echo "OK")
	if [ "$ok" == "OK" ]; then
		echo "* [X] $package"
	else
		echo "* [ ] $package"
		error=1
		#exit 1
	fi
	rm -f /tmp/extract.*
	#echo "ST count: $(ls symbotables/ | wc -l)"
done

if [ $error -eq 1 ]; then
	exit 1
fi
