# Go Programming Language Specification Notes

Below you can find a notes that can be useful while implementing symbol
extraction and type deduction & propagation.

## Constants

Types of constants:
* boolean
* numeric
  * rune
  * integer
  * floating-point
  * complex
* string

Constant value representation:
* by literal
  * rune
  * integer
  * floating-point
  * imaginary
  * string
* by identifier
* by constant expression
* by conversion
* by built-in functions (TODO(jkucera): list them all)
  * `unsafe.Sizeof`
  * `cap`
  * `len`
  * `real`
  * `imag`
  * `complex`
* by boolean truth values
  * `true`
  * `false`
* by `iota`

Numeric constants:
* exact values
* arbitrary precision

Untyped constants:
* literals
* `true`
* `false`
* `iota`
* constant expressions containing only untyped constant operands

Default type of untyped constant:
* used when there is no explicit type
* `bool` for boolean constant
* `rune` for rune constant
* `int` for integer constant
* `float64` for floating-point constant
* `complex128` for complex constant
* `string` for string constant

Typed constants:
* those with explicit type

Giving a type to constant:
* explicitly
  * by constant declaration
  * by conversion
* implicitly
  * when used in a variable declaration
  * when used in an assignment
  * when used as an operand in an expression

Examples of a giving a type to constant implicitly:
```go
const (
  // Can be assigned to a variable with any integer or floating-point type.
  SmallConst = 3.0
  // Can be assigned to a variable with a float32, float64, or uint32 type.
  HugeConst = 1 << 31
)

var (
  a int = SmallConst       // implicit int
  b float32 = SmallConst   // implicit float32
  c int32 = HugeConst      // error
  d complex128 = HugeConst // implicit complex128
  e float32 = HugeConst    // implicit float32
  f string = HugeConst     // error
)
```

## Variables

Uninitialized variable has the zero value.

Storage reservation:
* by variable declaration
* by function declaration signature
  * if there are function parameters and/or (named) results
* by function literal
  * as in function declaration signature
* by built-in function `new`
* by taking the address of a composite literal

Structured variable is a variable of a type:
* array
* slice
* struct

Elements and fields of structured variables acts like variables.

Static type of a variable is the type:
* given in its declaration
* provided in the `new` call
* provided in the composite literal
* of an element of a structured variable

Dynamic type:
* concrete type of the value assigned to the variable at run time
* variables of interface type have a distinct *dynamic type*
* values stored in interface variables are always assignable to the
  static type of the variable

## Types

`nil` has no type.

Type may be:
* denoted by a *type name*
* specified using a *type literal*
* predeclared
  * `bool byte complex64 complex128 error float32 float64`
  * `int int8 int16 int32 int64 rune string`
  * `uint uint8 uint16 uint32 uint64 uintptr`
* introduced with type declarations
* constructed using type literals (composite types)

Underlying type:
```
underlying-type(T):
    is-predeclared(T) -> T
    is-type-literal(T) -> T
    definition-of(T) is `"type" T "=" U` -> underlying-type(U)
    definition-of(T) is `"type" T U` -> underlying-type(U)
```

Grammar for type:
```
Type      = TypeName | TypeLit | "(" Type ")" .
TypeName  = identifier | QualifiedIdent .
TypeLit   = ArrayType | StructType | PointerType | FunctionType |
            InterfaceType | SliceType | MapType | ChannelType .
```

### Method sets

Method set of a type determines:
* the interfaces that the type implements
* the methods that can be called using a receiver of that type
* each method must have a unique non-blank method name

Method set can be formally defined as
```
method-set(T):
    is-interface-type(T) -> methods-in-interface(T)
    otherwise ->
        { M: M has-receiver T }
        U { M: T has-form *U, M has-receiver U }
```

### Boolean types

Properties:
* set of Boolean truth values
* predeclared type is `bool`
* predeclared constants are `true` and `false`

### Numeric types

Properties:
* sets of integer or floating-point values
* all numeric type are distinct, except aliases
* `byte` is alias for `uint8`
* `rune` is alias for `int32`
* different numeric types cannot be mixed, conversion is needed

Predeclared numeric types:
* architecture-independent
  * two's complement arithmetic
  * `uint8 uint16 uint32 uint64`
  * `int8 int16 int32 int64`
  * `float32 float64`
  * `complex64 complex128`
  * `byte rune`
* implementation-specific sizes
  * `uint int uintptr`

### String types

Properties:
* set of string values
* string value is a sequence of bytes (not runes!)
* strings are immutable
* predeclared type is `string`
* if `s` is a constant string, then `len(s)` is a compile-time constant
* addressing string's elements is illegal (i.e. `&s[i]` is invalid)

### Array types

Properties:
* array is a numbered sequence of elements of a single type
* one-dimensional
* indexable
* have a length
  * must be evaluable as a non-negative `int` constant
  * can be retrieved by `len(a)` built-in function

Grammar for array type:
```
ArrayType   = "[" ArrayLength "]" ElementType .
ArrayLength = Expression .
ElementType = Type .
```

### Slice types

Properties:
* set of all slices of arrays of its element type
* slice
  * is a descriptor for a contiguous segment of an *underlying array*
  * provides access to a numbered sequence of elements from its underlying
    array
* once slice is initialized, its underlying array cannot be changed for another
* one array can be shared by multiple slices
* `nil` is the value of uninitialized slice
* indexable
* have a length
  * may change during execution using *slicing*
* slice index is less than or equal to underlying array index
* slice has a capacity
  * length of the slice + length of the array beyond the slice
  * can be retrieved by `cap(s)`
  * `len(s) <= cap(s)`

Creating a new slice using `make([]T, length, capacity)`:
* `capacity` is optional
* creates new, hidden array

Grammar for slice type:
```
SliceType = "[" ElementType "]" .
```

### Struct types

Properties:
* struct
  * is a sequence of named elements, called fields
  * each field has a name and a type
  * field names may be specified
    * explicitly (IdentifierList)
    * implicitly (EmbeddedField)
  * non-blank field names must be unique

Embedded field:
* a type name `T`
* a pointer to a non-interface type name `*T`, `T` may not be a pointer type
* the unqualified type name acts as the field name

Promoted field or method `f`:
* let `f` be a field or method of an embedded field `F` in a struct `x`
  * `f` is called *promoted* if `x.f` is a legal selector that denotes that
    field or method `f`
* promoted fields act like ordinary fields of a struct
* promoted fields cannot be used as field names in composite literals of the
  struct
* let `S` be a struct type, `T` be a type name, and `M(S)` and `M(*S)` be a
  method set of the struct
  * if `S` contains an embedded field `T`
    * promoted methods with receiver `T` are included in both `M(S)` and
      `M(*S)`
    * promoted methods with receiver `*T` are included in `M(*S)`
  * if `S` contains an embedded field `*T`
    * promoted methods with receiver `T` or `*T` are included in both `M(S)`
      and `M(*S)`

Tag:
* optional string literal that follows a field declaration
* attribute for all the fields in the corresponding field declaration
* empty tag is equivalent to absent tag
* tags
  * are made visible through a reflection interface
  * take part in type identity for structs
  * can be ignored

Example:
```go
struct {
  T             // Embedded type, field name is T.
  *P.TT         // Embedded type, field name is TT.
  _     [4]byte // Anonymous field, padding.
  a, b  int     // Field names are a and b.
}

struct {
  x, y rune        "" // Empty tag.
  z    string      `proto:"str"`
  _    [8] float64 "last padding"
}
```

Grammar for struct type:
```
StructType    = "struct" "{" { FieldDecl ";" } "}" .
FieldDecl     = (IdentifierList Type | EmbeddedField) [ Tag ] .
EmbeddedField = [ "*" ] TypeName .
Tag           = string_lit .
```
