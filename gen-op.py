import os

# Generate goscript into the stdout. run
# python gen-op.py > result.go; go fmt result.go

#TODO: support unary types too
BIN_OP = {
    "ADD": "+",
    "SUB": "-",
    "MUL": "*",
    "QUO": "/",
    "REM": "%",

    "AND":     "&",
    "OR":      "|",
    "XOR":     "^",
    "SHL":     "<<",
    "SHR":     ">>",
    "AND_NOT": "&^",

    "LAND":  "&&",
    "LOR":   "||",
#    "ARROW": "<-",

    "EQL":    "==",
    "LSS":    "<",
    "GTR":    ">",

    "NEQ":      "!=",
    "LEQ":      "<=",
    "GEQ":      ">=",
}

#TODO just check validity - unary op can't be used inside print func
UN_OP = {
    "INC":   "++",
    "DEC":   "--",
#    "ASSIGN": "=",
#    "NOT":    "!",
}

TYPES = [
     ("int(5)"          , "int"       , "IntegerType"  ),
     ("int16(5)"        , "int16"     , "IntegerType"  ), # MAY DUPLA
     ("int32(5)"        , "int32"     , "IntegerType"  ), # MAY DUPLA
     ("int64(5)"        , "int64"     , "IntegerType"  ), # MAY DUPLA
     ("uint(5)"         , "uint"      , "IntegerType"  ),
     ("uint16(5)"       , "uint16"    , "IntegerType"  ), # MAY DUPLA
     ("uint32(5)"       , "uint32"    , "IntegerType"  ), # MAY DUPLA
     ("uint64(5)"       , "uint64"    , "IntegerType"  ), # MAY DUPLA
     ("float32(5.0)"    , "float32"   , "FloatType"    ),
     ("float64(5.0)"    , "float64"   , "FloatType"    ), # MAY DUPLA
     ("complex(1,2)"    , "complex128", "ImagType"     ),
# yes, byte & rune are aliases for int int32, but who knows in future
     ("byte('c')"       , "byte"      , "CharType"     ),
     ("rune('c')"       , "rune"      , "CharType"     ),
     ("string(\"Ahoj\")", "string"    , "StringType"   ),
    #"NilType"     : "nil"
]

gofile_top="""
package main

import (
\t"fmt"
)

"""

#Will we want to generate own types in future (for this test) too?
#var (
#"""

def test_validity(type1, type2, op):
    with open("mygo_val.go", 'w') as handle:
        handle.write(gofile_top)
        handle.write("func main() {\n")
        opstr = "%s %s %s" % (type1[0], op ,type2[0])
        handle.write("fmt.Printf(\"%s\", %s, %s, %s)\n" % ("%T, %T, %T",
                                                           type1[0],
                                                           type2[0],
                                                           opstr))
        handle.write("}\n")
    rval = os.system("go run mygo_val.go >/dev/null 2>/dev/null")
    os.system("rm -f mygo_val.go")
    return rval == 0



print(gofile_top)
print("func main() {\n")
for v in TYPES:
    for vv in TYPES:
        for ok, ov in BIN_OP.iteritems():
            opstr = "%s %s %s" % (v[0], ov ,vv[0])
            if test_validity(v, vv, ov):
                print("fmt.Printf(\"%s\", %s, %s, \"%s\", %s)" % (
                            "%T, %T, %s, %T\\n",
                            v[0],
                            vv[0],
                            ov,
                            opstr))
print "}"
