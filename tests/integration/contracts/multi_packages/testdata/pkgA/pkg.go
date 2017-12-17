package pkgA

// definitions of data types and methods to be imported in other package(s)

type A struct {
	f int
}

func (a *A) method(a, b int) string {
	return ""
}

type B []A

type C map[int]B
