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
* by built-in functions
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

Implementation restriction: every implementation must (apply both to literal
constants and to the result of evaluating constant expressions):
* represent integer constants with at least 256 bits
* represent floating-point constants, including the parts of complex constants
  * with a mantissa of at least 256 bits
  * with a signed binary exponent of at least 16 bits
* give an error if unable to represent an integer constant precisely
* give an error if unable to represent a floating-point or complex constant due
  to overflow
* round to the nearest representable constant if unable to represent
  floating-point or complex constant due to limits on precision

Examples of a giving a type to constant implicitly:
```go
const (
	// Can be assigned to a variable with any integer or floating-point
	// type.
	SmallConst = 3.0
	// Can be assigned to a variable with a float32, float64, or uint32
	// type.
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

Boolean type:
* set of Boolean truth values
* predeclared type is `bool`
* predeclared constants are `true` and `false`

### Numeric types

Numeric type:
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

String type:
* set of string values
* string value is a sequence of bytes (not runes!)
* strings are immutable
* predeclared type is `string`
* if `s` is a constant string, then `len(s)` is a compile-time constant
* addressing string's elements is illegal (i.e. `&s[i]` is invalid)

### Array types

Array type:
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

Slice type:
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

Struct type:
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

Examples:
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

### Pointer types

Pointer type:
* set of all pointers to variables of a given type, the *base type*
* uninitialized pointer's value is `nil`

Grammar for pointer type:
```
PointerType = "*" BaseType .
BaseType    = Type .
```

### Function types

Function type:
* set of all functions with the same parameter and result types
* the value of uninitialized variable of function type is `nil`
* names of results must either all be present or all be absent
* all non-blank names in the signature must be unique
* *variadic* parameter `...`
  * only final incoming parameter in a function signature can be variadic
  * stands for zero or more parameters

Grammar for function type:
```
FunctionType = "func" Signature .
Signature = Parameters [ Result ] .
Result = Parameters | Type .
Parameters = "(" [ ParameterList [ "," ] ] ")" .
ParameterList = ParameterDecl { "," ParameterDecl } .
ParameterDecl = [ IdentifierList ] [ "..." ] Type .
```

### Interface types

Interface type:
* specifies a method set called its *interface*
* each method in an interface type must have a unique non-blank name
* the value of an uninitialized variable of interface type is `nil`
* let `I` be an interface and let `T` be a type such that `I` is subset of
  `M(T)`
  * for `var i I` and `var t T`, `i` can store a value of `t`
  * `T` implements `I`
* more than one type can implement an interface
* one type can implement several distinct interfaces

Embedding:
* let `T` and `E` be interfaces
  * if `E` is embedded in `T`, then `E` adds all its methods to `T`
* recursion (cyclic embedding) is not allowed

Grammar for an interface type:
```
InterfaceType     = "interface" "{" { MethodSpec ";" } "}" .
MethodSpec        = MethodName Signature | InterfaceTypeName .
MethodName        = identifier .
InterfaceTypeName = TypeName .
```

### Map types

Map type:
* map is an unordered group of elements of one type indexed by a set of unique
  keys of another type
* map has dynamic capacity
* uninitialized map has `nil` value
* `nil` maps are immutable
* key type must support the comparison operators
  * for an interface type, comparison operators must be defined for the dynamic
    key values
* `len(m)` returns a number of elements of map `m`
* assignment adds an element
* indexing retrieves an element
* `delete(m, k)` removes element at index `k`
* `make(map[T]U, capacity)`
  * creates a new empty map of a given type
  * `capacity` is optional

Grammar for a map type:
```
MapType = "map" "[" KeyType "]" ElementType .
KeyType = Type .
```

### Channel types

Channel type:
* uninitialized channel has the `nil` value
* a `nil` channel is never ready for communication
* channel acts as FIFO queue
* kinds of channels
  * sending `chan<-`
  * receiving `<-chan`
  * bidirectional `chan`
* channel may be constrained only to send/receive by conversion/assignment
* single channel may be used by any number of goroutines without further
  synchronization
  * in send statement
  * in receive operations
  * in calls to `len(ch)` and `cap(ch)`
* `<-` associates with the leftmost `chan` possible
* `make(chan T, capacity)`
  * creates new, initialized channel value
  * `capacity` is optional
* `close(ch)`
  * close the channel
  * `v, ok = <-ch` reports whether a received value was sent before the channel
    was closed

Buffering:
* capacity of a channel is a number of its elements, the size of its buffer
* unbuffered channel has zero capacity
  * communication succeed only when both a sender and receiver are ready
* buffered channel
  * communication succeeds without blocking
    * if the buffer is not full (sends)
    * if the buffer is not empty (receives)

Grammar for channel type:
```
ChannelType = ( "chan" | "chan" "<-" | "<-" "chan" ) ElementType .
```

## Properties of types and values

### Type identity

A defined type is always different from any other type.

Otherwise, two types are identical if their underlying type literals are
structurally equivalent:
* two array types are identical if they have
  * identical element types
  * same array length
* two slice types are identical if they have
  * identical element types
* two struct types are identical if they have
  * the same sequence of fields
    * two corresponding fields are have
      * the same names
      * identical types
      * identical tags
  * non-exported field names from different packages are always different
* two pointer types are identical if they have
  * identical base types
* two function types are identical if they have
  * the same number of parameters
    * two corresponding parameters must have identical types
  * the same number of result values
    * two corresponding result values must have identical types
  * either both functions are variadic or neither is
* two interface types are identical if they have
  * the same set of methods with
    * the same names
    * identical function types
  * non-exported method names from different packages are always different
* two map types are identical if they have
  * identical key types
  * identical value types
* two channel types are identical if they have
  * identical value types
  * same direction

### Assignability

A value `x` of type `V` is *assignable* to a variable of type `T` in any of
these cases:
* `V` is identical to `T`
* `V` and `T` have identical underlying types
  * at least one of `V` or `T` is not a defined type
* `T` is an interface type
  * `x` implements `T`
* `T` is a channel type
  * `x` is a bidirectional channel value
  * `V` and `T` have identical element types
  * at least one of `V` or `T` is not a defined type
* `x` is a predeclared identifier `nil`
  * `T` is a pointer, function, slice, map, channel, or interface type
* `x` is an untyped constant representable by a value of type `T`

## Blocks

Explicit or implicit.

Implicit blocks:
* the *universe block* encompasses all Go sources
* a *package block* containing all Go source text for the package
* a *file block* containing all Go source text in the file
* are around each `if`, `for`, and `switch`
* clauses in `switch` or `select` acts as implicit blocks

Grammar for blocks:
```
Block         = "{" StatementList "}" .
StatementList = { Statement ";" } .
```

## Declarations and scope

Declaration:
* binds a non-blank identifier to a
  * constant
  * type
  * variable
  * function
  * label
  * package
* identifier
  * in a program, every identifier must be declared
  * in the same block, no identifier may be declared twice
  * no identifier may be declared in both the file and package block
* the blank identifier `_`
  * may be used like any other identifier in a declaration
  * it does not introduce a binding and thus is not declared
* the identifier `init`
  * may only be used for `init` function declarations
  * it does not introduce a new binding

Scope in general:
* the scope of predeclared identifier is the universe block
* the scope of an identifier denoting a constant, type, variable, or function
  (but not method) declared at top level is the package block
* the scope of the package name of an imported package is the file block of the
  file containing the import declaration
* the scope of an identifier denoting a method receiver, function parameter, or
  result variable is the function body
* the scope of a constant or variable identifier declared inside a function
  begins at end of the ConstSpec or VarSpec (or ShortVarDecl) and ends at the
  end of the innermost containing block
* the scope of a type identifier declared inside a function begins at the
  identifier in the TypeSpec and ends at the end of the innermost containing
  block
* identifier may be redeclared in inner block and its declaration last until
  the end of this block

Package clause:
* is not declaration
* package name does not appear in any scope
* identifies the files belonging to the same package
* specifies the default package name for import declarations

Grammar for declarations:
```
Declaration  = ConstDecl | TypeDecl | VarDecl .
TopLevelDecl = Declaration | FunctionDecl | MethodDecl .
```

### Label scopes

Label:
* declared by label statement
* used in `break`, `continue`, and `goto`
* when declared, it must be used
* not block scoped
* has its own namespace (do not conflict with same named non-label identifiers)
* scope is the body of the function in which it is declared
  * excludes the body of any nested function

### Blank identifier

Blank identifier:
* represented by `_`
* serves as an anonymous placeholder
* has special meaning in
  * declarations
  * as an operand
  * in assignments

### Predeclared identifies

Declared implicitly in the universe block:
* types
  * `bool byte complex64 complex128 error float32 float64`
  * `int int8 int16 int32 int64 rune string`
  * `uint uint8 uint16 uint32 uint64 uintptr`
* constants
  * `true false iota`
* zero value
  * `nil`
* functions
  * `append cap close complex copy delete imag len`
  * `make new panic print println real recover`

### Exported identifiers

Permit access to it from another package.

An identifier is exported if both:
* the 1st character of its name is a Unicode upper case letter
* it is declared in the package block or it is a field name or method name

All other identifiers are not exported.

### Uniqueness of identifiers

Given a set of identifiers. An identifier is called *unique* if it is
*different* from every other in the set.

Two identifiers are *different* if they have different names, or if they appear
in different packages and are not exported.

Otherwise, they are the same.

### Constant declarations

Constant declaration:
* binds a list of identifiers to the values of a list of constant expressions
  * if the type is present
    * all constants take the type specified
    * expressions must be assignable to that type
  * if the type is omitted
    * the constants take the individual types of the corresponding expressions
      * if the expression values are untyped constants
        * the declared constants remain untyped
        * the constant identifiers denote the constant values

Omitting the expression list:
* in parenthesized `const` declaration list
* from any but the first declaration
  * such an empty list is equivalent to the textual substitution of the first
    preceding non-empty expression list and its type if any

Grammar for constant declaration:
```
ConstDecl      = "const" ( ConstSpec | "(" { ConstSpec ";" } ")" ) .
ConstSpec      = IdentifierList [ [ Type ] "=" ExpressionList ] .

IdentifierList = identifier { "," identifier } .
ExpressionList = Expression { "," Expression } .
```

### Iota

Iota:
* within a constant declaration
  * represents successive untyped integer constants
  * reset to 0 whenever the reserved word `const` appears in the source
  * increments after each ConstSpec
* within an ExpressionList
  * the value of each `iota` is the same

### Type declarations

Type declaration:
* binds an identifier, the *type name*, to a type
* come in two forms
  * alias declarations
  * type definitions

Grammar for type declaration:
```
TypeDecl = "type" ( TypeSpec | "(" { TypeSpec ";" } ")" ) .
TypeSpec = AliasDecl | TypeDef .
```

#### Alias declarations

Alias declaration:
* binds an identifier to the given type
* within the scope of the identifier, it serves as an *alias* for the type

Grammar for alias declaration:
```
AliasDecl = identifier "=" Type .
```

#### Type definitions

Type definition:
* creates new, distinct type with the same underlying type and operations as
  the given type, and binds an identifier to it
  * the new type is called *defined type*

Defined type:
* may have methods associated with it
* does not inherit any methods bound to the given type, but
  * the method set of an interface type remains unchanged
  * the method set of elements of a composite type remains unchanged

```go
// Suppose A0 has non-empty method set and it is not an interface type.

// The method set of A1 is empty.
type A1 A0

// The method set of the base type of A2 remains unchanged, but the method set
// of A2 is empty.
type A2 *A0

// The method set of *A3 contains the methods bound to its embedded field A0.
type A3 struct {
	A0
}

// Suppose that I0 is an interface type. I1 has the same method set as I0.
type I1 I0
```

Grammar for type definition:
```
TypeDef = identifier Type .
```

### Variable declarations

Variable declaration:
* creates one or more variables
* binds corresponding identifiers to them
* gives each a type and an initial value
* if a list of expressions is given
  * variables are initialized with the expressions following the rules for
    assignments
  * otherwise, each variable is initialized to its zero value
* if a type is present
  * each variable is given that type
  * otherwise, each variable is given the type of the corresponding
    initialization value in the assignment
    * if that value is an untyped constant
      * it is first converted to its default type
    * if it is an untyped boolean value
      * it is first converted to type `bool`
* `nil` cannot be used to initialize a variable with no explicit type

```go
var d = math.Sin(0.5)  // d is float64
var i = 42             // i is int
var t, ok = x.(T)      // t is T, ok is bool
var n = nil            // illegal
```

Implementation restriction:
* a compiler may make it illegal to declare a variable inside a function body
  if the variable is never used

Grammar for variable declaration:
```
VarDecl = "var" ( VarSpec | "(" { VarSpec ";" } ")" ) .
VarSpec = IdentifierList ( Type [ "=" ExpressionList ] | "=" ExpressionList ) .
```

### Short variable declarations

Short variable declaration:
* may *redeclare* variables provided they were originally declared earlier in
  the same block/parameter lists with the same type, and at least one of the
  non-blank variables is new
  * redeclaration does not introduce a new variable; it just assigns a new
    value to the original
* may appear only inside functions

Grammar for short variable declaration:
```
ShortVarDecl = IdentifierList ":=" ExpressionList .
```

### Function declarations

Function declaration:
* binds an identifier, the *function name*, to a function
* if the function's signature declares result parameters, the function body's
  statement list must end in a terminating statement
* may omit the body
  * such a declaration provides the signature for a function implemented
    outside Go

Grammar for function declaration:
```
FunctionDecl = "func" FunctionName ( Function | Signature ) .
FunctionName = identifier .
Function     = Signature FunctionBody .
FunctionBody = Block .
```

### Method declarations

Method declaration:
* a method is a function with *receiver*
* binds an identifier, the *method name*, to a method, and associates the
  method with the receiver's *base type*

Receiver:
* must be a single non-variadic parameter declared in an extra parameter
  section preceding the method name
* its type must be of the form `T` or `*T`
  * `T` is a type name
  * `T` is called the receiver *base type*
  * `T` must not be a pointer or interface type
  * `T` must be defined in the same package as the method
* receiver identifier
  * non-blank receiver identifier must be unique in the method signature
  * it may be omitted in the declaration if it is not used inside method's body

Base type:
* non-blank names of methods bound to the base type must be unique
* in case of a struct type, the non-blank method and field names must be
  distinct

Method:
* let `T` be a type that satisfies the restrictions from the Receiver paragraph
  above; then
  * the method is said to be *bound* to the base type `T`
  * the method name is visible only within selectors for type `T` or `*T`
* the type of a method is the type of a function with the receiver as first
  argument

Grammar for method declaration:
```
MethodDecl = "func" Receiver MethodName ( Function | Signature ) .
Receiver   = Parameters .
```
