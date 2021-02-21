package testdata

func summarize() uint8 {
	for j := 0; j < 64; j += 8 {
		return uint8(uint64(2) >> j)
	}
}
