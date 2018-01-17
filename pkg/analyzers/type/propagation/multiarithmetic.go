package propagation

import (
	"fmt"
	"go/token"
	"math"
	"math/big"
	"strings"

	// Big thanks to https://godoc.org/github.com/shopspring/decimal
	"github.com/shopspring/decimal"
)

type MultiArith struct {
	X, Y, Z    *decimal.Decimal
	Xi, Yi, Zi *decimal.Decimal
	xc, yc, zc bool
	zBool      bool
	err        error
}

func NewMultiArith() *MultiArith {
	return &MultiArith{}
}

func (m *MultiArith) parseLiteral(literal string) (*decimal.Decimal, *decimal.Decimal, bool, error) {
	ll := len(literal)
	if literal[ll-1] == 'i' {
		parts := strings.Split(literal[:ll-1], ",")
		// no real part
		if len(parts) == 1 {
			m.xc = true
			real := decimal.NewFromFloat(0)
			imag, err := decimal.NewFromString(parts[0])
			if err != nil {
				return nil, nil, false, fmt.Errorf("Unable to create complex part of a complex number: %v", err)
			}
			return &real, &imag, true, nil
		}

		real, err := decimal.NewFromString(parts[0])
		if err != nil {
			return nil, nil, false, fmt.Errorf("Unable to create real part of a complex number: %v", err)
		}

		imag, err := decimal.NewFromString(parts[1])
		if err != nil {
			return nil, nil, false, fmt.Errorf("Unable to create complex part of a complex number: %v", err)
		}

		return &real, &imag, true, nil
	}

	// starts with 0x? => hexadecimal number
	if strings.HasPrefix(literal, "0x") {
		i := new(big.Int)
		if d, _ := i.SetString(literal, 0); d == nil {
			return nil, nil, false, fmt.Errorf("Unable to parse literal %v", literal)
		}
		d := decimal.NewFromBigInt(i, 0)
		return &d, nil, false, nil
	}

	d, err := decimal.NewFromString(literal)
	if err != nil {
		return nil, nil, false, err
	}
	return &d, nil, false, nil
}

func (m *MultiArith) AddXFromString(literal string) *MultiArith {
	if m.err != nil {
		return m
	}

	real, image, ic, err := m.parseLiteral(literal)
	if err != nil {
		m.err = err
		return m
	}

	m.X = real
	m.Xi = image
	m.xc = ic

	return m
}

func (m *MultiArith) AddYFromString(literal string) *MultiArith {
	if m.err != nil {
		return m
	}

	real, image, ic, err := m.parseLiteral(literal)
	if err != nil {
		m.err = err
		return m
	}

	m.Y = real
	m.Yi = image
	m.yc = ic

	return m
}

func (m *MultiArith) ZFloor() *MultiArith {
	d := m.Z.Floor()
	m.Z = &d
	return m
}

func (m *MultiArith) PerformUnary(op token.Token) *MultiArith {
	if m.err != nil {
		return m
	}

	switch op {
	case token.SUB:
		z := m.X.Neg()
		m.Z = &z
		if m.xc {
			z := m.Xi.Neg()
			m.Zi = &z
		}
		return m
	default:
		panic(fmt.Sprintf("MultiArith.Perform NYI for %v op", op))
	}
}

func (m *MultiArith) Perform(op token.Token) *MultiArith {
	if m.err != nil {
		return m
	}

	mulCC := func(a, b, c, d decimal.Decimal) (decimal.Decimal, decimal.Decimal) {
		ac := a.Mul(c)
		bd := b.Mul(d)
		// ac - bd
		real := ac.Sub(bd)

		ad := a.Mul(d)
		bc := b.Mul(c)
		// ad + bc
		imag := ad.Add(bc)

		return real, imag
	}

	switch op {
	case token.MUL:
		if !m.xc && !m.yc {
			z := m.X.Mul(*m.Y)
			m.Z = &z
			return m
		}

		var real, imag decimal.Decimal
		if m.xc && m.yc {
			real, imag = mulCC(*m.X, *m.Xi, *m.Y, *m.Yi)
		} else if !m.xc {
			real = m.Y.Mul(*m.X)
			imag = m.Yi.Mul(*m.X)
		} else {
			real = m.X.Mul(*m.Y)
			imag = m.Xi.Mul(*m.Y)
		}

		m.Z = &real
		m.Zi = &imag
	case token.QUO:
		if !m.xc && !m.yc {
			z := m.X.Div(*m.Y)
			m.Z = &z
			return m
		}
		// (a+bi)/(c+di) = (a+bi)*(c-di)/(c^2+d^2)
		var real, imag decimal.Decimal
		if m.xc && m.yc {
			cc := m.Y.Mul((*m.Y))
			dd := m.Yi.Mul((*m.Yi))
			ccdd := cc.Add(dd)
			x, y := mulCC(*m.X, *m.Xi, *m.Y, m.Yi.Neg())
			real, imag = x.Div(ccdd), y.Div(ccdd)
		} else if !m.xc {
			cc := m.Y.Mul(*m.Y)
			dd := m.Yi.Mul(*m.Yi)
			ccdd := cc.Add(dd)
			x, y := mulCC(*m.X, decimal.NewFromFloat(0), *m.Y, m.Yi.Neg())
			real, imag = x.Div(ccdd), y.Div(ccdd)
		} else {
			real = m.X.Div(*m.Y)
			imag = m.Xi.Div(*m.Y)
		}
		m.Z = &real
		m.Zi = &imag
	case token.REM:
		z := m.X.Mod(*m.Y)
		m.Z = &z
	case token.ADD:
		z := m.X.Add(*m.Y)
		m.Z = &z
		if m.xc || m.yc {
			if m.Xi == nil {
				m.Zi = m.Yi
				return m
			}
			if m.Yi == nil {
				m.Zi = m.Xi
				return m
			}
			zi := m.Xi.Add(*m.Yi)
			m.Zi = &zi
		}
	case token.SUB:
		z := m.X.Sub(*m.Y)
		m.Z = &z
		if m.xc && m.yc {
			zi := m.Xi.Sub(*m.Yi)
			m.Zi = &zi
			return m
		}
		if m.xc {
			m.Zi = m.Xi
			return m
		}
		if m.yc {
			d := m.Yi.Neg()
			m.Zi = &d
			return m
		}
	case token.SHL:
		// check the x is integral and the y is non-negative integral
		if !isInt(m.X) {
			m.err = fmt.Errorf("%v is not int", m.X.String())
			return m
		}
		if !isInt(m.Y) || m.Y.LessThan(decimal.NewFromFloat(0)) {
			m.err = fmt.Errorf("%v is not uint", m.Y.String())
			return m
		}
		// x << y = x*2^y
		z := (*m.X).Mul(decimal.NewFromFloat(2).Pow(*m.Y))
		m.Z = &z
	case token.SHR:
		// check the x is integral and the y is non-negative integral
		if !isInt(m.X) {
			m.err = fmt.Errorf("%v is not int", m.X.String())
			return m
		}
		if !isInt(m.Y) || m.Y.LessThan(decimal.NewFromFloat(0)) {
			m.err = fmt.Errorf("%v is not uint", m.Y.String())
			return m
		}
		// x >> y = x/2^y
		z := (*m.X).Div(decimal.NewFromFloat(2).Pow(*m.Y)).Floor()
		m.Z = &z
	case token.EQL:
		m.zBool = (*m.X).Equal(*m.Y)
	case token.NEQ:
		m.zBool = !(*m.X).Equal(*m.Y)
	case token.LEQ:
		m.zBool = (*m.X).LessThanOrEqual(*m.Y)
	case token.LSS:
		m.zBool = (*m.X).LessThan(*m.Y)
	case token.GEQ:
		m.zBool = (*m.X).GreaterThanOrEqual(*m.Y)
	case token.GTR:
		m.zBool = (*m.X).GreaterThan(*m.Y)
	case token.AND, token.OR, token.XOR, token.AND_NOT:
		xF, _, eX := new(big.Float).Parse((*m.X).String(), 10)
		if eX != nil {
			m.err = eX
			return m
		}
		xI, _ := xF.Int(nil)
		yF, _, eY := new(big.Float).Parse((*m.Y).String(), 10)
		if eY != nil {
			m.err = eY
			return m
		}
		yI, _ := yF.Int(nil)
		zI := new(big.Int)
		switch op {
		case token.AND:
			zI.And(xI, yI)
		case token.OR:
			zI.Or(xI, yI)
		case token.XOR:
			zI.Xor(xI, yI)
		case token.AND_NOT:
			xBL := xI.BitLen()
			yBL := yI.BitLen()
			zI.Set(xI)
			for i := 0; i < xBL && i < yBL; i++ {
				if yI.Bit(i) == 1 {
					zI.SetBit(zI, i, 0)
				}
			}
		}
		z, _ := decimal.NewFromString(zI.String())
		m.Z = &z
	default:
		panic(fmt.Sprintf("MultiArith.Perform NYI for %v op", op))
	}
	return m
}

func checkIntegerRanges(x *decimal.Decimal, targetType string) error {
	var top, bottom decimal.Decimal
	switch targetType {
	case "int":
		// TODO(jchaloup): the architecture must be input of the processing
		top = decimal.NewFromFloat(math.MaxInt32)
		bottom = decimal.NewFromFloat(math.MinInt32)
	case "int8":
		top = decimal.NewFromFloat(math.MaxInt8)
		bottom = decimal.NewFromFloat(math.MinInt8)
	case "int16":
		top = decimal.NewFromFloat(math.MaxInt16)
		bottom = decimal.NewFromFloat(math.MinInt16)
	case "int32":
		top = decimal.NewFromFloat(math.MaxInt32)
		bottom = decimal.NewFromFloat(math.MinInt32)
	case "int64":
		// for some reason the NewFromFloat creates negative float64 number
		top, _ = decimal.NewFromString(fmt.Sprintf("%v", math.MaxInt64))
		bottom = decimal.NewFromFloat(math.MinInt64)
	case "uint":
		// TODO(jchaloup): the architecture must be input of the processing
		top = decimal.NewFromFloat(math.MaxUint32)
		bottom = decimal.NewFromFloat(0)
	case "uint8":
		top = decimal.NewFromFloat(math.MaxUint8)
		bottom = decimal.NewFromFloat(0)
	case "uint16":
		top = decimal.NewFromFloat(math.MaxUint16)
		bottom = decimal.NewFromFloat(0)
	case "uint32":
		top = decimal.NewFromFloat(math.MaxUint32)
		bottom = decimal.NewFromFloat(0)
	case "uint64":
		top = decimal.NewFromFloat(math.MaxUint64)
		bottom = decimal.NewFromFloat(0)
	case "rune":
		top = decimal.NewFromFloat(math.MaxInt32)
		bottom = decimal.NewFromFloat(math.MinInt32)
	case "byte":
		top = decimal.NewFromFloat(math.MaxUint8)
		bottom = decimal.NewFromFloat(0)
	case "uintptr":
		if (*x).Cmp(decimal.NewFromFloat(0)) < 0 {
			return fmt.Errorf("%v overflows %v", x.String(), targetType)
		}
		return nil
	default:
		panic(fmt.Errorf("%v type not recognized", targetType))
		return fmt.Errorf("%v type not recognized", targetType)
	}
	if (*x).Cmp(top) > 0 {
		return fmt.Errorf("%v top %v overflows %v", x.String(), top.String(), targetType)
	}
	if (*x).Cmp(bottom) < 0 {
		return fmt.Errorf("%v bottom %v overflows %v", x.String(), bottom.String(), targetType)
	}
	return nil
}

func (m *MultiArith) IsXComplex() bool {
	return m.Xi != nil && !m.Xi.Equal(decimal.NewFromFloat(0))
}

func (m *MultiArith) IsXFloat() bool {
	return !isXComplex(m.Xi)
}

func (m *MultiArith) IsXInt() bool {
	return !isXComplex(m.Xi) && isInt(m.X)
}

func (m *MultiArith) IsXUint() bool {
	return isInt(m.X) && !m.X.LessThan(decimal.NewFromFloat(0))
}

func isXComplex(xI *decimal.Decimal) bool {
	return xI != nil && !xI.Equal(decimal.NewFromFloat(0))
}

func isInt(x *decimal.Decimal) bool {
	iX := (*x).Round(0)
	return (*x).Equal(iX)
}

func checkFloatRanges(x *decimal.Decimal, targetType string) error {
	var top, bottom decimal.Decimal
	switch targetType {
	case "float32":
		top, _ = decimal.NewFromString("3.40282346638528859811704183484516925440e+38")
		bottom, _ = decimal.NewFromString("-3.40282346638528859811704183484516925440e+38")
	case "float64":
		top, _ = decimal.NewFromString("1.797693134862315708145274237317043567981e+308")
		bottom, _ = decimal.NewFromString("-1.797693134862315708145274237317043567981e+308")
	default:
		// integer type?
		iX := (*x).Round(0)
		if (*x).Equal(iX) {
			return checkIntegerRanges(&iX, targetType)
		}
		return fmt.Errorf("%v truncated to integer", (*x).String())
	}
	if (*x).Cmp(top) > 0 {
		return fmt.Errorf("%v up overflows %v", x.String(), targetType)
	}
	if (*x).Cmp(bottom) < 0 {
		return fmt.Errorf("%v down overflows %v", x.String(), targetType)
	}
	return nil
}

func (m *MultiArith) toLiteral(real, imag *decimal.Decimal, targetType string, typeCheck bool) (string, error) {
	switch targetType {
	case "complex64":
		if typeCheck {
			if err := checkFloatRanges(real, "float32"); err != nil {
				return "", err
			}
			if imag != nil {
				if err := checkFloatRanges(imag, "float32"); err != nil {
					return "", err
				}
			}
		}
		if imag != nil {
			return fmt.Sprintf("%v,%vi", real.String(), imag.String()), nil
		}
		return real.String(), nil
	case "complex128":
		if typeCheck {
			if err := checkFloatRanges(real, "float64"); err != nil {
				return "", err
			}
			if imag != nil {
				if err := checkFloatRanges(imag, "float64"); err != nil {
					return "", err
				}
			}
		}
		if imag != nil {
			return fmt.Sprintf("%v,%vi", real.String(), imag.String()), nil
		}
		return real.String(), nil
	default:
		if typeCheck {
			if err := checkFloatRanges(real, targetType); err != nil {
				return "", err
			}
		}
		return real.String(), nil
	}
}

func (m *MultiArith) XToLiteral(targetType string) (string, error) {
	return m.toLiteral(m.X, m.Xi, targetType, true)
}

func (m *MultiArith) YToLiteral(targetType string) (string, error) {
	return m.toLiteral(m.Y, m.Yi, targetType, true)
}

func (m *MultiArith) ZToLiteral(targetType string, typeCheck bool) (string, error) {
	return m.toLiteral(m.Z, m.Zi, targetType, typeCheck)
}

func (m *MultiArith) ZBool() bool {
	return m.zBool
}

func (m *MultiArith) Error() error {
	return m.err
}
