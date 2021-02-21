package testdata

const (
	pallocChunksL1Bits  = 13
	_PageShift          = 13
	GoosDarwin          = 0
	GoarchArm64         = 0
	GoarchMipsle        = 0
	GoarchMips          = 0
	GoarchWasm          = 0
	_64bit              = 1 << (^uintptr(0) >> 63) / 2
	pageShift           = _PageShift
	logPallocChunkPages = 9
	logPallocChunkBytes = logPallocChunkPages + pageShift
	heapAddrBits        = (_64bit*(1-GoarchWasm)*(1-GoosDarwin*GoarchArm64))*48 + (1-_64bit+GoarchWasm)*(32-(GoarchMips+GoarchMipsle)) + 33*GoosDarwin*GoarchArm64
	pallocChunksL2Bits  = heapAddrBits - logPallocChunkBytes - pallocChunksL1Bits
)

func main() {
	uint(1) & (1<<pallocChunksL2Bits - 1)
}
