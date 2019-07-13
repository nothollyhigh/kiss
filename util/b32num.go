package util

var (
	b32RandLetters   = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"
	b32RandLetterMap = map[rune]int64{}
)

func init() {
	for i, v := range b32RandLetters {
		b32RandLetterMap[v] = int64(i)
	}
}

// func Num2InviteCode(n int64) string {
// 	s := ""
// 	for i := 0; n > 0 || i < 6; i++ {
// 		s = string(b32RandLetters[n%32]) + s
// 		n = n / 32
// 	}
// 	return s
// }

func B32ToString(n int64) string {
	s := ""
	for i := 0; n > 0; i++ {
		s = string(b32RandLetters[n%32]) + s
		n = n / 32
	}
	return s
}

func B32ToNum(s string) int64 {
	n := int64(0)
	for _, v := range s {
		if i, ok := b32RandLetterMap[v]; ok {
			n = n*32 + i
		} else {
			return -1
		}
	}
	return n
}
