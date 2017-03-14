test:
	go test -v github.com/gofed/symbols-extractor/pkg/types

gen:
	./gentypes.sh | gofmt > pkg/types/types.go
