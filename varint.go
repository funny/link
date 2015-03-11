package link

func UvarintSize(x uint64) int {
	i := 0
	for x >= 0x80 {
		x >>= 7
		i++
	}
	return i + 1
}

func VarintSize(x int64) int {
	ux := uint64(x) << 1
	if x < 0 {
		ux = ^ux
	}
	return UvarintSize(ux)
}
