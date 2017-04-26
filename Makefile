test:
	go test -v github.com/gofed/symbols-extractor/pkg/parser
	go test -v github.com/gofed/symbols-extractor/pkg/parser/file
	go test -v github.com/gofed/symbols-extractor/pkg/types
	go test -v github.com/gofed/symbols-extractor/pkg/parser/expression
	go test -v github.com/gofed/symbols-extractor/pkg/parser/statement

gen:
	./gentypes.sh
