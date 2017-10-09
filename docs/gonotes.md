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
FunctionType  = "func" Signature .
Signature     = Parameters [ Result ] .
Result        = Parameters | Type .
Parameters    = "(" [ ParameterList [ "," ] ] ")" .
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

## Expressions

### Operands

Grammar for operand:
```
Operand     = Literal | OperandName | MethodExpr | "(" Expression ")" .
Literal     = BasicLit | CompositeLit | FunctionLit .
BasicLit    = int_lit | float_lit | imaginary_lit | rune_lit | string_lit .
OperandName = identifier | QualifiedIdent .
```

Notes:
* the blank identifier may appear as an operand only on the left-hand side of
  an assignment.

### Qualified identifiers

Qualified identifier:
* accesses an identifier in a different package, which must be imported
  * identifier must be exported and declared in the package block of that
    package

Grammar for qualified identifier:
```
QualifiedIdent = PackageName "." identifier .
```

Notes:
* both package name and the identifier must not be blank

### Composite literals

Composite literal:
* creates a new value each time it is evaluated
* taking the address of a composite literal generates a pointer to a unique
  variable initialized with the literal's value

Grammar for composite literal:
```
CompositeLit = LiteralType LiteralValue .
LiteralType  = StructType | ArrayType | "[" "..." "]" ElementType |
               SliceType | MapType | TypeName .
LiteralValue = "{" [ ElementList [ "," ] ] "}" .
ElementList  = KeyedElement { "," KeyedElement } .
KeyedElement = [ Key ":" ] Element .
Key          = FieldName | Expression | LiteralValue .
FieldName    = identifier .
Element      = Expression | LiteralValue .
```

Notes:
* the LiteralType's underlying type must be a
  * struct, array, slice, or map type
* the types of the elements and keys must be assignable to the respective
  field, element, and key types of the literal type
  * there is no additional conversion
* the key is interpreted as
  * a field name for struct literals
  * an index for array and slice literals
  * a key for map literals
* map literals
  * all elements must have a key
* multiple key elements with the same field name or constant key value are
  forbidden
* struct literals
  * a key must be a field name declared in the struct type
  * an element list that does not contain any keys must list an element for
    each struct field in the order in which the fields are declared
  * if any element has a key, every element must have a key
  * an element list that contains keys does not need to have an element for
    each struct field
    * omitted fields get the zero value for that field
  * a literal may omit the element list
    * such a literal evaluates to the zero value for its type
  * specifying an element for a non-exported field of a struct belonging to a
    different package is forbidden
* array literals
  * the length of an array literal is the length specified in the literal type
    * if fewer elements than the length are provided in the literal, the
      missing elements are set to the zero value for the array element type
  * providing elements with index values outside the index range of the array
    is forbidden
  * `...` specifies an array length equal to the maximum element index plus 1
* slice literal
  * describes the entire underlying array literal
* array and slice literals
  * each element has an associated integer index marking its position in the
    array
  * an element with a key uses the key as its index
    * the key must be a non-negative constant representable by a value of type
      `int`
      * if it is typed it must be of integer type
  * an element without a key uses the previous element's index plus one
    * if the first element has no key, its index is zero
* within a composite literal of array, slice, or map type `T`
  * elements or map keys that are themselves composite literals
    * may elide the respective literal type if it is identical to the element
      or key type of `T`
  * elements or keys that are addresses of composite literals may elide the
    `&T` when the element or key type is `*T`

### Function literals

Function literal:
* represents anonymous function
* can be assigned to a variable or invoked directly
* is a *closure*
  * may refer to variables defined in a surrounding function
    * those variables are shared between the surrounding function and the
      function literal
    * they survive as long as they are accessible

Grammar for function literal:
```
FunctionLit = "func" Function .
```

### Primary expressions

Grammar for primary expression:
```
PrimaryExpr =
	Operand |
	Conversion |
	PrimaryExpr Selector |
	PrimaryExpr Index |
	PrimaryExpr Slice |
	PrimaryExpr TypeAssertion |
	PrimaryExpr Arguments .

Selector      = "." identifier .
Index         = "[" Expression "]" .
Slice         = "[" [ Expression ] ":" [ Expression ] "]" |
                "[" [ Expression ] ":" Expression ":" Expression "]" .
TypeAssertion = "." "(" Type ")" .
Arguments     = "(" [
	( ExpressionList | Type [ "," ExpressionList ] ) [ "..." ] [ "," ]
] ")" .
```

### Selectors

Let `x` be a primary expression that is not a package name. The
*selector expression* `x.f`:
* denotes the field or method `f` of the value `x` (or `*x` sometimes)
* `f` is called the *selector*
* `f` must not be the blank identifier
* the type of `x.f` is the type of `f`
* a selector `f` may denote a field or method `f` of a type `T`, or it may
  refer to a field or method `f` of a nested embedded field of `T`
  * the number of embedded fields traversed to reach `f` is called its *depth*
    in `T`

Rules:
* for a value `x` of type `T` or `*T` where `T` is not a pointer or interface
  type
  * `x.f` denotes the field or method at the shallowest depth in `T` where
    there is such an `f`
    * if there is not exactly one `f` with shallowest depth, the selector
      expression is illegal
* for a value `x` of type `I` where `I` is an interface type
  * `x.f` denotes the actual method with name `f` of the dynamic value of `x`
    * if there is no method with name `f` in the method set of `I`, the
      selector expression is illegal
* if the type of `x` is a named pointer type and `(*x).f` is a valid selector
  expression denoting a field (but not a method)
  * `x.f` is shorthand for `(*x).f`
* in all other cases, `x.f` is illegal
* if `x` is of pointer type and has the value `nil` and `x.f` denotes a struct
  field
  * assigning to or evaluating `x.f` causes a run-time panic
* if `x` is of interface type and has the value `nil`
  * calling or evaluating the method `x.f` causes a run-time panic

### Method expressions

Method expression:
* if `M` is in the method set of type `T`
  * `T.M` is a function that is callable as a regular function with the same
    arguments as `M` prefixed by an additional argument that is the receiver of
    the method
* for a method with a value receiver, one can derive a function with an
  explicit pointer receiver
  * such a function indirects through the receiver to create a value to pass as
    as the receiver to the underlying method
    * the method does not overwrite the value whose address is passed in the
      function call
* a value-receiver function for a pointer-receiver method is illegal because
  pointer-receiver methods are not in the method set of the value type
* function values derived from methods are called with function call syntax
  * see the line marked with `(*)` in the example below
* it is legal to derive a function value from a method of an interface type
  * the resulting function takes an explicit receiver of that interface type

Examples:
```go
// Consider:
type T struct {
	a int
}
func (tv  T) Mv(a int) int         { return 0 } // value receiver
func (tp *T) Mp(f float32) float32 { return 1 } // pointer receiver

var t T

// 1) The expression
T.Mv
// yields a function value representing Mv with signature
func(tv T, a int) int
// That function may be called normally with an explicit receiver, so these
// five invocations are equivalent:
t.Mv(7)
T.Mv(t, 7)
(T).Mv(t, 7)
f1 := T.Mv; f1(t, 7)   // (*)
f2 := (T).Mv; f2(t, 7)

// 2) The expression
(*T).Mp
// yields a function value representing Mp with signature
func(tp *T, f float32) float32

// 3) The expression
(*T).Mv
// yields a function value representing Mv with signature
func(tv *T, a int) int

// 4) The expression
T.Mp
// is illegal, because Mp is not in the method set of T.
```

Grammar for method expression:
```
MethodExpr   = ReceiverType "." MethodName .
ReceiverType = TypeName | "(" "*" TypeName ")" | "(" ReceiverType ")" .
```

### Method values

Method value:
* if the expression `x` has static type `T` and `M` is in the method set of
  type `T`, `x.M` is called a *method value*
  * `T` may be an interface or non-interface type
  * `x.M` is a function value that is callable with the same arguments as a
    method call of `x.M`
  * `x` is evaluated and saved during the evaluation of the method value
    * the saved copy is then used as the receiver in any calls, which may be
      executed later
* a reference to a non-interface method with
  * a value receiver using a pointer will automatically dereference that
    pointer (see 3-3 in the examples below)
  * a pointer receiver using an addressable value will automatically take the
    address of that value (see 3-4 in the examples below)
* it is legal to create a method value from a value of interface type (see 4 in
  the examples below)

Examples:
```go
// Consider:
type T struct {
	a int
}
func (tv  T) Mv(a int) int         { return 0 } // value receiver
func (tp *T) Mp(f float32) float32 { return 1 } // pointer receiver

var t T
var pt *T
func makeT() T

// 1) The expression
t.Mv
// yields a function value of type
func(int) int
// These two invocations are equivalent:
t.Mv(7)
f := t.Mv; f(7)

// 2) The expression
pt.Mp
// yields a function value of type
func(float32) float32

// 3)
f := t.Mv; f(7)   // (3-1) like t.Mv(7)
f := pt.Mp; f(7)  // (3-2) like pt.Mp(7)
f := pt.Mv; f(7)  // (3-3) like (*pt).Mv(7)
f := t.Mp; f(7)   // (3-4) like (&t).Mp(7)
f := makeT().Mp   // (3-5) invalid: result of makeT() is not addressable

// 4)
var i interface { M(int) } = myVal
f := i.M; f(7)  // like i.M(7)
```

### Index expressions

Index expression:
* a primary expression of the form `a[x]`
  * denotes the element of the array, pointer to array, slice, string or map
    `a` indexed by `x`
  * the value of `x` is called the *index* or *map key*, respectively
* assigning to an element of a `nil` map causes a run-time panic

Rules:
* if `a` is not a map
  * `x` must be of integer type or untyped
    * if `0 <= x < len(a)`, it is *in range*
    * otherwise, it is *out of range*
  * a constant index must be non-negative and representable by a value of type
    `int`
* for `a` of array type `A`
  * a constant index must be in range
  * if `x` is out of range at run time, a run-time panic occurs
  * `a[x]` is the array element at index `x` and the type of `a[x]` is the
    element type of `A`
* for `a` of pointer to array type
  * `a[x]` is shorthand for `(*a)[x]`
* for `a` of slice type `S`
  * if `x` is out of range at run time, a run-time panic occurs
  * `a[x]` is the slice element at index `x` and the type of `a[x]` is the
    element type of `S`
* for `a` of string type
  * a constant index must be in range if `a` is also constant
  * if `x` is out of range at run time, a run-time panic occurs
  * `a[x]` is the non-constant byte value at index `x` and the type of `a[x]`
    is `byte`
  * `a[x]` is may not be assigned to
* for `a` of map type `M`
  * `x`'s type must be assignable to the key type of `M`
  * if the map contains an entry with key `x`
    * `a[x]` is the map value with key `x`
    * the type of `a[x]` is the value type of `M`
  * if the map is `nil` or does not contain such an entry
    * `a[x]` is the zero value for the value type of `M`
* otherwise `a[x]` is illegal

On a map `a` of type `map[K]V` used in an assignment or initialization of the
special form
```go
v, ok = a[x]
v, ok := a[x]
var v, ok = a[x]
var v, ok T = a[x]
```
an index expression yields an additional boolean value; the value of `ok` is
* `true` if the key `x` is present in the map
* `false` otherwise

### Slice expressions

Slice expression:
* constructs a substring or slice from a string, array, pointer to array, or
  slice

#### Simple slice expressions

Simple slice expression:
* for a string, array, pointer to array, or slice `a`
  * the primary expression `a[low : high]` constructs a substring or slice
    * the *indices* `low` and `high` select which elements of `a` appear in the
      result
    * the result has indices starting from 0 and length equal to `high - low`
    * any of the indices may be omitted
      * missing `low` defaults to 0
      * missing `high` defaults to the length of the sliced operand
* for a pointer to array `a`
  * `a[low : high]` is shorthand for `(*a)[low : high]`
* for array or string `a`
  * if `0 <= low <= high <= len(a)`, the indices are *in range*
  * otherwise they are *out of range*
* for slice `a`
  * the upper index bound is `cap(a)`
* a constant index must be non-negative and representable by a value of type
  `int`
  * for arrays or constant strings
    * constant indices must also be in range
  * if both indices are constant
    * they must satisfy `low <= high`
* if the indices are out of range at run time, a run-time panic occurs

Result of the slice operation:
* except for untyped strings, if the sliced operand is a string or slice
  * the result of the slice operation is a non-constant value of the same type
    as the operand
* for untyped string operands
  * the result is a non-constant value of type `string`
* if the sliced operand is an array
  * it must be addressable
  * the result of the slice operation is a slice with the same element type as
    the array
* if the sliced operand of a valid slice expression is a `nil` slice
  * the result is a `nil` slice
* otherwise, the result shares its underlying array with the operand

#### Full slice expressions

Full slice expression:
* for an array, pointer to array, or slice `a` (but not a string)
  * the primary expression `a[low : high : max]` constructs a slice of the same
    type, and with the same length and elements as the simple slice expression
    `a[low : high]`
    * additionally, it sets the resulting slice's capacity to `max - low`
    * only the first index may be omitted
      * it defaults to 0
    * if `0 <= low <= high <= max <= cap(a)`
      * the indices are *in range*
      * otherwise they are *out of range*
* a constant index must be non-negative and representable by a value of type
  `int`
  * for arrays
    * constant indices must also be in range
  * if multiple indices are constant
    * the constants that are present must be in range relative to each other
* if the indices are out of range at run time, a run-time panic occurs
* for a pointer to array `a`
  * `a[low : high : max]` is shorthand for `(*a)[low : high : max]`
* if the sliced operand is an array
  * it must be addressable

### Type assertions

Type assertion:
* for an expression `x` of interface type and a type `T`
  * the primary expression `x.(T)` asserts that
    * `x` is not `nil`
    * the value stored in `x` is of type `T`
  * the notation `x.(T)` is called a *type assertion*
* if `T` is not an interface type
  * `x.(T)` asserts that the dynamic type of `x` is identical to the type `T`
    * `T` must implement the (interface) type of `x`
    * otherwise the type assertion is invalid since it is not possible for
      `x` to store a value of type `T`
* if `T` is an interface type
  * `x.(T)` asserts that the dynamic type of `x` implements the interface `T`
* if the type assertion holds
  * the value of the expression is the value stored in `x` and its type is `T`
* if the type assertion is false, a run-time panic occurs

The special form of a type assertion, used in an assignment or initialization,
```go
v, ok = x.(T)
v, ok := x.(T)
var v, ok = x.(T)
var v, ok T1 = x.(T)
```
yields an additional untyped boolean value:
* the value of `ok` is `true` if the assertion holds
* otherwise it is `false` and the value of `v` is the zero value for type `T`

### Calls

Call:
* given an expression `f` of function type `F`
  * `f(a1, a2, ..., an)` calls `f` with arguments `a1`, `a2`, ..., `an`
    * arguments must be single-valued expressions assignable to the parameter
      types of `F`
    * arguments are evaluated before the function is called
    * the type of the expression is the result type of `F`
    * the function value and arguments are evaluated in the usual order
      * after they are evaluated, the parameters of the call are passed by
        value to the function and the called function begins execution
      * the return parameters of the function are passed by value back to the
        calling function when the function returns
    * if the return values of a function or method `g` are equal in number and
      individually assignable to the parameters of another function or method
      `f`
      * the call `f(g(parameters_of_g))` will invoke `f` after binding the
        return values of `g` to the parameters of `f` in order
      * the call of `f` must contain no parameters other than the call of `g`
      * `g` must have at least one return value
      * if `f` has a final `...` parameter
        * it is assigned the return values of `g` that remain after assignment
          of regular parameters
* a method invocation is similar
  * the method itself is specified as a selector upon a value of the receiver
    type for the method
* a method call `x.m()`
  * is valid if
    * the method set of (the type of) `x` contains `m`
    * the argument list can be assigned to the parameter list of `m`
  * if `x` is addressable and `&x`'s method set contains `m`
    * `x.m()` is shorthand for `(&x).m()`
* there is no distinct method type
* there are no method literals
* calling a `nil` function value causes a run-time panic

### Passing arguments to ... parameters

Rules:
* if `f` is variadic with a final parameter `p` of type `...T`
  * then within `f` the type of `p` is equivalent to type `[]T`
* if `f` is invoked with no actual arguments for `p`
  * the value passed to `p` is `nil`
* otherwise, the value passed is a new slice of type `[]T` with a new
  underlying array whose successive elements are the actual arguments, which
  all must be assignable to `T`
  * the length and capacity of the slice is therefore the number of arguments
    bound to `p` and may differ for each call site
* if the final argument is assignable to a slice type `[]T`
  * it may be passed unchanged as the value for `a ...T` parameter if the
    argument is followed by `...`
    * in this case no new slice is created

### Operators

Rules of combining operands into expressions:
* for a binary operators other than comparison
  * the operand types must be identical unless the operation involves shifts or
    untyped constants
  * if all operands are constants, it is the constant expression case
  * if the operation is not shift operation
    * if one operand is an untyped constant and the other operand is not
      * the constant is converted to the type of the other operand
  * in case of shift operation
    * the right operand in a shift expression must have unsigned integer type
      or be an untyped constant representable by a value of type `uint`
    * if the left operand of a non-constant shift expression is an untyped
      constant
      * it is first converted to the type it would assume if the shift
        expression were replaced by its left operand alone

Grammar for expression:
```
Expression = UnaryExpr | Expression binary_op Expression .
UnaryExpr  = PrimaryExpr | unary_op UnaryExpr .

binary_op  = "||" | "&&" | rel_op | add_op | mul_op .
rel_op     = "==" | "!=" | "<" | "<=" | ">" | ">=" .
add_op     = "+" | "-" | "|" | "^" .
mul_op     = "*" | "/" | "%" | "<<" | ">>" | "&" | "&^" .

unary_op   = "+" | "-" | "!" | "^" | "*" | "&" | "<-" .
```

### Arithmetic operators

Rules for arithmetic operators:
* apply to a numeric values
* yield a result of the same type as the first operand

```
+    sum                    integers, floats, complex values, strings
-    difference             integers, floats, complex values
*    product                integers, floats, complex values
/    quotient               integers, floats, complex values
%    remainder              integers

&    bitwise AND            integers
|    bitwise OR             integers
^    bitwise XOR            integers
&^   bit clear (AND NOT)    integers

<<   left shift             integer << unsigned integer
>>   right shift            integer >> unsigned integer
```

#### Integer operators

Rules for integer operators:
* for two integer values `x` and `y`
  * the integer quotient `q = x / y` and remainder `r = x % y` satisfy the
    following relationships, with `x / y` truncated towards zero
    * `x = q*y + r`
    * `|r| < |y|`
  * exception of the previous rule: if the divident `x` is the most negative
    value for the int type of `x`
    * the quotient `q = x / -1` is equal to `x` (and `r = 0`)
  * if the divisor is a constant
    * it must not be zero
  * if the divisor is zero at run time, a run-time panic occurs
  * if the dividend is non-negative and the divisor is a constant power of 2
    * the division may be replaced by a right shift
    * and computing the remainder may be replaced by a bitwise AND operation
  * the shift operators shift the left operand by the shift count specified by
    the right operand
    * if the left operand is a signed integer, they implement arithmetic shifts
    * if it is an unsigned integer, they implement logical shifts
    * there is no upper limit on the shift count
    * shifts behave as if the left operand is shifted `n` times by 1 for a
      shift count of `n`
      * as a result
        * `x << 1` is the same as `x * 2`
        * `x >> 1` is the same as `x / 2` but truncated towards negative
          infinity
* for integer operands
  * the unary operators `+`, `-`, and `^` are defined as follows:
    ```
    +x                          is 0 + x
    -x    negation              is 0 - x
    ^x    bitwise complement    is m ^ x  with m = "all bits set to 1" for
                                          unsigned x and m = -1 for signed x
    ```

#### Integer overflow

For unsigned integer values:
* the operations `+`, `-`, `*`, and `<<` are computed modulo `2**n`, where `n`
  is the bit width of the unsigned integer's type

For signed integers:
* the operations `+`, `-`, `*`, and `<<` may legally overflow and the resulting
  value exists and is deterministically defined by the signed integer
  representation, the operation, and its operands
  * no exception is raised as a result of overflow
  * a compiler may not optimize code under the assumption that overflow does
    not occur; for instance, it may not assume that `x < x + 1` is always true

#### Floating-point operators

For floating-point and complex numbers:
* `+x` is the same as `x`
* `-x` is the negation of `x`
* the result of a floating-point or complex division by zero
  * is not specified beyond the IEEE-754 standard
  * whether a run-time panic occurs is implementation-specific

Combining multiple floating-point operations into a single fused operation:
* depends on implementation
* possibly across statements
* a result that differs from the value obtained by executing and rounding the
  instructions individually may be produced
* a floating-point type conversion explicitly rounds to the precision of the
  target type, preventing fusion that would discard that rounding

#### String concatenation

String concatenation:
* is done by `+` operator or `+=` assignment operator
* creates a new string by concatenating the operands

### Comparison operators

Comparison operators:
```
==    equal
!=    not equal
<     less
<=    less or equal
>     greater
>=    greater or equal
```
* compare two operands
* yield an untyped boolean value
* in any comparison
  * the first operand must be assignable to the type of the second operand, or
    vice versa
* the equality operators `==` and `!=` apply to operands that are *comparable*
* the ordering operators `<`, `<=`, `>`, and `>=` apply to operands that are
  *ordered*

Rules:
* boolean values are comparable
  * two boolean values are equal if they are either both `true` or both `false`
* integer values are comparable and ordered, in the usual way
* floating-point values are comparable and ordered, as defined by the IEEE-754
  standard
* complex values are comparable
  * two complex values `u` and `v` are equal if both `real(u) == real(v)` and
    `imag(u) == imag(v)`
* string values are comparable and ordered, lexically byte-wise
* pointer values are comparable
  * two pointer values are equal if they point to the same variable or if both
    have value `nil`
  * pointers to distinct zero-size variables may or may not be equal
* channel values are comparable
  * two channel values are equal if they were created by the same call to make
    or if both have value `nil`
* interface values are comparable
  * two interface values are equal if they have identical dynamic types and
    equal dynamic values or if both have value `nil`
* a value `x` of non-interface type `X` and a value `t` of interface type `T`
  * are comparable when
    * values of type `X` are comparable
    * `X` implements `T`
  * are equal if
    * `t`'s dynamic type is identical to `X`
    * `t`'s dynamic value is equal to `x`
* struct values are comparable if all their fields are comparable
  * two struct values are equal if their corresponding non-blank fields are
    equal
* array values are comparable if values of the array element type are
  comparable
  * two array values are equal if their corresponding elements are equal
* a comparison of two interface values with identical dynamic types causes a
  run-time panic if values of that type are not comparable
  * applies also when comparing arrays of interface values or structs with
    interface-valued fields
* slice, map, and function values are not comparable
  * a slice, map, or function value may be compared to the predeclared
    identifier `nil`
* pointer, channel, and interface values may be also compared to `nil`

### Logical operators

Logical operators:
* apply to boolean values
* yield a result of the same type as the operands
* the right operand is evaluated conditionally
```
&&    conditional AND    p && q  is  "if p then q else false"
||    conditional OR     p || q  is  "if p then true else q"
!     NOT                !p      is  "not p"
```

### Address operators

For an operand `x` of type `T`:
* the address operation `&x` generates a pointer of type `*T` to `x`
* the operand must be *addressable*
  * either a variable, pointer indirection, or slice indexing operation
  * or a field selector of an addressable struct operand
  * or an array indexing operation of an addressable array
* `x` may also be a (possibly parenthesized) composite literal
* if the evaluation of `x` would cause a run-time panic
  * the evaluation of `&x` does too

For an operand `x` of pointer type `*T`
* the pointer indirection `*x` denotes the variable of type `T` pointed to by
  `x`
  * if `x` is `nil`, an attempt to evaluate `*x` will cause a run-time panic

### Receive operator

For an operand `ch` of channel type:
* the value of the receive operation `<-ch` is the value received from the
  channel `ch`
* the channel direction must permit receive operations
* the type of the receive operation is the element type of the channel
* the expression blocks until a value is available
* receiving from a `nil` channel blocks forever
* a receive operation on a closed channel can always proceed immediately
  * it yields the element type's zero value after any previously sent values
    have been received

The special form of a receive expression, used in an assignment or
initialization,
```go
x, ok = <-ch
x, ok := <-ch
var x, ok = <-ch
var x, ok T = <-ch
```
yields an additional untyped boolean result reporting whether the communication
succeeded:
* if the value received was delivered by a successful send operation to the
  channel
  * the value of `ok` is `true`
* if it is a zero value generated because the channel is closed and empty
  * the value of `ok` is `false`

### Conversions

Conversions:
* are expressions of the form `T(x)` where `T` is a type and `x` is an
  expression that can be converted to type `T`
* a constant value `x` can be converted to type `T` in any of these cases
  * `x` is representable by a value of type `T`
  * `x` is a floating-point constant, `T` is a floating-point type, and `x` is
    representable by a value of type `T` after rounding using IEEE 754
    round-to-even rules, but with an IEEE `-0.0` further rounded to an unsigned
    `0.0`
    * the constant `T(x)` is the rounded value
  * `x` is an integer constant and `T` is a string type
    * the same rule as for non-constant `x` applies in this case
* converting a constant yields a typed constant as result
* a non-constant value `x` can be converted to type `T` in any of these cases
  * `x` is assignable to `T`
  * ignoring struct tags (see below), `x`'s type and `T` have identical
    underlying types
  * ignoring struct tags (see below), `x`'s type and `T` are pointer types that
    are not defined types, and their pointer base types have identical
    underlying types
  * `x`'s type and `T` are both integer or floating point types
  * `x`'s type and `T` are both complex types
  * `x` is an integer or a slice of bytes or runes and `T` is a string type
  * `x` is a string and `T` is a slice of bytes or runes
* struct tags are ignored when comparing struct types for identity for the
  purpose of conversion

Grammar for conversion:
```
Conversion = Type "(" Expression [ "," ] ")" .
```

Notes:
* there is no linguistic mechanism to convert between pointers and integers
  * the package `unsafe` implements this functionality under restricted
    circumstances

#### Conversion between numeric types

Rules for the conversion of non-constant numeric values:
* when converting between integer types
  * if the value is a signed integer
    * it is sign extended to implicit infinite precision
  * otherwise it is zero extended
  * it is then truncated to fit in the result type's size
* when converting a floating-point number to an integer
  * the fraction is discarded (truncation towards zero)
* when converting an integer or floating-point number to a floating-point type,
  or a complex number to another complex type
  * the result value is rounded to the precision specified by the destination
    type
* in all non-constant conversions involving floating-point or complex values
  * if the result type cannot represent the value the conversion succeeds but
    the result value is implementation-dependent

#### Conversion to and from a string type

Rules:
* converting a signed or unsigned integer value to a string type
  * yields a string containing the UTF-8 representation of the integer
  * values outside the range of valid Unicode code points are converted to
    `"\uFFFD"`
* converting a slice of bytes to a string type
  * yields a string whose successive bytes are the elements of the slice
* converting a slice of runes to a string type
  * yields a string that is the concatenation of the individual rune values
    converted to strings
* converting a value of a string type to a slice of bytes type
  * yields a slice whose successive elements are the bytes of the string
* converting a value of a string type to a slice of runes type
  * yields a slice containing the individual Unicode code points of the string

### Constant expressions

Constant expressions:
* may contain only constant operands
* are evaluated at compile time
* untyped boolean, numeric, and string constants
  * may be used as operands wherever it is legal to use an operand of boolean,
    numeric, or string type, respectively
* except for shift operations, if the operands of a binary operation are
  different kinds of untyped constants
  * the operation and, for non-boolean operations, the result use the kind that
    appears later in this list: integer, rune, floating-point, complex
* a constant comparison always yields an untyped boolean constant
* if the left operand of a constant shift expression is an untyped constant
  * the result is an integer constant
  * otherwise it is a constant of the same type as the left operand, which must
    be of integer type
* applying all other operators to untyped constants
  * results in an untyped constant of the same kind (that is, a boolean,
    integer, floating-point, complex, or string constant)
* applying the built-in function `complex` to untyped integer, rune, or
  floating-point constants
  * yields an untyped complex constant
* constant expressions are always evaluated exactly
* the divisor of a constant division or remainder operation must not be zero
* the values of *typed* constants must always be accurately representable as
  values of the constant type
* the mask used by the unary bitwise complement operator `^` matches the rule
  for non-constants
  * the mask is all 1s for unsigned constants and -1 for signed and untyped
    constants

Implementation restriction:
* a compiler may use rounding while computing untyped floating-point or complex
  constant expressions (see the implementation restriction in the section on
  constants)
* this rounding may cause a floating-point constant expression to be invalid in
  an integer context, even if it would be integral when calculated using
  infinite precision, and vice versa

### Order of evaluation

Order of evaluation rules:
* at package level
  * initialization dependencies determine the evaluation order of individual
    initialization expressions in variable declarations
    * they override the left-to-right rule for individual initialization
      expressions noted below, but not for operands within each expression
      (see 3 in examples below)
* otherwise, when evaluating the operands of an expression, assignment, or
  return statement
  * all function calls, method calls, and communication operations are
    evaluated in lexical left-to-right order (see 1 in examples below)
* floating-point operations within a single expression are evaluated according
  to the associativity of the operators
  * explicit parentheses affect the evaluation by overriding the default
    associativity

Examples:
```go
// 1) In the (function-local) assignment
y[f()], ok = g(h(), i()+x[j()], <-c), k()

// the function calls and communication happen in the order
//   f(), h(), i(), j(), <-c, g(), and k().
// The order of those events compared to the evaluation and indexing of x and
// the evaluation of y is not specified.

// 2)
a := 1
f := func() int { a++; return a }
x := []int{a, f()}            // x may be [1, 2] or [2, 2]:
                              //   evaluation order between a and f() is not
                              //   specified;
m := map[int]int{a: 1, a: 2}  // m may be {2: 1} or {2: 2}:
                              //   evaluation order between the two map
                              //   assignments is not specified;
n := map[int]int{a: f()}      // n may be {2: 3} or {3: 3}:
                              //   evaluation order between the key and the
                              //   value is not specified.

// 3) In
var a, b, c = f() + v(), g(), sqr(u()) + v()

func f() int        { return c }
func g() int        { return a }
func sqr(x int) int { return x*x }

// functions u and v are independent of all other variables and functions. The
// function calls happen in the order u(), sqr(), v(), f(), v(), and g().
```
