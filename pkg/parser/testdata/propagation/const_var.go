package propagation

/*
  Forms of variable/constant declarations to be considered
  ========================================================

  We combine the following together:
  * one identifier vs list of identifiers
  * missing type vs given type
  * missing value vs one value vs more values
  * single result vs multiple results
  * iota
  * group

  `expr` stands for single-result expression
  `mexpr` stands for multiple-result expression
  `iota` stands for iota variable
  `<x1, x2, ..., xn; m>` denotes an element from the set {x1, x2, ..., xn}^m
    (S^m means an m-th power of S); `!` after `xi` means that `xi` must be
    present in the element

  Constant declarations:

	// BAD! ("missing value in const declaration")
	const X
	// OK
	const X = expr
	// OK
	const X = iota
	// BAD! ("multiple-value in single-value context")
	const X = mexpr
	// BAD! ("extra expression in const declaration")
	const X = <expr, iota, mexpr; 2>

	// BAD! ("const declaration cannot have type without expression" and "missing value in const declaration")
	const X T
	// OK
	const X T = expr
	// OK
	const X T = iota
	// BAD! ("multiple-value in single-value context")
	const X T = mexpr
	// BAD! ("extra expression in const declaration")
	const X T = <expr, iota, mexpr; 2>

	// BAD! ("missing value in const declaration")
	const X, Y
	// BAD! ("missing value in const declaration")
	const X, Y = expr
	// BAD! ("missing value in const declaration")
	const X, Y = iota
	// OK
	const X, Y = mexpr
	// OK
	const X, Y = <expr, iota; 2>
	// BAD! ("multiple-value in single-value context")
	const X, Y = <mexpr!, expr, iota; 2>
	// BAD! ("extra expression in const declaration")
	const X, Y = <mexpr, expr, iota; 3>
	// BAD! ("missing value in const declaration" or "multiple-value in single-value context")
	const X, Y, Z = <mexpr, expr, iota; 2>

	// BAD! ("const declaration cannot have type without expression" and "missing value in const declaration")
	const X, Y T
	// BAD! ("missing value in const declaration")
	const X, Y T = expr
	// BAD! ("missing value in const declaration")
	const X, Y T = iota
	// OK
	const X, Y T = mexpr
	// OK
	const X, Y T = <expr, iota; 2>
	// BAD! ("multiple-value in single-value context")
	const X, Y T = <mexpr!, expr, iota; 2>
	// BAD! ("extra expression in const declaration")
	const X, Y T = <mexpr, expr, iota; 3>
	// BAD! ("missing value in const declaration" or "multiple-value in single-value context")
	const X, Y, Z T = <mexpr, expr, iota; 2>

	// BAD! ("missing value in const declaration")
	const (
		X
	)

	// OK
	const (
		A = expr
		B
		C = iota
		D
	)

	// BAD! ("missing value in const declaration")
	const (
		X = expr
		Y, Z
	)

	// BAD! ("multiple-value mexpr in single-value context")
	const (
		X = mexpr
	)

	// BAD! ("extra expression in const declaration")
	const (
		X = <mexpr, expr, iota; 2>
	)

	// BAD! ("const declaration cannot have type without expression" and "missing value in const declaration")
	const (
		X T
	)

	// BAD! ("const declaration cannot have type without expression")
	const (
		X T = expr
		Y T
	)

	// OK
	const (
		A T = iota
		B
	)

	// BAD! ("const declaration cannot have type without expression" and "missing value in const declaration")
	const (
		X T = expr
		Y, Z T
	)

	// BAD! ("multiple-value mexpr in single-value context")
	const (
		X T = mexpr
	)

	// BAD! ("extra expression in const declaration")
	const (
		X T = <mexpr, expr, iota; 2>
	)

	// BAD! ("missing value in const declaration")
	const (
		X, Y
	)

	// BAD! ("missing value in const declaration")
	const (
		X, Y = expr
	)

	// BAD! ("missing value in const declaration")
	const (
		X, Y = iota
	)

	// BAD! ("extra expression in const declaration")
	const (
		A, B = mexpr
		C
	)

	// OK
	const (
		A, B = mexpr
		C, D
		E, F = <expr, iota; 2>
		G, H
	)

	// BAD! ("multiple-value in single-value context")
	const (
		X, Y = <mexpr!, iota, expr; 2>
	)

	// BAD! ("extra expression in const declaration")
	const (
		X, Y = <mexpr, expr, iota; 3>
	)

	// BAD! ("missing value in const declaration" or "multiple-value in single-value context")
	const (
		X, Y, Z = <mexpr, expr, iota; 2>
	)

	// BAD! ("const declaration cannot have type without expression" and "missing value in const declaration")
	const (
		X, Y T
	)

	// BAD! ("missing value in const declaration")
	const (
		X, Y T = expr
	)

	// BAD! ("missing value in const declaration")
	const (
		X, Y T = iota
	)

	// BAD! ("extra expression in const declaration")
	const (
		A, B T = mexpr
		C
	)

	// OK
	const (
		A, B T = mexpr
		C, D
		E, F T = <expr, iota; 2>
		G, H
	)

	// BAD! ("const declaration cannot have type without expression")
	const (
		X, Y T = mexpr
		W, Z T
	)

	// BAD! ("multiple-value in single-value context")
	const (
		X, Y T = <mexpr!, iota, expr; 2>
	)

	// BAD! ("extra expression in const declaration")
	const (
		X, Y T = <mexpr, expr, iota; 3>
	)

	// BAD! ("missing value in const declaration" or "multiple-value in single-value context")
	const (
		X, Y, Z T = <mexpr, expr, iota; 2>
	)

  Variable declarations:
  - are made in similar manner, but `iota` is undefined

*/

func F() (int, int) {return 1, 2}

// ============================================================================
// == Const Declarations
// ============================================================================

// ----------------------------------------------------------------------------
// -- const X
//const X // missing constant value

// ----------------------------------------------------------------------------
// -- const X = <expr, iota; 1>
// Boolean untyped
const B1 = true
const B2 = false
// Rune untyped
const R1 = '0'
const R2 = 'ř'
// Integer untyped
const I1 = 1
const I2 = iota
// Float untyped
const F1 = 0.0
const F2 = 1.5
// Complex untyped
const C1 = 0i
const C2 = 1i
const C3 = 0.5i
// String untyped
const S1 = ""
const S2 = "a"
const S3 = "abc"

// ----------------------------------------------------------------------------
// -- const X = mexpr
// ValueSpec `X1 = F()` has different number of identifiers on LHS (1) than a
// number of results of invocation on RHS (2)
//const X1 = F()
