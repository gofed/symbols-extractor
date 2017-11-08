# Test suite for Symbols Extractor

This document describes a design and implementation of various tests used for
verification that *Symbols Extractor* works properly.

## Data types propagation tests

Data types can be propagated:
* within a same scope, file and package
* across scopes, but within a same file and package
* across scopes and files, but within a same package
* across packages

### Propagation within a same scope

#### Suite no. 1

##### Test no. 1

File set:
* `$TESTSROOT/typeprop/suite001/test001/ver1.00/main.go`
  ```go
  // Original package, no propagation
  package main

  import "fmt"

  var x int = 1

  func main() {
  	fmt.Printf("%v is of a type %T\n", x, x)
  }
  ```
* `$TESTSROOT/typeprop/suite001/test001/ver1.01/main.go`
  ```go
  // Original package, x's name was changed to y
  package main

  import "fmt"

  var y int = 1

  func main() {
  	fmt.Printf("%v is of a type %T\n", y, y)
  }
  ```
* `$TESTSROOT/typeprop/suite001/test001/ver1.02/main.go`
  ```go
  // Original package, x's type was changed
  package main

  import "fmt"

  var x byte = 1

  func main() {
  	fmt.Printf("%v is of a type %T\n", x, x)
  }
  ```

Expected results:
```
v1.00 X v1.01: (x, int, main) -> (y, int, main)
v1.00 X v1.02: (x, int, main) -> (x, byte, main)
```
