test:
	go test -v github.com/gofed/symbols-extractor/pkg/types
	go test -v github.com/gofed/symbols-extractor/pkg/parser

gen:
	./gentypes.sh | gofmt > pkg/types/types.go
