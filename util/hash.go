package util

func Hash(str string) uint64 {
	var h = uint64(0)

	for _, c := range str {
		h = 31*h + uint64(c)
	}
	return h
}
