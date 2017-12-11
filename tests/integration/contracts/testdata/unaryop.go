package contracts

func f() string {
	a := "ahoj"
	ra := &a

	chanA := make(chan int)
	chanValA := <-chanA

	uopa := ^1
	uopb := -1
	uopc := !true
	uopd := +1
}
