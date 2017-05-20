all:
	go build  -o extract github.com/gofed/symbols-extractor/cmd

test:
	#go test -v github.com/gofed/symbols-extractor/pkg/parser #-stderrthreshold=INFO
	go test -v github.com/gofed/symbols-extractor/pkg/parser/file #-stderrthreshold=INFO
	go test -v github.com/gofed/symbols-extractor/pkg/types
	go test -v github.com/gofed/symbols-extractor/pkg/parser/expression #-stderrthreshold=INFO
	go test -v github.com/gofed/symbols-extractor/pkg/parser/statement #-stderrthreshold=INFO

gen:
	./gentypes.sh

clean:
	rm extract
	rm -f symboltables/*

