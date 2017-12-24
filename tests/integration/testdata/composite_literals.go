package testdata

type Int int

func f() {
	list := []Int{
		1,
		2,
	}

	mapV := map[string]int{
		"3": 3,
		"4": 4,
	}

	structV := struct {
		key1 string
		key2 int
	}{
		key1: "key1",
		key2: 2,
	}

	structV2 := struct {
		key1 string
		key2 int
	}{
		"key1",
		2,
	}

	listV2 := [][]int{
		{
			1,
			2,
		},
		{
			3,
			4,
		},
	}
}
