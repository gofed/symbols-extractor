package testdata

type Int int

func f() {
	// case token.EQL, token.NEQ, token.LEQ, token.LSS, token.GEQ, token.GTR:
	bopa := 1 == 2
	bopa = 1 != 2
	bopa = 1 <= 2
	bopa = 1 < 2
	bopa = 1 >= 2
	bopa = 1 > 2
	// case token.SHL, token.SHR
	bopb := 8.0 << 1
	bopb = 8.0 >> 1
	var bopc int32
	bopc = bopc << 1
	bopc = bopc >> 1
	// case token.AND, token.OR, token.MUL, token.SUB, token.QUO, token.ADD, token.AND_NOT, token.REM, token.XOR:
	bopd := 1 & 0
	bopd = 1 | 0
	bopd = 1 &^ 0
	bopd = 1 ^ 0
	bope := 1 * 1
	bope = 1 - 1
	bope = 1 / 1
	bope = 1 + 1
	bope = 1 % 1
	var bopf int16
	bope = bopf % 1
	bopd = true && false
	var bopg Int
	bope = bopg + 1
	bope = 1 + bopg
}
