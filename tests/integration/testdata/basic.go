package testdata

// Just get a list of all contracts and propagate all data types on more time

type A struct {
	i int
	s string
	f struct {
		f float32
	}
}

func (a *A) method(x, y int) int {
	return x + y
}

var C = &A{}

func g() int {
	a := &A{
		i: 1,
		s: "1",
		f: struct {
			f float32
		}{
			f: 1.0,
		},
	}

	b := a.method(a.i, a.i)
	return b
}

type MyInt int

const i8 int8 = 2
const i16 int16 = 2
const i32 int32 = 2
const i64 int64 = 2
const u8 uint8 = 2
const ui uint = 2
const uip uintptr = 2
const f32 float32 = 2
const f64 float64 = 2

const m8 MyInt = 2

const j int8 = 2 * 10 /////////////
const j1 int8 = 3 / 2
const j2 = 2.0 << 1.0

var j3 uint = 2
var j4 = 2 << j3
var j5 = j3 << 2.0
var j6 = (2 << j3) * 2
var j7 = j3 * 2.0
var j8 = ((2 << j3) * uint8(2)) * 2.0

var j9 = 2 * (2 << j3)
var j10 = 2.0 * j3
var j11 = 2.0 * (uint8(2) * (2 << j3))

var j12 = (2 << j3) * (2 << j3)
var j13 = (2 << j3) * j3
var j14 = j3 * (2 << j3)
var j15 = j3 * j3

// test valueLess numbers
var j16 = 2 & 1
var j17 = 2 % (2 << j3)
var j18 = i8 % (2 << j3)
var j19 = 2 % j3
var j20 = ui % j3

var j21 = 2 % (2 << j3)
var j22 = i8 % (2 << j3)
var j23 = 2 % j3
var j24 = ui % j3

var j25 = (2 << j3) % (2 << j3)
var j26 = (2 << j3) % j3
var j27 = j3 % (2 << j3)
var j28 = j3 % j3

var j29 = 2 * 2.1
var j30 = 5 * 1
var j31 = 5 * 2.0
var j32 = 5 * uip

var j33 = ((2 << j3) << j3)
var j34 = ((2 << j3) % (2 << j3)) << j3

var j35 = 1 * 5
var j36 = 2.0 * 5
var j37 = uip * 5

var j38 = (2 << j3) * 5
var j39 = j37 * 5
var j40 = 5 * (2 << j3)
var j41 = 5 * j37

var j42 = 1 % 5
var j43 = 5 % 1
var j44 = uip % 5
var j45 = 5 % uip
var j46 = 5 % 5
var j47 = 5 * 5

var j49 = (2 << j3) % 5
var j50 = j37 % 5
var j51 = 5 * (2 << j3)
var j52 = 5 * j37

var j53 = 1 << 5
var j54 = 5 << 1
var j55 = 5 << 5
var j56 = (1 << 5) * 2.1
var j57 = (1 << 5) * (1 << 5)
var j58 = (5 << 1) * 2.0
var j59 = (2.0 + 0.0i) * f64
var j60 = complex64(f32) * (2.0 + 0.1i)

var j61 = 1 < 5
var j62 = (2 << j3) > 0
var j63 = j3 > 0.0
var j64 = 5 > 0.0
var j65 = 0.0 < j3
var _ = 0.0 < 5
var _ = j3 > j3
var _ = i8 > i8
var _ = j3 > ui

type TS struct {
	a [2 + 3.0]int
}

//const j70 = 1 + len(TS{}.a)
const j71 = len([...]int{1, 2, 3})

type I1 interface {
	a(x, y int) (int, string)
}

type I2 interface {
	a(x, y int) (int, string)
}

type T1 struct{}

func (t *T1) a(x, y int) (int, string) { return 0, "0" }

type T2 struct{}

func (t *T2) a(x, y int) (int, string) { return 0, "0" }

var i1 I1 = &T1{}
var i2 I2 = &T2{}

var _ = i1 == i2
var _ = i1 == &T2{}

const _ = complex(1, 1)
const _ = complex(float32(1), 1)
const _ = complex(float64(1), float64(1))

var vf32 float32 = 2
var vf64 float64 = 2
var _ = complex(1, vf32)
var _ = complex(vf64, 1)
var _ = complex(vf64, vf64)

//var j8 = 2.0 * j3

// const j8 = m8 * 2
//
// // var vm8 MyInt = 2
// // var vm8_2 = vm8 * 2
// // var vm8_8 = vm8 * vm8
//
// const f = 3.1
// const f2 = f * 3.2
//
// const f32 float32 = 3.0
// const f64 float64 = 3.0
//
// const f3 = f32 * f
// const f4 = f * f32
// const f5 = f3 * f3
//
// const i_f int8 = 2 * 3.0
// const i_f2 = 2 * i_f
// const i_f3 = f2 + 100
// const i_f4 = i_f3 + i_f3
//
// const s_1 = 4.0 << 2.0
// const s_2 = 9.0 >> 2.0
// const s_3 = 4 << 2
// const s_4 = 9.0 % 2.0
// const s_5 = 3.0 == 2
// const s_6 = 3.0 != 2
// const s_7 = 3.0 >= 2
// const s_8 = 3.0 > 2
// const s_9 = 3.0 <= 2
// const s_10 = 3.0 < 2
// const s_11 = 2 & 3
// const s_12 = 2 | 3
// const s_13 = 7 &^ 3
// const s_14 = 7 ^ 3
//
// const b_1 = true && false
// const b_2 = true && true
// const b_3 = true || false
//
// type MyBool bool
//
// const b_4 MyBool = true
// const b_5 = b_4 && false
// const b_6 = true || b_4
//
// var b_7 = b_6
// var b_8 = b_7 || b_4
//
// type MyStr string
//
// const ss_1 = "\"a" + "a"
// const ss_2 = "a" > "a"
//
// var ss_3 MyStr = "a"
// var ss_4 = ss_3 > "a"
// var ss_5 = ss_3 > ss_3
//
// var v_1, _, v_3 int
// var v_4, v_5 = 1, 2
// var v_6 = ss_4 && true
// var v_7 = b_4 && true
// var v_8 = b_4 && (ss_3 > ss_3)
// var v_9 = v_7 && (ss_3 > ss_3)
// var v_10 = v_9 || v_8
//
// const cx_1 = 3.1111e2i + 2.0
// const cx_2 = cx_1 + cx_1
// const cx_3 = cx_1 * cx_1
// const cx_4 = cx_3 / cx_1
// const cx_5 = cx_3 - cx_1
// const cx_6 = cx_3 == cx_1

// const bf_1 = unsafe.Sizeof(struct{}{})

// var b_7 = true

// const (
// 	a, aa = i8 * iota, 10 * iota
// 	b, bb
// )
