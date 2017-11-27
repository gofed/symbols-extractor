#!/bin/bash

# pushd /usr/lib/golang/src 1>/dev/null 2>/dev/null && gofed inspect -p

for package in $(cat gopackages); do
	#echo "./extract --package-path=$package --symbol-table-dir symboltables --cgo-symbols-path cgo/cgo.yml 1>/tmp/log.log 2>&1"
	ok=$(./extract --package-path=$package --symbol-table-dir symboltables --cgo-symbols-path cgo/cgo.yml 1>/tmp/log.log 2>&1 && echo "OK")
	if [ "$ok" == "OK" ]; then
		echo "* [X] $package"
	else
		echo "* [ ] $package"
		#exit 1
	fi
	#echo "ST count: $(ls symbotables/ | wc -l)"
done
