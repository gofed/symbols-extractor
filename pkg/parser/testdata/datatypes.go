package testdata

type Struct struct {
	// Simple ID
	simpleID uint64
	// Pointer
	simplePointerID *string
	// Channel
	simpleChannel chan struct{}
	// Struct
	simpleStruct struct {
		simpleID uint64
	}
	// Map
	simpleMap [string]struct{}
	// Array
	simpleSlice []string
	simpleArray [4]string
	// Function
	simpleMethod func(arg1, arg2 string) (string, error)
	// interface
	simpleInterface interface {
		simpleMethod(arg1, arg2 string) (string, error)
	}
	// Ellipsis
	simpleMethodWithEllipsis func(arg1 string, ellipsis ...string) (string, error)
	//
}
